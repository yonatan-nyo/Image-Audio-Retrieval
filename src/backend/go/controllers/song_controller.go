package controllers

import (
	"bos/pablo/helpers"
	"bos/pablo/models"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetUnassociatedSongs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var songs []models.Song
		err := db.Where("album_id IS NULL").Limit(5).Find(&songs).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch unassociated songs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": songs})
	}
}

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
				convertedMidiPath, jsonPath, err := helpers.ConvertToMidi(filePath)

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
					MidiJSON:          jsonPath, //change this
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
			// ext := strings.ToLower(filepath.Ext(extractedPaths[0]))
			var convertedMidiPath string
			var jsonPath string

			// Convert the uploaded file to .midi (use an external tool or library)
			convertedMidiPath, jsonPath, err = helpers.ConvertToMidi(extractedPaths[0])
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert file to MIDI"})
				return
			}

			// Extract the filename
			fileName := filepath.Base(convertedMidiPath)

			// Create the song record
			song := models.Song{
				Name:              fileName,
				AudioFilePath:     extractedPaths[0],
				AudioFilePathMidi: convertedMidiPath,
				MidiJSON:          jsonPath, //change this
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

// SearchByHumming handles the search functionality for humming or audio file similarity
func SearchByHumming(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadFolder := "hummings"

		// Save the uploaded humming or audio file
		uploadedFilePaths, err := helpers.SaveUploadedFile(c, "public/uploads", uploadFolder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// defer delete after
		defer func() {
			for _, path := range uploadedFilePaths {
				err := os.Remove(path)
				if err != nil {
					fmt.Println("Failed to delete file: ", path)
				}
			}
		}()

		audioFilePath := uploadedFilePaths[0]

		// Check if the file needs to be converted to MIDI
		var jsonHummingPath string
		_, jsonHummingPath, err = helpers.ConvertToMidi(audioFilePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to convert file to MIDI"})
			return
		}

		// Fetch all songs from the database
		var songs []models.Song
		err = db.Find(&songs).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch songs"})
			return
		}

		// Channel to collect results from goroutines
		type MatchResult struct {
			ID              uint    `json:"ID"`
			Name            string  `json:"Name"`
			AudioFilePath   string  `json:"AudioFilePath"`
			AlbumID         *uint   `json:"AlbumID,omitempty"`
			SimilarityScore float64 `json:"SimilarityScore"`
		}
		resultChan := make(chan MatchResult, len(songs))

		// start benchmarking
		startTime := time.Now()

		// WaitGroup to wait for all goroutines to finish
		var wg sync.WaitGroup

		// Calculate similarity between the uploaded MIDI and each song's MIDI concurrently
		for _, song := range songs {
			wg.Add(1)
			go func(song models.Song) {
				defer wg.Done()
				// Log checking similarity
				fmt.Printf("[a]: %s\n[b]: %s\n", jsonHummingPath, song.MidiJSON)

				// Calculate similarity score
				similarityScore := helpers.CheckAudioSimilarity(jsonHummingPath, song.MidiJSON)
				if similarityScore > 0.0 {
					// Send matched song with score to the channel
					resultChan <- MatchResult{
						ID:              song.ID,
						Name:            song.Name,
						AudioFilePath:   song.AudioFilePath,
						AlbumID:         song.AlbumID,
						SimilarityScore: similarityScore,
					}
				}
				fmt.Printf("Checking similarity %s = %f\n", song.Name, similarityScore)
			}(song)
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(resultChan)

		// Collect results from the channel
		var matchedResults []MatchResult
		for result := range resultChan {
			matchedResults = append(matchedResults, result)
		}

		// Sort matched results by similarity score in descending order
		sort.Slice(matchedResults, func(i, j int) bool {
			return matchedResults[i].SimilarityScore > matchedResults[j].SimilarityScore
		})

		// Limit the result to the top 9 most similar songs
		if len(matchedResults) > 9 {
			matchedResults = matchedResults[:9]
		}

		// Delete the uploaded humming/audio file after comparison
		err = os.Remove(audioFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the uploaded file"})
			return
		}

		// Respond with the matched songs and their similarity scores
		if len(matchedResults) > 0 {
			c.JSON(http.StatusOK, gin.H{"data": matchedResults, "time": time.Since(startTime).Seconds()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"message": "No similar songs found"})
		}
	}
}
