package router

import (
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-gonic/gin"
)

func setDesktopRouter(api *gin.RouterGroup) {
	desktop := api.Group("/desktop", middleware.DisableCache(), middleware.AnonymousRequestBodyLimit())
	desktop.POST("/device/code", middleware.CriticalRateLimit(), controller.CreateDesktopDeviceCode)
	// Polling is bounded per grant in the database and by the global API limit;
	// the login-attempt limiter is intentionally not reused for normal polling.
	desktop.POST("/device/token", controller.ExchangeDesktopDeviceCode)
	desktop.POST("/refresh", middleware.CriticalRateLimit(), controller.RefreshDesktopSession)
	desktop.GET("/device/authorization", middleware.UserAuth(), middleware.UserCriticalRateLimit("desktop-approval"), controller.GetDesktopDeviceAuthorization)
	desktop.POST("/device/authorization", middleware.UserAuth(), middleware.SessionCookieOriginGuard(), middleware.UserCriticalRateLimit("desktop-approval"), controller.ApproveDesktopDevice)
	desktop.POST("/relay-token", middleware.UserAuth(), middleware.UserCriticalRateLimit("desktop-credential"), controller.CreateDesktopRelayToken)
	desktop.DELETE("/session", middleware.UserAuth(), controller.DeleteDesktopSession)
}
