package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	TeamRoleOwner  = "owner"
	TeamRoleAdmin  = "admin"
	TeamRoleMember = "member"

	TeamJoinRequestPending  = "pending"
	TeamJoinRequestApproved = "approved"
	TeamJoinRequestRejected = "rejected"

	TeamReservationReserved = "reserved"
	TeamReservationSettled  = "settled"
	TeamReservationRefunded = "refunded"
)

var (
	ErrTeamAccessDenied       = errors.New("team access denied")
	ErrTeamMembershipInvalid  = errors.New("team membership is invalid")
	ErrTeamQuotaInsufficient  = errors.New("team quota is insufficient")
	ErrTeamMonthlyLimit       = errors.New("team member monthly limit exceeded")
	ErrTeamInviteInvalid      = errors.New("team invite is invalid or expired")
	ErrTeamJoinAlreadyPending = errors.New("team join request is already pending")
)

type Team struct {
	Id          int    `json:"id"`
	Name        string `json:"name" gorm:"type:varchar(80)"`
	JoinCode    string `json:"join_code,omitempty" gorm:"type:varchar(32);uniqueIndex"`
	Quota       int    `json:"balance_quota"`
	CreatedBy   int    `json:"created_by" gorm:"index"`
	CreatedTime int64  `json:"created_time" gorm:"bigint"`
	UpdatedTime int64  `json:"updated_time" gorm:"bigint"`
}

type TeamMember struct {
	Id                int    `json:"id"`
	TeamId            int    `json:"team_id" gorm:"uniqueIndex:idx_team_member,priority:1;index"`
	UserId            int    `json:"user_id" gorm:"uniqueIndex:idx_team_member,priority:2;index"`
	Role              string `json:"role" gorm:"type:varchar(16)"`
	Department        string `json:"department" gorm:"type:varchar(80)"`
	MonthlyLimitQuota int    `json:"monthly_limit_quota"`
	CreatedTime       int64  `json:"created_time" gorm:"bigint"`
}

type TeamMonthlyUsage struct {
	Id      int    `json:"id"`
	TeamId  int    `json:"team_id" gorm:"uniqueIndex:idx_team_month_usage,priority:1;index"`
	UserId  int    `json:"user_id" gorm:"uniqueIndex:idx_team_month_usage,priority:2;index"`
	Month   string `json:"month" gorm:"type:varchar(7);uniqueIndex:idx_team_month_usage,priority:3"`
	Quota   int    `json:"usage_quota"`
	Updated int64  `json:"updated_time" gorm:"bigint"`
}

type TeamJoinRequest struct {
	Id          int    `json:"id"`
	TeamId      int    `json:"team_id" gorm:"uniqueIndex:idx_team_join_request,priority:1;index"`
	UserId      int    `json:"user_id" gorm:"uniqueIndex:idx_team_join_request,priority:2;index"`
	Status      string `json:"status" gorm:"type:varchar(16);index"`
	CreatedTime int64  `json:"created_time" gorm:"bigint"`
	UpdatedTime int64  `json:"updated_time" gorm:"bigint"`
}

type TeamInvite struct {
	Id          int    `json:"id"`
	TeamId      int    `json:"team_id" gorm:"index"`
	TokenHash   string `json:"-" gorm:"type:varchar(128);uniqueIndex"`
	ExpiresAt   int64  `json:"expires_at" gorm:"bigint;index"`
	MaxUses     int    `json:"max_uses"`
	UsedCount   int    `json:"used_count"`
	RevokedAt   int64  `json:"revoked_at" gorm:"bigint"`
	CreatedBy   int    `json:"created_by"`
	CreatedTime int64  `json:"created_time" gorm:"bigint"`
}

