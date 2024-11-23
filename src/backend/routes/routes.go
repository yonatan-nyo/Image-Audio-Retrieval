package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(router *gin.Engine, db *gorm.DB) {
	router.GET("/", func(c *gin.Context) {
		c.String(200, "Running...")
	})

	api := router.Group("/api")
	{
		SetupUploadRoutes(api, db)
	}

}
