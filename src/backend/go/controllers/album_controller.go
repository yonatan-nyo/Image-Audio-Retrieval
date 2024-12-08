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
	"sync"
	"time"

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

		imageFilePath := uploadedFilePaths[0]
		log.Printf("Step 2: Image uploaded successfully, file path: %s\n", imageFilePath)

		// Convert the image to PNG format if necessary
		fileExtension := strings.ToLower(filepath.Ext(imageFilePath))
		if fileExtension != ".png" {
			log.Println("Step 3: Converting image to PNG format")
			convertedImagePath, err := helpers.ConvertToPng(imageFilePath)
			if err != nil {
				log.Printf("Error in converting image to PNG: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert image to PNG"})
				return
			}
			imageFilePath = convertedImagePath
		}

		// Preprocess the uploaded image
		uploadedImageVector, err := helpers.PreprocessImage(imageFilePath, 120, 120)
		if err != nil {
			log.Printf("Error in preprocessing image: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to preprocess image"})
			return
		}

		// Fetch all albums from the database
		log.Println("Step 4: Fetching all albums from the database")
		var albums []models.Album
		if err := db.Find(&albums).Error; err != nil {
			log.Printf("Error in fetching albums from database: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch albums"})
			return
		}
		log.Printf("Step 5: Successfully fetched %d albums from the database\n", len(albums))

		// Use a goroutine pool to parallelize similarity calculations
		var wg sync.WaitGroup
		results := make(chan models.Album, len(albums)) // Channel to collect matched albums

		// start benchmarking
		startTime := time.Now()

		for _, album := range albums {
			wg.Add(1)
			go func(album models.Album) {
				defer wg.Done()
				// Skip if album image is missing
				if _, err := os.Stat(album.PicFilePath); os.IsNotExist(err) {
					log.Printf("Album image not found: %s\n", album.PicFilePath)
					return
				}

				// Calculate similarity
				similarityScore := helpers.CheckPictureSimilarity(uploadedImageVector, album.PicFilePath)
				log.Printf("Similarity score for %d: %.4f\n", album.ID, similarityScore)

				if similarityScore > 0.8 { // If similarity is greater than 80%
					results <- album
				}
			}(album)
		}

		// Wait for all goroutines to complete
		go func() {
			wg.Wait()
			close(results) // Close the channel once all routines are done
		}()

		// Collect results
		var matchedAlbums []models.Album
		for album := range results {
			matchedAlbums = append(matchedAlbums, album)
		}

		// end benchmarking
		endTime := time.Since(startTime)

		// Limit to top 9 matches
		if len(matchedAlbums) > 9 {
			matchedAlbums = matchedAlbums[:9]
		}
		log.Printf("Step 6: Found %d matched albums\n", len(matchedAlbums))

		// Clean up uploaded image
		log.Println("Step 7: Deleting the uploaded image after comparison")
		if err := os.Remove(imageFilePath); err != nil {
			log.Printf("Error deleting uploaded file: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete uploaded file"})
			return
		}

		log.Printf("completed in %v\n", endTime)
		// Respond with results
		if len(matchedAlbums) > 0 {
			c.JSON(http.StatusOK, gin.H{"data": matchedAlbums, "time": endTime.Seconds()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"message": "No similar albums found"})
		}
	}
}
