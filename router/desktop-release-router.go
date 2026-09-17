package router

import (
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-gonic/gin"
)

func SetDesktopReleaseRouter(router *gin.Engine) {
	internal := router.Group("/v1/internal/releases")
	internal.Use(middleware.RouteTag("internal"), middleware.DesktopReleaseInternalAuth())
	{
		internal.GET("", controller.ListDesktopReleases)
		internal.POST("", controller.UploadDesktopReleaseDraft)
		internal.POST("/:version/publish", controller.PublishDesktopRelease)
	}

	public := router.Group("")
	public.Use(middleware.RouteTag("web"))
	{
		public.GET("/downloads.json", controller.GetDesktopDownloadManifest)
		public.GET("/release-artifacts/:version/:target", controller.DownloadDesktopReleaseArtifact)
	}
}
