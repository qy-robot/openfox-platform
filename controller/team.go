package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

type teamCreateRequest struct {
	Name string `json:"name"`
}

type teamJoinRequest struct {
	Code string `json:"code"`
}

type teamInviteRequest struct {
	ExpiresInSeconds int64 `json:"expires_in_seconds"`
	MaxUses          int   `json:"max_uses"`
}

type teamMemberUpdateRequest struct {
	Role              *string `json:"role"`
	Department        *string `json:"department"`
	MonthlyLimitQuota *int    `json:"monthly_limit_quota"`
}

type teamFundRequest struct {
	Quota     int    `json:"quota"`
	RequestId string `json:"request_id"`
}

func teamIDFromParam(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "invalid team id")
		return 0, false
	}
	return id, true
}

func requireTeamManager(c *gin.Context, teamID int) (*model.TeamMember, bool) {
	member, err := model.GetTeamMembership(teamID, c.GetInt("id"))
	if err != nil || (member.Role != model.TeamRoleOwner && member.Role != model.TeamRoleAdmin) {
		common.ApiErrorMsg(c, model.ErrTeamAccessDenied.Error())
		return nil, false
	}
	return member, true
}

func ListMyTeams(c *gin.Context) {
	teams, err := model.GetTeamMemberships(c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, teams)
}

func CreateTeam(c *gin.Context) {
	var request teamCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	team, err := model.CreateTeam(c.GetInt("id"), request.Name)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"id": team.Id, "name": team.Name, "join_code": team.JoinCode, "balance_quota": team.Quota, "role": model.TeamRoleOwner, "member_count": 1, "created_time": team.CreatedTime})
}

func GetTeam(c *gin.Context) {
	teamID, ok := teamIDFromParam(c)
	if !ok {
		return
	}
	team, member, err := model.GetTeamForUser(teamID, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	members, err := model.ListTeamMembers(teamID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	joinCode := ""
	if member.Role == model.TeamRoleOwner || member.Role == model.TeamRoleAdmin {
		joinCode = team.JoinCode
	}
	common.ApiSuccess(c, gin.H{"id": team.Id, "name": team.Name, "join_code": joinCode, "balance_quota": team.Quota, "role": member.Role, "member_count": len(members), "created_time": team.CreatedTime})
}

func RequestJoinTeam(c *gin.Context) {
	var request teamJoinRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	joinRequest, err := model.CreateTeamJoinRequest(c.GetInt("id"), request.Code)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, joinRequest)
}

func ListTeamJoinRequests(c *gin.Context) {
	teamID, ok := teamIDFromParam(c)
	if !ok {
		return
	}
	if _, ok := requireTeamManager(c, teamID); !ok {
		return
	}
	requests, err := model.ListTeamJoinRequests(teamID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, requests)
}

func DecideTeamJoinRequest(c *gin.Context) {
	teamID, ok := teamIDFromParam(c)
	if !ok {
		return
	}
	if _, ok := requireTeamManager(c, teamID); !ok {
		return
	}
	requestID, err := strconv.Atoi(c.Param("request_id"))
	if err != nil || requestID <= 0 {
		common.ApiErrorMsg(c, "invalid join request id")
		return
	}
	approve := c.Request.Method == http.MethodPost
	if err := model.DecideTeamJoinRequest(teamID, requestID, approve); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func CreateTeamInvite(c *gin.Context) {
	teamID, ok := teamIDFromParam(c)
	if !ok {
		return
	}
	if _, ok := requireTeamManager(c, teamID); !ok {
		return
	}
	var request teamInviteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	invite, raw, err := model.CreateTeamInvite(teamID, c.GetInt("id"), request.ExpiresInSeconds, request.MaxUses)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"id": invite.Id, "token": raw, "expires_at": invite.ExpiresAt, "max_uses": invite.MaxUses, "used_count": invite.UsedCount, "revoked_at": invite.RevokedAt, "created_time": invite.CreatedTime})
}

