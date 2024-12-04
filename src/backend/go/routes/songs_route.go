package routes

import (
	"bos/pablo/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupSongsRoutes(router *gin.RouterGroup, db *gorm.DB) {
	router.GET("/songs", controllers.GetAllSongsWithPagination(db))

	songs := router.Group("/songs")
	songs.POST("/upload", controllers.UploadAndCreateSong(db))
}
