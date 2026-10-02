package controller

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

var desktopSkillID = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func GetDesktopSkillAccess(c *gin.Context) { desktopSkillConsumer(c, "access") }
func ClaimDesktopSkill(c *gin.Context)     { desktopSkillConsumer(c, "claim") }
func DownloadDesktopSkill(c *gin.Context)  { desktopSkillConsumer(c, "download") }

// The platform verifies its own Desktop session; the content service owns rights.
// Neither a personal API key nor a browser-supplied account ID is a delegation.
func desktopSkillConsumer(c *gin.Context, action string) {
	c.Header("Cache-Control", "no-store")
	identity, ok := middleware.GetSessionAuthIdentity(c)
	if !ok {
		writeAuthSessionError(c, service.ErrAuthTokenInvalid)
		return
	}
	session, _, err := service.ValidateLoginSession(identity)
	if err != nil || session.LoginMethod != service.DesktopLoginMethod || !service.CentralAccountEnabled() ||
		session.AuthorityIssuer != service.CentralAccountIssuer() || session.AuthoritySubject == "" ||
		session.AuthoritySessionID == "" || session.AuthorityAuthVersion <= 0 {
		writeAuthSessionError(c, service.ErrAuthTokenInvalid)
		return
	}
	id := c.Param("id")
	if len(id) > 64 || !desktopSkillID.MatchString(id) || c.Request.URL.RawQuery != "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "SKILL_REQUEST_INVALID"})
		return
	}
	upstream := strings.TrimSpace(os.Getenv("ROBO_SKILL_SERVICE_URL"))
	if upstream == "" {
		upstream = "http://127.0.0.1:8766"
	}
	base, err := url.Parse(upstream)
	credential := strings.TrimSpace(os.Getenv("ROBO_SKILL_CONSUMER_TOKEN"))
	if err != nil || base.Scheme != "http" || (base.Hostname() != "127.0.0.1" && base.Hostname() != "::1") ||
		base.User != nil || (base.Path != "" && base.Path != "/") || base.RawQuery != "" || base.Fragment != "" || credential == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "SKILL_SERVICE_UNAVAILABLE"})
		return
	}
	delegation, err := common.Marshal(gin.H{"issuer": session.AuthorityIssuer, "subject": session.AuthoritySubject,
		"sessionId": session.AuthoritySessionID, "authVersion": session.AuthorityAuthVersion})
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "SKILL_SERVICE_UNAVAILABLE"})
		return
	}
	base.Path = "/v1/skill-consumer/skills/" + id + "/" + action
	var payload io.Reader
	if c.Request.Method == http.MethodPost {
		payload = strings.NewReader("{}")
	}
	request, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, base.String(), payload)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "SKILL_SERVICE_UNAVAILABLE"})
		return
	}
	request.Header.Set("X-Skill-Consumer-Token", credential)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Skill-Consumer-Identity", base64.RawURLEncoding.EncodeToString(delegation))
	client := &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": "SKILL_SERVICE_UNAVAILABLE"})
		return
	}
	defer response.Body.Close()
	limit := int64(1 << 20)
	contentType := "application/json"
	if action == "download" && response.StatusCode == http.StatusOK {
		if response.Header.Get("Content-Type") != "application/zip" {
			c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": "SKILL_PACKAGE_INVALID"})
			return
		}
		limit, contentType = 10<<20, "application/zip"
	} else if !strings.HasPrefix(response.Header.Get("Content-Type"), "application/json") {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": "SKILL_RESPONSE_INVALID"})
		return
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || int64(len(body)) > limit || response.StatusCode < 200 || response.StatusCode >= 500 || (response.StatusCode >= 300 && response.StatusCode < 400) {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": "SKILL_SERVICE_UNAVAILABLE"})
		return
	}
	if contentType == "application/zip" {
		c.Header("Content-Disposition", `attachment; filename="`+id+`.zip"`)
	}
	c.Data(response.StatusCode, contentType, body)
}