type TeamQuotaReservation struct {
	Id          int    `json:"id"`
	RequestId   string `json:"request_id" gorm:"type:varchar(64);uniqueIndex"`
	TeamId      int    `json:"team_id" gorm:"index"`
	UserId      int    `json:"user_id" gorm:"index"`
	Month       string `json:"month" gorm:"type:varchar(7);index"`
	Quota       int    `json:"quota"`
	Status      string `json:"status" gorm:"type:varchar(16);index"`
	CreatedTime int64  `json:"created_time" gorm:"bigint"`
	UpdatedTime int64  `json:"updated_time" gorm:"bigint"`
}

type TeamQuotaTransfer struct {
	Id          int    `json:"id"`
	RequestId   string `json:"request_id" gorm:"type:varchar(64);uniqueIndex"`
	TeamId      int    `json:"team_id" gorm:"index"`
	UserId      int    `json:"user_id" gorm:"index"`
	Quota       int    `json:"quota"`
	CreatedTime int64  `json:"created_time" gorm:"bigint"`
}

type TeamMembershipSummary struct {
	Id                int    `json:"id"`
	Name              string `json:"name"`
	Role              string `json:"role"`
	Department        string `json:"department"`
	BalanceQuota      int    `json:"balance_quota"`
	MonthlyLimitQuota int    `json:"monthly_limit_quota"`
	MonthlyUsedQuota  int    `json:"monthly_used_quota"`
	Month             string `json:"month"`
}

type TeamMemberView struct {
	UserId            int    `json:"user_id"`
	Username          string `json:"username"`
	Role              string `json:"role"`
	Department        string `json:"department"`
	MonthlyLimitQuota int    `json:"monthly_limit_quota"`
	MonthlyUsedQuota  int    `json:"monthly_used_quota"`
	Month             string `json:"month"`
	CreatedTime       int64  `json:"created_time"`
}

func CurrentTeamMonth() string {
	return time.Now().In(time.FixedZone("Asia/Shanghai", 8*60*60)).Format("2006-01")
}

func teamInviteHash(token string) string {
	return common.GenerateHMACWithKey([]byte("team-invite-v1:"+common.SessionSecret), token)
}

func CreateTeam(userID int, name string) (*Team, error) {
	name = strings.TrimSpace(name)
	if userID <= 0 || name == "" || len([]rune(name)) > 80 {
		return nil, errors.New("invalid team name")
	}
	joinCode, err := common.GenerateRandomCharsKey(12)
	if err != nil {
		return nil, err
	}
	now := common.GetTimestamp()
	team := &Team{Name: name, JoinCode: strings.ToUpper(joinCode), CreatedBy: userID, CreatedTime: now, UpdatedTime: now}
	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(team).Error; err != nil {
			return err
		}
		return tx.Create(&TeamMember{TeamId: team.Id, UserId: userID, Role: TeamRoleOwner, MonthlyLimitQuota: 0, CreatedTime: now}).Error
	})
	return team, err
}

func GetTeamMembership(teamID, userID int) (*TeamMember, error) {
	var member TeamMember
	err := DB.Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error
	return &member, err
}

func GetTeamMemberships(userID int) ([]TeamMembershipSummary, error) {
	month := CurrentTeamMonth()
	var rows []TeamMembershipSummary
	err := DB.Table("team_members AS tm").
		Select("t.id, t.name, tm.role, tm.department, t.quota AS balance_quota, tm.monthly_limit_quota, COALESCE(u.quota, 0) AS monthly_used_quota, ? AS month", month).
		Joins("JOIN teams AS t ON t.id = tm.team_id").
		Joins("LEFT JOIN team_monthly_usages AS u ON u.team_id = tm.team_id AND u.user_id = tm.user_id AND u.month = ?", month).
		Where("tm.user_id = ?", userID).Order("t.id").Scan(&rows).Error
	return rows, err
}

func GetTeamForUser(teamID, userID int) (*Team, *TeamMember, error) {
	member, err := GetTeamMembership(teamID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrTeamAccessDenied
		}
		return nil, nil, err
	}
	var team Team
	if err := DB.First(&team, teamID).Error; err != nil {
		return nil, nil, err
	}
	return &team, member, nil
}

