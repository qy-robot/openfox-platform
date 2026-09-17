package service_test

import (
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTeamBillingTest(t *testing.T) {
	t.Helper()
	previousDB := model.DB
	previousLogDB := model.LOG_DB
	previousRedis := common.RedisEnabled
	previousBatch := common.BatchUpdateEnabled
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.User{}, &model.Team{}, &model.TeamMember{}, &model.TeamMonthlyUsage{},
		&model.TeamJoinRequest{}, &model.TeamInvite{}, &model.TeamQuotaReservation{},
		&model.TeamQuotaTransfer{}, &model.Task{}, &model.Channel{}, &model.Log{},
	))
	model.DB = db
	model.LOG_DB = db
	common.RedisEnabled = false
	common.BatchUpdateEnabled = false
	t.Cleanup(func() {
		model.DB = previousDB
		model.LOG_DB = previousLogDB
		common.RedisEnabled = previousRedis
		common.BatchUpdateEnabled = previousBatch
	})
}

func seedTeamUser(t *testing.T, username string, quota int) *model.User {
	t.Helper()
	user := &model.User{Username: username, AffCode: username, Quota: quota, Status: common.UserStatusEnabled, Role: common.RoleCommonUser}
	require.NoError(t, model.DB.Create(user).Error)
	return user
}

func TestTeamQuotaLifecycleAndInviteSecurity(t *testing.T) {
	setupTeamBillingTest(t)
	owner := seedTeamUser(t, "team-owner", 300)
	member := seedTeamUser(t, "team-member", 0)
	outsider := seedTeamUser(t, "team-outsider", 0)

	team, err := model.CreateTeam(owner.Id, "Robocoding")
	require.NoError(t, err)
	require.NoError(t, model.TransferUserQuotaToTeam("fund-team", owner.Id, team.Id, 200))
	require.NoError(t, model.TransferUserQuotaToTeam("fund-team", owner.Id, team.Id, 200), "transfer request is idempotent")
	ownerQuota, err := model.GetUserQuota(owner.Id, true)
	require.NoError(t, err)
	assert.Equal(t, 100, ownerQuota)
	require.Error(t, model.TransferUserQuotaToTeam("missing-team", owner.Id, team.Id+999, 10))
	ownerQuota, err = model.GetUserQuota(owner.Id, true)
	require.NoError(t, err)
	assert.Equal(t, 100, ownerQuota, "failed transfer preserves personal wallet")
	require.NoError(t, model.DB.Model(&model.Team{}).Where("id = ?", team.Id).Update("quota", common.MaxWalletQuota).Error)
	require.Error(t, model.TransferUserQuotaToTeam("full-team", owner.Id, team.Id, 1))
	ownerQuota, err = model.GetUserQuota(owner.Id, true)
	require.NoError(t, err)
	assert.Equal(t, 100, ownerQuota, "overflowing transfer preserves personal wallet")
	require.NoError(t, model.DB.Model(&model.Team{}).Where("id = ?", team.Id).Update("quota", 200).Error)

	invite, rawToken, err := model.CreateTeamInvite(team.Id, owner.Id, 3600, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, rawToken)
	assert.NotEqual(t, rawToken, invite.TokenHash)
	_, err = model.JoinTeamByInvite(member.Id, rawToken)
	require.NoError(t, err)
	_, err = model.JoinTeamByInvite(outsider.Id, rawToken)
	require.ErrorIs(t, err, model.ErrTeamInviteInvalid)

	limit := 100
	require.NoError(t, model.UpdateTeamMember(team.Id, member.Id, nil, nil, &limit))
	require.NoError(t, model.ReserveTeamQuota("team-request-1", team.Id, member.Id, 80))
	require.NoError(t, model.ReserveTeamQuota("team-request-1", team.Id, member.Id, 80), "same target is idempotent")
	require.ErrorIs(t, model.ReserveTeamQuota("team-request-2", team.Id, member.Id, 30), model.ErrTeamMonthlyLimit)

	require.NoError(t, model.RemoveTeamMember(team.Id, member.Id))
	require.NoError(t, model.SettleTeamQuota("team-request-1", 150), "admitted work settles after membership removal")
	require.NoError(t, model.SettleTeamQuota("team-request-1", 150), "settlement retry is idempotent")
	var storedTeam model.Team
	require.NoError(t, model.DB.First(&storedTeam, team.Id).Error)
	assert.Equal(t, -30, storedTeam.Quota, "final provider cost may create bounded team debt")
	var usage model.TeamMonthlyUsage
	require.NoError(t, model.DB.Where("team_id = ? AND user_id = ?", team.Id, member.Id).First(&usage).Error)
	assert.Equal(t, 230, usage.Quota)
	require.ErrorIs(t, model.ReserveTeamQuota("team-request-after-removal", team.Id, member.Id, 1), model.ErrTeamMembershipInvalid)

	rows, total, err := model.ListTeamUsage(team.Id, model.CurrentTeamMonth())
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, member.Id, rows[0].UserId, "removed member usage remains visible")
	assert.Equal(t, 230, total)
}