func ListTeamInvites(c *gin.Context) {
	teamID, ok := teamIDFromParam(c)
	if !ok {
		return
	}
	if _, ok := requireTeamManager(c, teamID); !ok {
		return
	}
	invites, err := model.ListTeamInvites(teamID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, invites)
}

func RevokeTeamInvite(c *gin.Context) {
	teamID, ok := teamIDFromParam(c)
	if !ok {
		return
	}
	if _, ok := requireTeamManager(c, teamID); !ok {
		return
	}
	inviteID, err := strconv.Atoi(c.Param("invite_id"))
	if err != nil || inviteID <= 0 {
		common.ApiErrorMsg(c, "invalid invite id")
		return
	}
	if err := model.RevokeTeamInvite(teamID, inviteID); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func JoinTeamByInvite(c *gin.Context) {
	var request struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	member, err := model.JoinTeamByInvite(c.GetInt("id"), request.Token)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, member)
}

func ListTeamMembers(c *gin.Context) {
	teamID, ok := teamIDFromParam(c)
	if !ok {
		return
	}
	if _, ok := requireTeamManager(c, teamID); !ok {
		return
	}
	members, err := model.ListTeamMembers(teamID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, members)
}

func UpdateTeamMember(c *gin.Context) {
	teamID, ok := teamIDFromParam(c)
	if !ok {
		return
	}
	manager, ok := requireTeamManager(c, teamID)
	if !ok {
		return
	}
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil || userID <= 0 {
		common.ApiErrorMsg(c, "invalid user id")
		return
	}
	var request teamMemberUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	if request.Role != nil && manager.Role != model.TeamRoleOwner {
		common.ApiErrorMsg(c, model.ErrTeamAccessDenied.Error())
		return
	}
	target, err := model.GetTeamMembership(teamID, userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if target.Role == model.TeamRoleOwner && manager.Role != model.TeamRoleOwner {
		common.ApiErrorMsg(c, model.ErrTeamAccessDenied.Error())
		return
	}
	if err := model.UpdateTeamMember(teamID, userID, request.Role, request.Department, request.MonthlyLimitQuota); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func RemoveTeamMember(c *gin.Context) {
	teamID, ok := teamIDFromParam(c)
	if !ok {
		return
	}
	manager, ok := requireTeamManager(c, teamID)
	if !ok {
		return
	}
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil || userID <= 0 {
		common.ApiErrorMsg(c, "invalid user id")
		return
	}
	target, err := model.GetTeamMembership(teamID, userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if target.Role != model.TeamRoleMember && manager.Role != model.TeamRoleOwner {
		common.ApiErrorMsg(c, model.ErrTeamAccessDenied.Error())
		return
	}
	if err := model.RemoveTeamMember(teamID, userID); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func FundTeam(c *gin.Context) {
	teamID, ok := teamIDFromParam(c)
	if !ok {
		return
	}
	if _, ok := requireTeamManager(c, teamID); !ok {
		return
	}
	var request teamFundRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}
	requestID := strings.TrimSpace(request.RequestId)
	if len(requestID) > 64 {
		common.ApiErrorMsg(c, "request_id is too long")
		return
	}
	if requestID == "" {
		requestID = c.GetString(common.RequestIdKey)
	}
	if requestID == "" {
		requestID = common.NewRequestId()
	}
	if err := model.TransferUserQuotaToTeam(requestID, c.GetInt("id"), teamID, request.Quota); err != nil {
		common.ApiError(c, err)
		return
	}
	team, _, err := model.GetTeamForUser(teamID, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"balance_quota": team.Quota, "request_id": requestID})
}

func GetTeamUsage(c *gin.Context) {
	teamID, ok := teamIDFromParam(c)
	if !ok {
		return
	}
	if _, ok := requireTeamManager(c, teamID); !ok {
		return
	}
	month := strings.TrimSpace(c.Query("month"))
	rows, total, err := model.ListTeamUsage(teamID, month)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if month == "" {
		month = model.CurrentTeamMonth()
	}
	common.ApiSuccess(c, gin.H{"month": month, "total_usage_quota": total, "members": rows})
}
