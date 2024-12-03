package routes

import (
	"bos/pablo/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupSongsRoutes(router *gin.RouterGroup, db *gorm.DB) {
	songs := router.Group("/songs")

	songs.GET("/", controllers.GetAllSongsWithPagination(db))
	songs.POST("/", controllers.CreateSong(db))
	songs.POST("/upload", controllers.UploadAndCreateSong(db))
}