func ListTeamMembers(teamID int) ([]TeamMemberView, error) {
	month := CurrentTeamMonth()
	var rows []TeamMemberView
	err := DB.Table("team_members AS tm").
		Select("tm.user_id, users.username, tm.role, tm.department, tm.monthly_limit_quota, COALESCE(u.quota, 0) AS monthly_used_quota, ? AS month, tm.created_time", month).
		Joins("JOIN users ON users.id = tm.user_id").
		Joins("LEFT JOIN team_monthly_usages AS u ON u.team_id = tm.team_id AND u.user_id = tm.user_id AND u.month = ?", month).
		Where("tm.team_id = ?", teamID).Order("tm.id").Scan(&rows).Error
	return rows, err
}

func CreateTeamJoinRequest(userID int, code string) (*TeamJoinRequest, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	var team Team
	if err := DB.Where("join_code = ?", code).First(&team).Error; err != nil {
		return nil, ErrTeamAccessDenied
	}
	if _, err := GetTeamMembership(team.Id, userID); err == nil {
		return nil, errors.New("already a team member")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	now := common.GetTimestamp()
	request := &TeamJoinRequest{TeamId: team.Id, UserId: userID, Status: TeamJoinRequestPending, CreatedTime: now, UpdatedTime: now}
	var existing TeamJoinRequest
	err := DB.Where("team_id = ? AND user_id = ?", team.Id, userID).First(&existing).Error
	if err == nil {
		if existing.Status == TeamJoinRequestPending {
			return nil, ErrTeamJoinAlreadyPending
		}
		existing.Status = TeamJoinRequestPending
		existing.UpdatedTime = now
		return &existing, DB.Model(&existing).Select("status", "updated_time").Updates(&existing).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return request, DB.Create(request).Error
}

func ListTeamJoinRequests(teamID int) ([]TeamJoinRequest, error) {
	var requests []TeamJoinRequest
	err := DB.Where("team_id = ? AND status = ?", teamID, TeamJoinRequestPending).Order("id").Find(&requests).Error
	return requests, err
}

func DecideTeamJoinRequest(teamID, requestID int, approve bool) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var request TeamJoinRequest
		if err := lockForUpdate(tx).Where("id = ? AND team_id = ?", requestID, teamID).First(&request).Error; err != nil {
			return err
		}
		if request.Status != TeamJoinRequestPending {
			return errors.New("join request is no longer pending")
		}
		status := TeamJoinRequestRejected
		if approve {
			status = TeamJoinRequestApproved
			member := TeamMember{TeamId: teamID, UserId: request.UserId, Role: TeamRoleMember, CreatedTime: common.GetTimestamp()}
			if err := tx.Where("team_id = ? AND user_id = ?", teamID, request.UserId).FirstOrCreate(&member).Error; err != nil {
				return err
			}
		}
		return tx.Model(&request).Updates(map[string]any{"status": status, "updated_time": common.GetTimestamp()}).Error
	})
}

func CreateTeamInvite(teamID, createdBy int, expiresInSeconds int64, maxUses int) (*TeamInvite, string, error) {
	if expiresInSeconds < 60 || expiresInSeconds > 30*24*60*60 || maxUses < 1 || maxUses > 10_000 {
		return nil, "", errors.New("invalid invite limits")
	}
	raw, err := common.GenerateKey()
	if err != nil {
		return nil, "", err
	}
	now := common.GetTimestamp()
	invite := &TeamInvite{TeamId: teamID, TokenHash: teamInviteHash(raw), ExpiresAt: now + expiresInSeconds, MaxUses: maxUses, CreatedBy: createdBy, CreatedTime: now}
	return invite, raw, DB.Create(invite).Error
}

func ListTeamInvites(teamID int) ([]TeamInvite, error) {
	var invites []TeamInvite
	err := DB.Where("team_id = ?", teamID).Order("id desc").Find(&invites).Error
	return invites, err
}