func TestTeamRefundSurvivesMembershipRemovalAndMonthIsolation(t *testing.T) {
	setupTeamBillingTest(t)
	owner := seedTeamUser(t, "refund-owner", 0)
	member := seedTeamUser(t, "refund-member", 0)
	team, err := model.CreateTeam(owner.Id, "Refund Team")
	require.NoError(t, err)
	require.NoError(t, model.DB.Model(team).Update("quota", 100).Error)
	require.NoError(t, model.DB.Create(&model.TeamMember{TeamId: team.Id, UserId: member.Id, Role: model.TeamRoleMember, MonthlyLimitQuota: 100}).Error)
	require.NoError(t, model.ReserveTeamQuota("refund-request", team.Id, member.Id, 60))
	require.NoError(t, model.RemoveTeamMember(team.Id, member.Id))
	require.NoError(t, model.RefundTeamQuota("refund-request"))
	require.NoError(t, model.RefundTeamQuota("refund-request"))

	var storedTeam model.Team
	require.NoError(t, model.DB.First(&storedTeam, team.Id).Error)
	assert.Equal(t, 100, storedTeam.Quota)
	var usage model.TeamMonthlyUsage
	require.NoError(t, model.DB.Where("team_id = ? AND user_id = ?", team.Id, member.Id).First(&usage).Error)
	assert.Zero(t, usage.Quota)

	previousMonth := "2026-08"
	currentMonth := model.CurrentTeamMonth()
	require.NotEqual(t, previousMonth, currentMonth)
	require.NoError(t, model.DB.Create(&model.TeamMonthlyUsage{TeamId: team.Id, UserId: member.Id, Month: previousMonth, Quota: 60}).Error)
	require.NoError(t, model.DB.Model(&model.TeamMonthlyUsage{}).Where("team_id = ? AND user_id = ? AND month = ?", team.Id, member.Id, currentMonth).Update("quota", 25).Error)
	require.NoError(t, model.DB.Model(&model.Team{}).Where("id = ?", team.Id).Update("quota", 40).Error)
	require.NoError(t, model.DB.Create(&model.TeamQuotaReservation{RequestId: "previous-month-request", TeamId: team.Id, UserId: member.Id, Month: previousMonth, Quota: 60, Status: model.TeamReservationReserved}).Error)
	require.NoError(t, model.RefundTeamQuota("previous-month-request"))
	var currentUsage model.TeamMonthlyUsage
	require.NoError(t, model.DB.Where("team_id = ? AND user_id = ? AND month = ?", team.Id, member.Id, currentMonth).First(&currentUsage).Error)
	assert.Equal(t, 25, currentUsage.Quota, "refund uses the reservation month and cannot reduce the new month")
}

