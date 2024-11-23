package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the `filepath` parameter
		relativePath := c.Param("filepath")
		if relativePath == "" || relativePath == "/" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
			return
		}

		// Construct the full file path
		filePath := filepath.FromSlash(filepath.Join("public", "uploads", relativePath))

		// Check if the file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// Use a placeholder if the file doesn't exist
			filePathNoImage := filepath.FromSlash(filepath.Join("public", "uploads", "placeholder", "noimage.gif"))
			file, err := os.Open(filePathNoImage)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer file.Close()

			c.Header("Content-Type", "image/webp")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePathNoImage)))
			if _, err := io.Copy(c.Writer, file); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		// Serve the requested file
		file, err := os.Open(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer file.Close()

		// Set appropriate headers based on file type
		mimeType := http.DetectContentType(make([]byte, 512))
		c.Header("Content-Type", mimeType)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePath)))
		if _, err := io.Copy(c.Writer, file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

// UploadFile handles file uploads with unique filenames.
func UploadFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the relative folder path
		relativePath := c.Param("filepath")
		if relativePath == "" || relativePath == "/" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Target path is required"})
			return
		}

		// Construct the full folder path
		fullPath := filepath.FromSlash(filepath.Join("public", "uploads", relativePath))

		// Parse uploaded file
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File upload is required"})
			return
		}

		// Ensure the folder exists
		if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Generate a truly unique filename
		var uniqueName string
		ext := filepath.Ext(file.Filename)                                     // Get the file extension
		baseName := filepath.Base(file.Filename[:len(file.Filename)-len(ext)]) // Get the base name

		for {
			uniqueName = baseName + "-" + uuid.New().String() + ext
			destPath := filepath.Join(fullPath, uniqueName)
			// Check if the file already exists
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				break // Exit the loop if the file does not exist
			}
		}

		// Save the uploaded file
		destPath := filepath.Join(fullPath, uniqueName)
		if err := c.SaveUploadedFile(file, destPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "File uploaded successfully",
			"path":     destPath,
			"filename": uniqueName,
		})
	}
}

// DeleteFile handles file deletion.
func DeleteFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the relative file path
		relativePath := c.Param("filepath")
		if relativePath == "" || relativePath == "/" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
			return
		}

		// Construct the full file path
		filePath := filepath.FromSlash(filepath.Join("public", "uploads", relativePath))

		// Check if the file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		// Delete the file
		if err := os.Remove(filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully", "path": filePath})
	}
}