func RevokeTeamInvite(teamID, inviteID int) error {
	result := DB.Model(&TeamInvite{}).Where("id = ? AND team_id = ? AND revoked_at = 0", inviteID, teamID).Update("revoked_at", common.GetTimestamp())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func JoinTeamByInvite(userID int, raw string) (*TeamMember, error) {
	hash := teamInviteHash(strings.TrimSpace(raw))
	now := common.GetTimestamp()
	var member TeamMember
	err := DB.Transaction(func(tx *gorm.DB) error {
		var invite TeamInvite
		if err := lockForUpdate(tx).Where("token_hash = ?", hash).First(&invite).Error; err != nil {
			return ErrTeamInviteInvalid
		}
		if invite.RevokedAt != 0 || invite.ExpiresAt <= now || invite.UsedCount >= invite.MaxUses {
			return ErrTeamInviteInvalid
		}
		member = TeamMember{TeamId: invite.TeamId, UserId: userID, Role: TeamRoleMember, CreatedTime: now}
		var existing TeamMember
		if err := tx.Where("team_id = ? AND user_id = ?", invite.TeamId, userID).First(&existing).Error; err == nil {
			member = existing
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		result := tx.Model(&TeamInvite{}).Where("id = ? AND revoked_at = 0 AND expires_at > ? AND used_count < max_uses", invite.Id, now).Update("used_count", gorm.Expr("used_count + 1"))
		if result.Error != nil || result.RowsAffected != 1 {
			return ErrTeamInviteInvalid
		}
		return tx.Create(&member).Error
	})
	return &member, err
}

func UpdateTeamMember(teamID, userID int, role, department *string, monthlyLimit *int) error {
	member, err := GetTeamMembership(teamID, userID)
	if err != nil {
		return err
	}
	updates := map[string]any{}
	if role != nil {
		if member.Role == TeamRoleOwner {
			return ErrTeamAccessDenied
		}
		if *role != TeamRoleAdmin && *role != TeamRoleMember {
			return errors.New("invalid team role")
		}
		updates["role"] = *role
	}
	if department != nil {
		value := strings.TrimSpace(*department)
		if len([]rune(value)) > 80 {
			return errors.New("department is too long")
		}
		updates["department"] = value
	}
	if monthlyLimit != nil {
		if *monthlyLimit < 0 {
			return errors.New("monthly limit cannot be negative")
		}
		updates["monthly_limit_quota"] = *monthlyLimit
	}
	if len(updates) == 0 {
		return nil
	}
	result := DB.Model(&TeamMember{}).Where("team_id = ? AND user_id = ?", teamID, userID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrTeamAccessDenied
	}
	return nil
}

func RemoveTeamMember(teamID, userID int) error {
	result := DB.Where("team_id = ? AND user_id = ? AND role <> ?", teamID, userID, TeamRoleOwner).Delete(&TeamMember{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrTeamAccessDenied
	}
	return nil
}

func TransferUserQuotaToTeam(requestID string, userID, teamID, quota int) error {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return errors.New("request id is required")
	}
	if quota <= 0 {
		return errors.New("quota must be positive")
	}
	if err := common.ValidateWalletQuota(quota); err != nil {
		return err
	}
	var existing TeamQuotaTransfer
	if err := DB.Where("request_id = ?", requestID).First(&existing).Error; err == nil {
		if existing.UserId == userID && existing.TeamId == teamID && existing.Quota == quota {
			return nil
		}
		return errors.New("team transfer request conflict")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	cacheReserved := false
	if common.RedisEnabled {
		result, err := cacheTryReserveUserQuota(userID, int64(quota))
		if err == nil && result == cacheQuotaMiss {
			if _, hydrateErr := GetUserCache(userID); hydrateErr != nil {
				return hydrateErr
			}
			result, err = cacheTryReserveUserQuota(userID, int64(quota))
		}
		if err != nil {
			return fmt.Errorf("user quota cache unavailable: %w", err)
		}
		if result != cacheQuotaOK {
			return ErrTeamQuotaInsufficient
		}
		cacheReserved = true
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		var team Team
		if err := lockForUpdate(tx).First(&team, teamID).Error; err != nil {
			return err
		}
		if team.Quota > common.MaxWalletQuota-quota {
			return ErrWalletQuotaLimitExceeded
		}
		var user User
		if err := lockForUpdate(tx).First(&user, userID).Error; err != nil {
			return err
		}
		if user.Quota < quota {
			return ErrTeamQuotaInsufficient
		}
		if err := tx.Model(&User{}).Where("id = ?", userID).Update("quota", gorm.Expr("quota - ?", quota)).Error; err != nil {
			return err
		}
		if err := tx.Model(&Team{}).Where("id = ?", teamID).Updates(map[string]any{"quota": gorm.Expr("quota + ?", quota), "updated_time": common.GetTimestamp()}).Error; err != nil {
			return err
		}
		return tx.Create(&TeamQuotaTransfer{RequestId: requestID, TeamId: teamID, UserId: userID, Quota: quota, CreatedTime: common.GetTimestamp()}).Error
	})
	if err != nil && cacheReserved {
		compensated, compensateErr := cacheApplyUserQuotaDelta(userID, int64(quota))
		if compensateErr != nil || compensated != cacheQuotaOK {
			common.SysError(fmt.Sprintf("failed to compensate team transfer cache reserve: result=%d error=%v", compensated, compensateErr))
		}
	}
	return err
}

func validateTeamFundingTx(tx *gorm.DB, teamID, userID int) (*Team, *TeamMember, *TeamMonthlyUsage, error) {
	var team Team
	if err := lockForUpdate(tx).First(&team, teamID).Error; err != nil {
		return nil, nil, nil, err
	}
	var member TeamMember
	if err := lockForUpdate(tx).Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, ErrTeamMembershipInvalid
		}
		return nil, nil, nil, err
	}
	month := CurrentTeamMonth()
	usage := TeamMonthlyUsage{TeamId: teamID, UserId: userID, Month: month}
	if err := tx.Where("team_id = ? AND user_id = ? AND month = ?", teamID, userID, month).FirstOrCreate(&usage).Error; err != nil {
		return nil, nil, nil, err
	}
	if err := lockForUpdate(tx).Where("id = ?", usage.Id).First(&usage).Error; err != nil {
		return nil, nil, nil, err
	}
	return &team, &member, &usage, nil
}

