package onboarding

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/tuannm99/podzone/pkg/pdhttp"
)

func provideCORSMiddleware() pdhttp.Middleware {
	return func(router *gin.Engine) {
		router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "OPTIONS"},
			AllowHeaders:     []string{"Authorization", "Content-Type", "X-Tenant-ID"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: false,
		}))
	}
}
