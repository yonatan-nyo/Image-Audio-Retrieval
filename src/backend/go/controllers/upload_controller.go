package controllers

import (
	"bos/pablo/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		relativePath := c.Param("filepath")
		if relativePath == "" || relativePath == "/" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
			return
		}

		helpers.ServeFile(c, "public/uploads", relativePath, "public/not-found.jpg")
	}
}

func UploadFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the relative path from the URL parameters
		relativePath := c.Param("filepath")
		if relativePath == "" || relativePath == "/" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Target path is required"})
			return
		}

		// Call the helper function to save or extract the uploaded file
		extractedPaths, err := helpers.SaveUploadedFile(c, "public/uploads", relativePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// If there are extracted files, return a response indicating ZIP extraction
		if len(extractedPaths) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"message":        "ZIP file uploaded and extracted successfully",
				"extractedFiles": extractedPaths,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message":  "File uploaded successfully",
				"path":     extractedPaths[0],
				"filename": extractedPaths[0],
			})
		}
	}
}

func DeleteFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		relativePath := c.Param("filepath")
		if relativePath == "" || relativePath == "/" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File path is required"})
			return
		}

		err := helpers.DeleteFile("public/uploads", relativePath)
		if err != nil {
			if err.Error() == "file not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully", "path": relativePath})
	}
}