func ReserveTeamQuota(requestID string, teamID, userID, targetQuota int) error {
	if requestID == "" || teamID <= 0 || userID <= 0 || targetQuota < 0 {
		return errors.New("invalid team quota reservation")
	}
	if targetQuota == 0 {
		return nil
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var existing TeamQuotaReservation
		err := lockForUpdate(tx).Where("request_id = ?", requestID).First(&existing).Error
		if err == nil {
			if existing.TeamId != teamID || existing.UserId != userID || existing.Status != TeamReservationReserved {
				return errors.New("team reservation request conflict")
			}
			if existing.Quota >= targetQuota {
				return nil
			}
			return reserveTeamQuotaDeltaTx(tx, &existing, targetQuota-existing.Quota)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		team, member, usage, err := validateTeamFundingTx(tx, teamID, userID)
		if err != nil {
			return err
		}
		if team.Quota < targetQuota {
			return ErrTeamQuotaInsufficient
		}
		if member.MonthlyLimitQuota <= 0 || usage.Quota > member.MonthlyLimitQuota-targetQuota {
			return ErrTeamMonthlyLimit
		}
		now := common.GetTimestamp()
		reservation := TeamQuotaReservation{RequestId: requestID, TeamId: teamID, UserId: userID, Month: usage.Month, Quota: targetQuota, Status: TeamReservationReserved, CreatedTime: now, UpdatedTime: now}
		if err := tx.Create(&reservation).Error; err != nil {
			return err
		}
		if err := tx.Model(team).Updates(map[string]any{"quota": gorm.Expr("quota - ?", targetQuota), "updated_time": now}).Error; err != nil {
			return err
		}
		return tx.Model(usage).Updates(map[string]any{"quota": gorm.Expr("quota + ?", targetQuota), "updated": now}).Error
	})
}