func TestAsyncTeamTaskSettlementAcrossSubmissionAndCompletion(t *testing.T) {
	setupTeamBillingTest(t)
	gin.SetMode(gin.TestMode)
	owner := seedTeamUser(t, "async-owner", 0)
	member := seedTeamUser(t, "async-member", 0)
	team, err := model.CreateTeam(owner.Id, "Async Team")
	require.NoError(t, err)
	require.NoError(t, model.DB.Model(team).Update("quota", 100).Error)
	require.NoError(t, model.DB.Create(&model.TeamMember{TeamId: team.Id, UserId: member.Id, Role: model.TeamRoleMember, MonthlyLimitQuota: 200}).Error)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	info := &relaycommon.RelayInfo{
		RequestId: "async-success", UserId: member.Id, TeamId: team.Id,
		FundingMode: model.TokenFundingTeamOnly, TokenUnlimited: true, IsPlayground: true,
		ForcePreConsume: true, UserSetting: dto.UserSetting{BillingPreference: "wallet_only"},
	}
	session, apiErr := service.NewBillingSession(ctx, info, 60)
	require.Nil(t, apiErr)
	info.Billing = session
	require.NoError(t, service.PrepareAsyncTaskBilling(info, 40))

	var reservation model.TeamQuotaReservation
	require.NoError(t, model.DB.Where("request_id = ?", info.RequestId).First(&reservation).Error)
	assert.Equal(t, model.TeamReservationReserved, reservation.Status)
	assert.Equal(t, 40, reservation.Quota)
	task := &model.Task{TaskID: "async-success-task", UserId: member.Id, Quota: 40, PrivateData: model.TaskPrivateData{BillingSource: service.BillingSourceTeam, Execution: &model.TaskExecutionSnapshot{RequestID: info.RequestId}}}
	require.NoError(t, model.DB.Create(task).Error)
	service.RecalculateTaskQuota(t.Context(), task, 70, "async completion")
	require.NoError(t, model.DB.Where("request_id = ?", info.RequestId).First(&reservation).Error)
	assert.Equal(t, model.TeamReservationSettled, reservation.Status)
	assert.Equal(t, 70, reservation.Quota)
	var storedTeam model.Team
	require.NoError(t, model.DB.First(&storedTeam, team.Id).Error)
	assert.Equal(t, 30, storedTeam.Quota)
	var usage model.TeamMonthlyUsage
	require.NoError(t, model.DB.Where("team_id = ? AND user_id = ?", team.Id, member.Id).First(&usage).Error)
	assert.Equal(t, 70, usage.Quota)
	service.RecalculateTaskQuota(t.Context(), task, 70, "idempotent completion retry")
	require.NoError(t, model.DB.First(&storedTeam, team.Id).Error)
	assert.Equal(t, 30, storedTeam.Quota)
}

func TestAsyncTeamTaskFailureRefundsOpenSubmissionReservation(t *testing.T) {
	setupTeamBillingTest(t)
	gin.SetMode(gin.TestMode)
	owner := seedTeamUser(t, "async-refund-owner", 0)
	member := seedTeamUser(t, "async-refund-member", 0)
	team, err := model.CreateTeam(owner.Id, "Async Refund Team")
	require.NoError(t, err)
	require.NoError(t, model.DB.Model(team).Update("quota", 100).Error)
	require.NoError(t, model.DB.Create(&model.TeamMember{TeamId: team.Id, UserId: member.Id, Role: model.TeamRoleMember, MonthlyLimitQuota: 100}).Error)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	info := &relaycommon.RelayInfo{
		RequestId: "async-failure", UserId: member.Id, TeamId: team.Id,
		FundingMode: model.TokenFundingTeamOnly, TokenUnlimited: true, IsPlayground: true,
		ForcePreConsume: true, UserSetting: dto.UserSetting{BillingPreference: "wallet_only"},
	}
	session, apiErr := service.NewBillingSession(ctx, info, 50)
	require.Nil(t, apiErr)
	info.Billing = session
	require.NoError(t, service.PrepareAsyncTaskBilling(info, 30))
	task := &model.Task{TaskID: "async-failure-task", UserId: member.Id, Quota: 30, PrivateData: model.TaskPrivateData{BillingSource: service.BillingSourceTeam, Execution: &model.TaskExecutionSnapshot{RequestID: info.RequestId}}}
	require.NoError(t, model.DB.Create(task).Error)
	require.True(t, service.RefundTaskQuota(t.Context(), task, "provider failed"))
	require.True(t, service.RefundTaskQuota(t.Context(), task, "refund retry"))

	var reservation model.TeamQuotaReservation
	require.NoError(t, model.DB.Where("request_id = ?", info.RequestId).First(&reservation).Error)
	assert.Equal(t, model.TeamReservationRefunded, reservation.Status)
	assert.Zero(t, reservation.Quota)
	var storedTeam model.Team
	require.NoError(t, model.DB.First(&storedTeam, team.Id).Error)
	assert.Equal(t, 100, storedTeam.Quota)
	var usage model.TeamMonthlyUsage
	require.NoError(t, model.DB.Where("team_id = ? AND user_id = ?", team.Id, member.Id).First(&usage).Error)
	assert.Zero(t, usage.Quota)

	zeroInfo := &relaycommon.RelayInfo{
		RequestId: "async-zero-failure", UserId: member.Id, TeamId: team.Id,
		FundingMode: model.TokenFundingTeamOnly, TokenUnlimited: true, IsPlayground: true,
		ForcePreConsume: true, UserSetting: dto.UserSetting{BillingPreference: "wallet_only"},
	}
	zeroSession, apiErr := service.NewBillingSession(ctx, zeroInfo, 0)
	require.Nil(t, apiErr)
	zeroInfo.Billing = zeroSession
	require.NoError(t, service.PrepareAsyncTaskBilling(zeroInfo, 0))
	zeroTask := &model.Task{TaskID: "async-zero-failure-task", UserId: member.Id, PrivateData: model.TaskPrivateData{BillingSource: service.BillingSourceTeam, Execution: &model.TaskExecutionSnapshot{RequestID: zeroInfo.RequestId}}}
	require.NoError(t, model.DB.Create(zeroTask).Error)
	require.True(t, service.RefundTaskQuota(t.Context(), zeroTask, "provider failed without charge"))
	reservation = model.TeamQuotaReservation{}
	require.NoError(t, model.DB.Where("request_id = ?", zeroInfo.RequestId).First(&reservation).Error)
	assert.Equal(t, model.TeamReservationRefunded, reservation.Status)
}

