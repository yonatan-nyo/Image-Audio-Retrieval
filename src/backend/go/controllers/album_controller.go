package controllers

import (
	"bos/pablo/helpers"
	"bos/pablo/models"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetAllAlbumsWithPagination fetches all albums with pagination and search functionality.
func GetAllAlbumsWithPagination(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var page, pageSize int
		var search string

		// Pagination parameters
		page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "10"))
		if pageSize > 10 {
			pageSize = 10
		}
		search = "%" + c.DefaultQuery("search", "") + "%"

		offset := (page - 1) * pageSize

		// Get the total count of albums
		var totalItems int64
		if err := db.Model(&models.Album{}).Where("name LIKE ?", search).Count(&totalItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve record count"})
			return
		}

		// Retrieve paginated albums
		albums := []models.Album{}
		if err := db.Order("id DESC").Where("name LIKE ?", search).Limit(pageSize).Offset(offset).Find(&albums).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve records"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"totalItems": totalItems,
			"page":       page,
			"pageSize":   pageSize,
			"data":       albums,
		})
	}
}

// GetAlbumById fetches a single album by ID.
func GetAlbumById(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var album models.Album

		if err := db.Preload("Songs").First(&album, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Record not found!"})
			return
		}

		c.JSON(http.StatusOK, album)
	}
}

// AssignSongToAlbum assigns a song to an album.
func AssignSongToAlbum(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		albumIDParam := c.Param("id")
		songIDParam := c.Param("songId")

		albumID, err := strconv.Atoi(albumIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid album ID format"})
			return
		}

		songID, err := strconv.Atoi(songIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid song ID format"})
			return
		}

		var album models.Album
		if err := db.First(&album, albumID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Album not found!"})
			return
		}

		var song models.Song
		if err := db.First(&song, songID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Song not found!"})
			return
		}

		// update the song record with the album ID
		if err := db.Model(&song).Update("album_id", albumID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign song to album"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Song assigned to album successfully"})
	}
}

// UploadAndCreateAlbum handles file uploads and album creation.
func UploadAndCreateAlbum(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		relativePath := "albums"
		flattenedDir := "public/uploads/flattened_albums"

		// Save uploaded file
		extractedPaths, err := helpers.SaveUploadedFile(c, "public/uploads", relativePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, filePath := range extractedPaths {
			convertedPngPath := filePath

			// Convert non-PNG files to PNG
			if strings.ToLower(filepath.Ext(filePath)) != ".png" {
				convertedPngPath, err = helpers.ConvertToPng(filePath)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert file to PNG"})
					return
				}
				os.Remove(filePath)
			}

			// Generate flattened vector
			flattenedVector, err := helpers.PreprocessImage(convertedPngPath, 120, 120)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to preprocess image"})
				return
			}

			// Save flattened vector to file
			os.MkdirAll(flattenedDir, os.ModePerm)
			flattenedFilePath := filepath.Join(flattenedDir, fmt.Sprintf("%s.json", filepath.Base(convertedPngPath)))
			flattenedFile, err := os.Create(flattenedFilePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save flattened vector to file"})
				return
			}

			flattenedData, _ := json.Marshal(flattenedVector)
			_, err = flattenedFile.Write(flattenedData)
			flattenedFile.Close()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write flattened vector to file"})
				return
			}

			// Create album record
			album := models.Album{
				Name:        filepath.Base(convertedPngPath),
				PicFilePath: convertedPngPath,
				Flattened:   flattenedFilePath,
			}

			if err := db.Create(&album).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create album"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Albums created successfully"})
	}
}

// SearchByImage finds albums with similar cover images.
func SearchByImage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadFolder := "images"

		// Save uploaded image
		uploadedFilePaths, err := helpers.SaveUploadedFile(c, "public/uploads", uploadFolder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		imageFilePath := uploadedFilePaths[0]
		defer os.Remove(imageFilePath)

		// Convert to PNG if necessary
		if strings.ToLower(filepath.Ext(imageFilePath)) != ".png" {
			imageFilePath, err = helpers.ConvertToPng(imageFilePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert image to PNG"})
				return
			}
		}

		// Preprocess uploaded image
		uploadedImageVector, err := helpers.PreprocessImage(imageFilePath, 120, 120)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to preprocess image"})
			return
		}

		// Fetch all albums
		var albums []models.Album
		if err := db.Find(&albums).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch albums"})
			return
		}

		// start benchmarking
		startTime := time.Now()

		// Calculate similarity scores
		var matchedAlbums []models.Album
		for _, album := range albums {
			if album.Flattened != "" {
				// Read flattened vector from file
				file, err := os.Open(album.Flattened)
				if err != nil {
					continue
				}
				defer file.Close()

				var albumVector []float64
				json.NewDecoder(file).Decode(&albumVector)

				// Compute similarity
				similarity := helpers.CheckPictureSimilarity(uploadedImageVector, albumVector)
				if similarity > 0.8 {
					matchedAlbums = append(matchedAlbums, album)
				}
			}
		}

		// Limit results to top 9 matches
		if len(matchedAlbums) > 9 {
			matchedAlbums = matchedAlbums[:9]
		}

		if len(matchedAlbums) > 0 {
			c.JSON(http.StatusOK, gin.H{"data": matchedAlbums, "time": time.Since(startTime).Seconds()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"message": "No similar albums found"})
		}
	}
}