func reserveTeamQuotaDeltaTx(tx *gorm.DB, reservation *TeamQuotaReservation, delta int) error {
	if delta <= 0 {
		return nil
	}
	var team Team
	if err := lockForUpdate(tx).First(&team, reservation.TeamId).Error; err != nil {
		return err
	}
	var member TeamMember
	if err := lockForUpdate(tx).Where("team_id = ? AND user_id = ?", reservation.TeamId, reservation.UserId).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTeamMembershipInvalid
		}
		return err
	}
	var usage TeamMonthlyUsage
	if err := lockForUpdate(tx).Where("team_id = ? AND user_id = ? AND month = ?", reservation.TeamId, reservation.UserId, reservation.Month).First(&usage).Error; err != nil {
		return err
	}
	if team.Quota < delta {
		return ErrTeamQuotaInsufficient
	}
	if member.MonthlyLimitQuota <= 0 || usage.Quota > member.MonthlyLimitQuota-delta {
		return ErrTeamMonthlyLimit
	}
	now := common.GetTimestamp()
	if err := tx.Model(&team).Updates(map[string]any{"quota": gorm.Expr("quota - ?", delta), "updated_time": now}).Error; err != nil {
		return err
	}
	if err := tx.Model(&usage).Updates(map[string]any{"quota": gorm.Expr("quota + ?", delta), "updated": now}).Error; err != nil {
		return err
	}
	return tx.Model(reservation).Updates(map[string]any{"quota": gorm.Expr("quota + ?", delta), "updated_time": now}).Error
}

func SettleTeamQuota(requestID string, delta int) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var reservation TeamQuotaReservation
		if err := lockForUpdate(tx).Where("request_id = ?", requestID).First(&reservation).Error; err != nil {
			return err
		}
		if reservation.Status == TeamReservationSettled {
			return nil
		}
		if reservation.Status != TeamReservationReserved {
			return errors.New("team reservation cannot be settled")
		}
		if delta > 0 {
			if err := settleTeamQuotaDeltaTx(tx, &reservation, delta); err != nil {
				return err
			}
		} else if delta < 0 {
			refund := min(-delta, reservation.Quota)
			if err := applyTeamReservationRefundTx(tx, &reservation, refund); err != nil {
				return err
			}
		}
		return tx.Model(&TeamQuotaReservation{}).Where("id = ?", reservation.Id).Updates(map[string]any{"status": TeamReservationSettled, "updated_time": common.GetTimestamp()}).Error
	})
}

// ResizeReservedTeamQuota aligns an admitted asynchronous task reservation
// with the durable submit-time quota without closing it. The later terminal
// transition remains responsible for settling or refunding the reservation.
func ResizeReservedTeamQuota(requestID string, targetQuota int) error {
	if requestID == "" || targetQuota < 0 {
		return errors.New("invalid team reservation resize")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var reservation TeamQuotaReservation
		if err := lockForUpdate(tx).Where("request_id = ?", requestID).First(&reservation).Error; err != nil {
			return err
		}
		if reservation.Status != TeamReservationReserved {
			return errors.New("team reservation cannot be resized")
		}
		delta := targetQuota - reservation.Quota
		if delta > 0 {
			return errors.New("team reservation increases must be reserved before submission")
		}
		if delta < 0 {
			return applyTeamReservationRefundTx(tx, &reservation, -delta)
		}
		return nil
	})
}

