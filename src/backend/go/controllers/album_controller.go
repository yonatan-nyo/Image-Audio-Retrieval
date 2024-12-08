package controllers

import (
	"bos/pablo/helpers"
	"bos/pablo/models"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetAllAlbumsWithPagination fetches all albums with pagination and search functionality.
func GetAllAlbumsWithPagination(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var page, pageSize int
		var err error
		var search string

		// Pagination parameters
		if page, err = strconv.Atoi(c.DefaultQuery("page", "1")); err != nil || page < 1 {
			page = 1
		}
		if pageSize, err = strconv.Atoi(c.DefaultQuery("page_size", "10")); err != nil || pageSize < 1 {
			pageSize = 10
		}
		search = c.DefaultQuery("search", "")
		if search == "" {
			search = "%"
		} else {
			search = "%" + search + "%"
		}

		// Enforce maximum page size of 10
		if pageSize > 10 {
			pageSize = 10
		}

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

		// Include metadata in the response
		c.JSON(http.StatusOK, gin.H{
			"totalItems": totalItems,
			"page":       page,
			"pageSize":   pageSize,
			"data":       albums,
		})
	}
}

// UploadAndCreateAlbum uploads an album file (e.g., audio, zip) and creates an album record.
func UploadAndCreateAlbum(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		relativePath := "albums"

		// Call the helper function to save or extract the uploaded file
		extractedPaths, err := helpers.SaveUploadedFile(c, "public/uploads", relativePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// If ZIP extraction occurred, bulk create albums
		if len(extractedPaths) > 1 {
			var albums []models.Album
			for _, filePath := range extractedPaths {
				// Convert each extracted file to PNG if needed
				convertedPngPath, err := helpers.ConvertToPng(filePath)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert file to PNG"})
					return
				}

				// If the original file is not PNG, delete it
				if strings.ToLower(filepath.Ext(filePath)) != ".png" {
					err := os.Remove(filePath)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete original file"})
						return
					}
				}

				// Extract the filename
				fileName := filepath.Base(convertedPngPath)

				album := models.Album{
					Name:        fileName,
					PicFilePath: convertedPngPath,
				}
				albums = append(albums, album)
			}

			// Bulk create albums in the database
			if err := db.Create(&albums).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk create albums"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":        "ZIP file uploaded and extracted successfully",
				"extractedFiles": extractedPaths,
			})

		} else {
			// Handle non-ZIP files (e.g., .midi or other audio files)
			ext := strings.ToLower(filepath.Ext(extractedPaths[0]))
			var convertedPngPath string

			if ext != ".png" {
				// Convert the uploaded file to PNG
				convertedPngPath, err = helpers.ConvertToPng(extractedPaths[0])
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert file to PNG"})
					return
				}

				// If the original file is not PNG, delete it
				if strings.ToLower(filepath.Ext(extractedPaths[0])) != ".png" {
					err := os.Remove(extractedPaths[0])
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete original file"})
						return
					}
				}
			} else {
				convertedPngPath = extractedPaths[0] // No conversion needed if already in .png format
			}

			// Extract the filename
			fileName := filepath.Base(convertedPngPath)

			// Create the album record
			album := models.Album{
				Name:        fileName,
				PicFilePath: convertedPngPath,
			}

			if err := db.Create(&album).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create album"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":  "File uploaded and album created successfully",
				"path":     convertedPngPath,
				"filename": fileName,
			})
		}
	}
}

// SearchByImage handles the search functionality for image similarity (e.g., album cover similarity)
func SearchByImage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadFolder := "images" // Folder to save the uploaded images

		log.Println("Step 1: Starting image upload process")

		// Save the uploaded image file
		uploadedFilePaths, err := helpers.SaveUploadedFile(c, "public/uploads", uploadFolder)
		if err != nil {
			log.Printf("Error in saving uploaded file: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		imageFilePath := uploadedFilePaths[0] // Get the path of the uploaded image
		log.Printf("Step 2: Image uploaded successfully, file path: %s\n", imageFilePath)

		// Convert the image to PNG format if it's not already in PNG format
		fileExtension := strings.ToLower(filepath.Ext(imageFilePath))
		log.Printf("Step 3: Checking file extension: %s\n", fileExtension)

		if fileExtension != ".png" {
			// Convert the image to PNG
			log.Println("Step 4: Converting image to PNG format")
			convertedImagePath, err := helpers.ConvertToPng(imageFilePath)
			if err != nil {
				log.Printf("Error in converting image to PNG: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert image to PNG"})
				return
			}
			// Replace the original file path with the new converted PNG file path
			imageFilePath = convertedImagePath
			log.Printf("Step 5: Image converted successfully, new path: %s\n", imageFilePath)
		}

		// Fetch all albums from the database (not songs)
		log.Println("Step 6: Fetching all albums from the database")
		var albums []models.Album
		err = db.Find(&albums).Error
		if err != nil {
			log.Printf("Error in fetching albums from database: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch albums"})
			return
		}

		log.Printf("Step 7: Successfully fetched %d albums from the database\n", len(albums))

		var matchedAlbums []models.Album

		// Calculate similarity between the uploaded image and each album's image
		log.Println("Step 8: Calculating similarity between uploaded image and albums' cover images")
		
		for _, album := range albums {
			if _, err := os.Stat(album.PicFilePath); os.IsNotExist(err) {
        log.Printf("Album image not found: %s\n", album.PicFilePath)
        continue
			}
			similarityScore := helpers.CheckPictureSimilarity(imageFilePath, album.PicFilePath) // Compare with the album's image
			log.Printf("Checking similarity for %s %s\n", imageFilePath, album.PicFilePath)
			if similarityScore > 0.8 { // If similarity is greater than 80%
				matchedAlbums = append(matchedAlbums, album)
			}
		}

		// Limit the result to the top 9 most similar albums
		if len(matchedAlbums) > 9 {
			matchedAlbums = matchedAlbums[:9]
		}
		log.Printf("Step 9: Found %d matched albums\n", len(matchedAlbums))

		// Delete the uploaded image after comparison
		log.Println("Step 10: Deleting the uploaded image after comparison")
		err = os.Remove(imageFilePath) // Remove the uploaded file
		if err != nil {
			log.Printf("Error in deleting uploaded file: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the uploaded file"})
			return
		}

		// Respond with the matched albums
		if len(matchedAlbums) > 0 {
			log.Println("Step 11: Returning matched albums as response")
			c.JSON(http.StatusOK, gin.H{"data": matchedAlbums})
		} else {
			log.Println("Step 12: No similar albums found")
			c.JSON(http.StatusNotFound, gin.H{"message": "No similar albums found"})
		}
	}
}