func TestConcurrentTeamAdmissionCannotOverspend(t *testing.T) {
	setupTeamBillingTest(t)
	owner := seedTeamUser(t, "concurrent-owner", 0)
	member := seedTeamUser(t, "concurrent-member", 0)
	team, err := model.CreateTeam(owner.Id, "Concurrent Team")
	require.NoError(t, err)
	require.NoError(t, model.DB.Model(team).Update("quota", 100).Error)
	require.NoError(t, model.DB.Create(&model.TeamMember{TeamId: team.Id, UserId: member.Id, Role: model.TeamRoleMember, MonthlyLimitQuota: 100}).Error)

	results := make(chan error, 2)
	var start sync.WaitGroup
	start.Add(1)
	for _, requestID := range []string{"concurrent-1", "concurrent-2"} {
		go func(id string) {
			start.Wait()
			results <- model.ReserveTeamQuota(id, team.Id, member.Id, 80)
		}(requestID)
	}
	start.Done()
	firstErr, secondErr := <-results, <-results
	successes := 0
	if firstErr == nil {
		successes++
	}
	if secondErr == nil {
		successes++
	}
	assert.Equal(t, 1, successes)
	var storedTeam model.Team
	require.NoError(t, model.DB.First(&storedTeam, team.Id).Error)
	assert.Equal(t, 20, storedTeam.Quota)
	var usage model.TeamMonthlyUsage
	require.NoError(t, model.DB.Where("team_id = ? AND user_id = ?", team.Id, member.Id).First(&usage).Error)
	assert.Equal(t, 80, usage.Quota)
}

func TestTeamFundingAdmissionFallbackAndTenantAuthorization(t *testing.T) {
	setupTeamBillingTest(t)
	gin.SetMode(gin.TestMode)
	owner := seedTeamUser(t, "fallback-owner", 0)
	member := seedTeamUser(t, "fallback-member", 100)
	outsider := seedTeamUser(t, "fallback-outsider", 100)
	team, err := model.CreateTeam(owner.Id, "Fallback Team")
	require.NoError(t, err)
	require.NoError(t, model.DB.Model(team).Update("quota", 10).Error)
	require.NoError(t, model.DB.Create(&model.TeamMember{TeamId: team.Id, UserId: member.Id, Role: model.TeamRoleMember, MonthlyLimitQuota: 10}).Error)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	info := &relaycommon.RelayInfo{
		RequestId:       "fallback-request",
		UserId:          member.Id,
		TeamId:          team.Id,
		FundingMode:     model.TokenFundingTeamFirst,
		TokenUnlimited:  true,
		IsPlayground:    true,
		ForcePreConsume: true,
		UserSetting:     dto.UserSetting{BillingPreference: "wallet_only"},
	}
	session, apiErr := service.NewBillingSession(ctx, info, 20)
	require.Nil(t, apiErr)
	require.NotNil(t, session)
	assert.Equal(t, service.BillingSourceWallet, info.BillingSource)
	memberQuota, err := model.GetUserQuota(member.Id, true)
	require.NoError(t, err)
	assert.Equal(t, 80, memberQuota)
	require.NoError(t, session.Settle(20))

	invalidInfo := &relaycommon.RelayInfo{
		RequestId: "invalid-membership", UserId: outsider.Id, TeamId: team.Id,
		FundingMode: model.TokenFundingTeamFirst, TokenUnlimited: true, IsPlayground: true,
		ForcePreConsume: true, UserSetting: dto.UserSetting{BillingPreference: "wallet_only"},
	}
	_, apiErr = service.NewBillingSession(ctx, invalidInfo, 20)
	require.NotNil(t, apiErr)
	assert.Equal(t, types.ErrorCodeAccessDenied, apiErr.GetErrorCode())
	outsiderQuota, err := model.GetUserQuota(outsider.Id, true)
	require.NoError(t, err)
	assert.Equal(t, 100, outsiderQuota, "invalid membership never falls back to personal funds")

	recorder := httptest.NewRecorder()
	controllerCtx, _ := gin.CreateTestContext(recorder)
	controllerCtx.Set("id", outsider.Id)
	controllerCtx.Params = gin.Params{{Key: "id", Value: strconv.Itoa(team.Id)}}
	controller.GetTeam(controllerCtx)
	assert.Contains(t, recorder.Body.String(), `"success":false`)
	assert.NotContains(t, recorder.Body.String(), team.JoinCode)
}