// settleTeamQuotaDeltaTx completes an already-admitted request. Membership
// changes after admission must not redirect its payer or prevent settlement.
func settleTeamQuotaDeltaTx(tx *gorm.DB, reservation *TeamQuotaReservation, delta int) error {
	var team Team
	if err := lockForUpdate(tx).First(&team, reservation.TeamId).Error; err != nil {
		return err
	}
	var usage TeamMonthlyUsage
	if err := lockForUpdate(tx).Where("team_id = ? AND user_id = ? AND month = ?", reservation.TeamId, reservation.UserId, reservation.Month).First(&usage).Error; err != nil {
		return err
	}
	now := common.GetTimestamp()
	if err := tx.Model(&team).Updates(map[string]any{"quota": gorm.Expr("quota - ?", delta), "updated_time": now}).Error; err != nil {
		return err
	}
	if err := tx.Model(&usage).Updates(map[string]any{"quota": gorm.Expr("quota + ?", delta), "updated": now}).Error; err != nil {
		return err
	}
	return tx.Model(reservation).Updates(map[string]any{"quota": gorm.Expr("quota + ?", delta), "updated_time": now}).Error
}

func applyTeamReservationRefundTx(tx *gorm.DB, reservation *TeamQuotaReservation, quota int) error {
	if quota <= 0 {
		return nil
	}
	now := common.GetTimestamp()
	if err := tx.Model(&Team{}).Where("id = ?", reservation.TeamId).Updates(map[string]any{"quota": gorm.Expr("quota + ?", quota), "updated_time": now}).Error; err != nil {
		return err
	}
	if err := tx.Model(&TeamMonthlyUsage{}).Where("team_id = ? AND user_id = ? AND month = ?", reservation.TeamId, reservation.UserId, reservation.Month).Updates(map[string]any{"quota": gorm.Expr("CASE WHEN quota >= ? THEN quota - ? ELSE 0 END", quota, quota), "updated": now}).Error; err != nil {
		return err
	}
	reservation.Quota -= quota
	return tx.Model(&TeamQuotaReservation{}).Where("id = ?", reservation.Id).Updates(map[string]any{"quota": reservation.Quota, "updated_time": now}).Error
}

func RefundTeamQuota(requestID string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var reservation TeamQuotaReservation
		if err := lockForUpdate(tx).Where("request_id = ?", requestID).First(&reservation).Error; err != nil {
			return err
		}
		if reservation.Status == TeamReservationRefunded {
			return nil
		}
		if reservation.Status != TeamReservationReserved {
			return errors.New("team reservation cannot be refunded")
		}
		if err := applyTeamReservationRefundTx(tx, &reservation, reservation.Quota); err != nil {
			return err
		}
		return tx.Model(&TeamQuotaReservation{}).Where("id = ?", reservation.Id).Updates(map[string]any{"status": TeamReservationRefunded, "updated_time": common.GetTimestamp()}).Error
	})
}

func RollbackTeamQuotaDelta(requestID string, quota int) error {
	if quota <= 0 {
		return nil
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var reservation TeamQuotaReservation
		if err := lockForUpdate(tx).Where("request_id = ?", requestID).First(&reservation).Error; err != nil {
			return err
		}
		if reservation.Status != TeamReservationReserved || reservation.Quota < quota {
			return errors.New("team reservation rollback conflict")
		}
		return applyTeamReservationRefundTx(tx, &reservation, quota)
	})
}

func ListTeamUsage(teamID int, month string) ([]TeamMemberView, int, error) {
	if month == "" {
		month = CurrentTeamMonth()
	}
	if len(month) != 7 || month[4] != '-' {
		return nil, 0, errors.New("invalid month")
	}
	var rows []TeamMemberView
	err := DB.Table("team_monthly_usages AS u").
		Select("u.user_id, users.username, COALESCE(tm.role, '') AS role, COALESCE(tm.department, '') AS department, COALESCE(tm.monthly_limit_quota, 0) AS monthly_limit_quota, u.quota AS monthly_used_quota, u.month, COALESCE(tm.created_time, 0) AS created_time").
		Joins("JOIN users ON users.id = u.user_id").
		Joins("LEFT JOIN team_members AS tm ON tm.team_id = u.team_id AND tm.user_id = u.user_id").
		Where("u.team_id = ? AND u.month = ?", teamID, month).Order("u.quota desc").Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	total := 0
	for _, row := range rows {
		total += row.MonthlyUsedQuota
	}
	return rows, total, nil
}
