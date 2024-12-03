package helpers

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// ExtractZip extracts a ZIP file, placing all files directly into the specified destination folder,
// ensuring filenames are unique.
func ExtractZip(zipPath, destDir string) ([]string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var filePaths []string
	for _, f := range r.File {
		// Skip directories
		if f.FileInfo().IsDir() {
			continue
		}

		// Get the base name of the file (excluding directories)
		ext := filepath.Ext(f.Name)
		baseName := filepath.Base(f.Name[:len(f.Name)-len(ext)])

		// Generate a unique filename
		destPath := filepath.Join(destDir, baseName+ext)

		// Ensure the file name is unique by appending a UUID if a file with the same name exists
		for {
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				break // If the file does not exist, we're good to go
			}
			// Append a unique identifier to the filename
			uniqueName := fmt.Sprintf("%s-%s%s", baseName, uuid.New().String(), ext)
			destPath = filepath.Join(destDir, uniqueName)
		}

		// Create the destination file
		destFile, err := os.Create(destPath)
		if err != nil {
			return nil, err
		}

		// Open the source file from the ZIP archive
		srcFile, err := f.Open()
		if err != nil {
			return nil, err
		}

		// Copy the contents of the ZIP file to the destination file
		_, err = io.Copy(destFile, srcFile)
		destFile.Close()
		srcFile.Close()

		if err != nil {
			return nil, err
		}

		// Append the unique file path to the result list
		filePaths = append(filePaths, destPath)
	}

	return filePaths, nil
}
