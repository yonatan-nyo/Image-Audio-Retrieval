package routes

import (
	"bos/pablo/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupAlbumsRoutes(router *gin.RouterGroup, db *gorm.DB) {
	router.GET("/albums", controllers.GetAllAlbumsWithPagination(db))

	albums := router.Group("/albums")
	albums.GET("/:id", controllers.GetAlbumById(db))
	albums.GET("/:id/:songId", controllers.AssignSongToAlbum(db))
	albums.POST("/upload", controllers.UploadAndCreateAlbum(db))
	albums.POST("/search-by-image", controllers.SearchByImage(db))
}
