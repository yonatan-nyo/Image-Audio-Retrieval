package routes

import (
	"bos/pablo/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupUploadRoutes(router *gin.RouterGroup, db *gorm.DB) {
	uploads := router.Group("/uploads")

	uploads.GET("/*filepath", controllers.GetFile())
	uploads.POST("/*filepath", controllers.UploadFile())
	uploads.DELETE("/*filepath", controllers.DeleteFile())
}