func TestPersonalFirstFallsBackToTeamAndTeamOnlyNeverChargesPersonal(t *testing.T) {
	setupTeamBillingTest(t)
	gin.SetMode(gin.TestMode)
	owner := seedTeamUser(t, "policy-owner", 0)
	member := seedTeamUser(t, "policy-member", 0)
	personalRich := seedTeamUser(t, "team-only-member", 100)
	team, err := model.CreateTeam(owner.Id, "Policy Team")
	require.NoError(t, err)
	require.NoError(t, model.DB.Model(team).Update("quota", 50).Error)
	require.NoError(t, model.DB.Create(&model.TeamMember{TeamId: team.Id, UserId: member.Id, Role: model.TeamRoleMember, MonthlyLimitQuota: 50}).Error)
	require.NoError(t, model.DB.Create(&model.TeamMember{TeamId: team.Id, UserId: personalRich.Id, Role: model.TeamRoleMember, MonthlyLimitQuota: 50}).Error)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	personalFirst := &relaycommon.RelayInfo{
		RequestId: "personal-first", UserId: member.Id, TeamId: team.Id,
		FundingMode: model.TokenFundingPersonalFirst, TokenUnlimited: true, IsPlayground: true,
		ForcePreConsume: true, UserSetting: dto.UserSetting{BillingPreference: "wallet_only"},
	}
	session, apiErr := service.NewBillingSession(ctx, personalFirst, 20)
	require.Nil(t, apiErr)
	require.NotNil(t, session)
	assert.Equal(t, service.BillingSourceTeam, personalFirst.BillingSource)
	require.NoError(t, session.Settle(20))

	require.NoError(t, model.DB.Model(&model.Team{}).Where("id = ?", team.Id).Update("quota", 0).Error)
	teamOnly := &relaycommon.RelayInfo{
		RequestId: "team-only", UserId: personalRich.Id, TeamId: team.Id,
		FundingMode: model.TokenFundingTeamOnly, TokenUnlimited: true, IsPlayground: true,
		ForcePreConsume: true, UserSetting: dto.UserSetting{BillingPreference: "wallet_only"},
	}
	_, apiErr = service.NewBillingSession(ctx, teamOnly, 20)
	require.NotNil(t, apiErr)
	assert.Equal(t, types.ErrorCodeInsufficientUserQuota, apiErr.GetErrorCode())
	quota, err := model.GetUserQuota(personalRich.Id, true)
	require.NoError(t, err)
	assert.Equal(t, 100, quota)

	require.NoError(t, model.DB.Model(&model.Team{}).Where("id = ?", team.Id).Update("quota", 10).Error)
	zeroEstimate := &relaycommon.RelayInfo{
		RequestId: "zero-estimate", UserId: member.Id, TeamId: team.Id,
		FundingMode: model.TokenFundingTeamOnly, TokenUnlimited: true, IsPlayground: true,
		ForcePreConsume: true, UserSetting: dto.UserSetting{BillingPreference: "wallet_only"},
	}
	zeroSession, apiErr := service.NewBillingSession(ctx, zeroEstimate, 0)
	require.Nil(t, apiErr)
	require.NotNil(t, zeroSession)
	assert.Equal(t, 1, zeroSession.GetPreConsumedQuota())
	require.NoError(t, zeroSession.Settle(0))
	var zeroReservation model.TeamQuotaReservation
	require.NoError(t, model.DB.Where("request_id = ?", "zero-estimate").First(&zeroReservation).Error)
	assert.Equal(t, model.TeamReservationSettled, zeroReservation.Status)
	assert.Zero(t, zeroReservation.Quota)

	require.NoError(t, model.RemoveTeamMember(team.Id, personalRich.Id))
	revokedPersonalFirst := &relaycommon.RelayInfo{
		RequestId: "revoked-personal-first", UserId: personalRich.Id, TeamId: team.Id,
		FundingMode: model.TokenFundingPersonalFirst, TokenUnlimited: true, IsPlayground: true,
		ForcePreConsume: true, UserSetting: dto.UserSetting{BillingPreference: "wallet_only"},
	}
	_, apiErr = service.NewBillingSession(ctx, revokedPersonalFirst, 5)
	require.NotNil(t, apiErr)
	assert.Equal(t, types.ErrorCodeAccessDenied, apiErr.GetErrorCode())
	quota, err = model.GetUserQuota(personalRich.Id, true)
	require.NoError(t, err)
	assert.Equal(t, 100, quota, "revoked personal_first key fails before personal admission")
}

