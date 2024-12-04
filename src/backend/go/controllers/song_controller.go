package controllers

import (
	"bos/pablo/helpers"
	"bos/pablo/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAllSongsWithPagination(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get pagination parameters from the request query, with defaults
		var page, pageSize int
		var err error
		var search string

		if page, err = strconv.Atoi(c.DefaultQuery("page", "1")); err != nil || page < 1 {
			page = 1
		}
		if pageSize, err = strconv.Atoi(c.DefaultQuery("page_size", "10")); err != nil || pageSize < 1 {
			pageSize = 10
		}

		if search = c.DefaultQuery("search", ""); search == "" {
			search = "%"
		}

		// Enforce maximum page size of 10
		if pageSize > 10 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize

		// Get the total count of matching records
		var totalItems int64
		if err := db.Model(&models.Song{}).Where("name LIKE ?", search).Count(&totalItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve record count"})
			return
		}

		// Retrieve the paginated records
		modelSlice := &[]models.Song{}
		query := db.Order("id DESC").Where("name LIKE ?", search)
		if err := query.Limit(pageSize).Offset(offset).Find(modelSlice).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve records"})
			return
		}

		// Include metadata in the response
		c.JSON(http.StatusOK, gin.H{
			"totalItems": totalItems,
			"page":       page,
			"pageSize":   pageSize,
			"data":       modelSlice,
		})
	}
}

func UploadAndCreateSong(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		relativePath := "songs"

		// Call the helper function to save or extract the uploaded file
		extractedPaths, err := helpers.SaveUploadedFile(c, "public/uploads", relativePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// If ZIP extraction occurred, bulk create songs
		if len(extractedPaths) > 1 {
			var songs []models.Song
			for _, filePath := range extractedPaths {
				// Convert each extracted file to .midi if needed
				convertedMidiPath, err := convertToMidi(filePath)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert file to MIDI"})
					return
				}

				// Extract the filename
				fileName := filepath.Base(convertedMidiPath)

				song := models.Song{
					Name:              fileName,
					AudioFilePath:     filePath,
					AudioFilePathMidi: convertedMidiPath,
				}
				songs = append(songs, song)
			}

			// Bulk create songs in the database
			if err := db.Create(&songs).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bulk create songs"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":        "ZIP file uploaded and extracted successfully",
				"extractedFiles": extractedPaths,
			})

		} else {
			// Handle non-ZIP files (e.g., .midi or other audio files)
			ext := strings.ToLower(filepath.Ext(extractedPaths[0]))
			var convertedMidiPath string

			if ext != ".midi" {
				// Convert the uploaded file to .midi (use an external tool or library)
				convertedMidiPath, err = convertToMidi(extractedPaths[0])
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert file to MIDI"})
					return
				}
			} else {
				convertedMidiPath = extractedPaths[0] // No conversion needed if already in .midi format
			}

			// Extract the filename
			fileName := filepath.Base(convertedMidiPath)

			// Create the song record
			song := models.Song{
				Name:              fileName,
				AudioFilePath:     extractedPaths[0],
				AudioFilePathMidi: convertedMidiPath,
			}

			if err := db.Create(&song).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create song"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":  "File uploaded and song created successfully",
				"path":     convertedMidiPath,
				"filename": fileName,
			})
		}

	}
}

func convertToMidi(audioPath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(audioPath))
	if ext == ".mid" {
		log.Println("File is already a MIDI file:", audioPath)
		return audioPath, nil
	}

	// URL of the external FastAPI service
	apiURL := "http://127.0.0.1:8000/convert-to-midi/"

	// Prepare the request payload using json.Marshal to handle escaping automatically
	payload := map[string]string{"file_path": audioPath}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal request body: %v\n", err)
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(requestBody)))
	if err != nil {
		log.Printf("Failed to create HTTP request: %v\n", err)
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send request to API: %v\n", err)
		return "", fmt.Errorf("failed to send request to API: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v\n", err)
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Print the response body to see its content
	log.Printf("Response body: %s\n", string(respBody))

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("API returned non-200 status code: %d\n", resp.StatusCode)
		return "", fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	// Parse the response
	var response struct {
		FullPath string `json:"full_path"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Printf("Failed to parse API response: %v\n", err)
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}

	// Log the returned full path
	log.Printf("MIDI conversion successful, full path: %s\n", response.FullPath)
	return response.FullPath, nil
}
