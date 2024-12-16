package controllers

import (
	"bos/pablo/helpers"
	"bos/pablo/models"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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

		// Fetch all albums with their songs
		var albums []models.Album
		if err := db.Preload("Songs").Find(&albums).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch albums"})
			return
		}

		// Start benchmarking
		startTime := time.Now()

		// Calculate similarity scores
		var matchedAlbums []map[string]interface{}
		for _, album := range albums {
			if album.Flattened != "" {
				// Read flattened vector from file
				file, err := os.Open(album.Flattened)
				if err != nil {
					continue
				}
				defer file.Close()

				var albumVector []float64
				err = json.NewDecoder(file).Decode(&albumVector)
				if err != nil {
					continue
				}

				// Compute similarity
				similarity := helpers.CheckPictureSimilarity(uploadedImageVector, albumVector)
				if similarity > 0.8 {
					matchedAlbums = append(matchedAlbums, map[string]interface{}{
						"ID":          album.ID,
						"Name":        album.Name,
						"PicFilePath": album.PicFilePath,
						"Songs":       album.Songs,
						"similarity":  similarity,
					})
				}
			}
		}

		// Sort by similarity, handling type safety
		sort.Slice(matchedAlbums, func(i, j int) bool {
			simI := matchedAlbums[i]["similarity"].(float64)
			simJ := matchedAlbums[j]["similarity"].(float64)
			return simI > simJ
		})

		// Limit results to top 9 matches
		if len(matchedAlbums) > 9 {
			matchedAlbums = matchedAlbums[:9]
		}

		// Check results and respond
		if len(matchedAlbums) > 0 {
			c.JSON(http.StatusOK, gin.H{"data": matchedAlbums, "time": time.Since(startTime).Seconds()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"message": "No similar albums found"})
		}
	}
}
