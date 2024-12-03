package helpers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ServeFile serves a file if it exists, or a placeholder if it doesn't.
func ServeFile(c *gin.Context, baseDir, relativePath, placeholderPath string) {
	filePath := filepath.FromSlash(filepath.Join(baseDir, relativePath))

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		filePath = placeholderPath
	}

	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	mimeType := http.DetectContentType(make([]byte, 512))
	c.Header("Content-Type", mimeType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePath)))

	if _, err := io.Copy(c.Writer, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// SaveUploadedFile saves uploaded files or extracts ZIP files directly to the destination folder.
func SaveUploadedFile(c *gin.Context, baseDir, relativePath string) ([]string, error) {
	fullPath := filepath.FromSlash(filepath.Join(baseDir, relativePath))

	// Ensure the folder exists
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return nil, err
	}

	// Parse the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return nil, err
	}

	// Get the base name of the file (excluding extension)
	ext := filepath.Ext(file.Filename)
	baseName := filepath.Base(file.Filename[:len(file.Filename)-len(ext)])

	// Check if the file is a ZIP file
	if ext == ".zip" {
		// Save the ZIP file temporarily in memory or disk
		tempZipPath := filepath.Join(os.TempDir(), uuid.New().String()+".zip")
		if err := c.SaveUploadedFile(file, tempZipPath); err != nil {
			return nil, err
		}

		// Extract the ZIP contents to the destination folder
		extractedPaths, err := ExtractZip(tempZipPath, fullPath)
		if err != nil {
			return nil, err
		}

		// Remove the temporary ZIP file
		if err := os.Remove(tempZipPath); err != nil {
			return nil, err
		}

		return extractedPaths, nil
	}

	// If not a ZIP file, save directly with a unique name (basename + UUID)
	uniqueName := fmt.Sprintf("%s-%s%s", baseName, uuid.New().String(), ext)
	destPath := filepath.Join(fullPath, uniqueName)
	if err := c.SaveUploadedFile(file, destPath); err != nil {
		return nil, err
	}

	return []string{destPath}, nil
}

// DeleteFile deletes a file if it exists.
func DeleteFile(baseDir, relativePath string) error {
	filePath := filepath.FromSlash(filepath.Join(baseDir, relativePath))

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}

	return os.Remove(filePath)
}