func TestTeamAccountingSupportedDatabases(t *testing.T) {
	tests := []struct {
		name      string
		env       string
		dialector func(string) gorm.Dialector
	}{
		{name: "mysql", env: "TEAM_TEST_MYSQL_DSN", dialector: func(dsn string) gorm.Dialector { return mysql.Open(dsn) }},
		{name: "postgres", env: "TEAM_TEST_POSTGRES_DSN", dialector: func(dsn string) gorm.Dialector { return postgres.Open(dsn) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dsn := os.Getenv(test.env)
			if dsn == "" {
				t.Skip(test.env + " is not configured")
			}
			db, err := gorm.Open(test.dialector(dsn), &gorm.Config{})
			require.NoError(t, err)
			require.NoError(t, db.Migrator().DropTable(
				&model.TeamQuotaTransfer{}, &model.TeamQuotaReservation{}, &model.TeamInvite{},
				&model.TeamJoinRequest{}, &model.TeamMonthlyUsage{}, &model.TeamMember{}, &model.Team{},
			))
			require.NoError(t, db.AutoMigrate(
				&model.Team{}, &model.TeamMember{}, &model.TeamMonthlyUsage{}, &model.TeamJoinRequest{},
				&model.TeamInvite{}, &model.TeamQuotaReservation{}, &model.TeamQuotaTransfer{},
			))
			previousDB := model.DB
			model.DB = db
			t.Cleanup(func() { model.DB = previousDB })
			team := &model.Team{Name: "dialect", JoinCode: "DIALECTCODE", Quota: 100, CreatedBy: 1}
			require.NoError(t, db.Create(team).Error)
			require.NoError(t, db.Create(&model.TeamMember{TeamId: team.Id, UserId: 2, Role: model.TeamRoleMember, MonthlyLimitQuota: 100}).Error)
			require.NoError(t, model.ReserveTeamQuota("dialect-reserve", team.Id, 2, 60))
			require.NoError(t, model.ResizeReservedTeamQuota("dialect-reserve", 40), "submission keeps the adjusted reservation open")
			var reservation model.TeamQuotaReservation
			require.NoError(t, db.Where("request_id = ?", "dialect-reserve").First(&reservation).Error)
			assert.Equal(t, model.TeamReservationReserved, reservation.Status)
			require.NoError(t, model.SettleTeamQuota("dialect-reserve", 10), "completion applies the final delta")
			require.NoError(t, model.SettleTeamQuota("dialect-reserve", 10), "completion retry is idempotent")
			var stored model.Team
			require.NoError(t, db.First(&stored, team.Id).Error)
			assert.Equal(t, 50, stored.Quota)

			require.NoError(t, model.ReserveTeamQuota("dialect-refund", team.Id, 2, 20))
			require.NoError(t, model.ResizeReservedTeamQuota("dialect-refund", 15))
			require.NoError(t, model.RefundTeamQuota("dialect-refund"), "failure refunds the open submission reservation")
			require.NoError(t, model.RefundTeamQuota("dialect-refund"), "failure retry is idempotent")
			require.NoError(t, db.First(&stored, team.Id).Error)
			assert.Equal(t, 50, stored.Quota)
		})
	}
}
