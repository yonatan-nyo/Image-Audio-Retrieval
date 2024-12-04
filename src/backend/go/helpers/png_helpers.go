package helpers

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/image/webp"
)

// ConvertToPng converts an uploaded image file to PNG format and saves it
func ConvertToPng(filePath string) (string, error) {
	log.Println("Starting the conversion process for:", filePath)

	// Open the image file
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening file:", err)
		return "", err
	}
	defer file.Close()

	// Decode the image based on its format
	var img image.Image
	ext := filepath.Ext(filePath)
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".gif":
		img, err = gif.Decode(file)
	case ".webp":
		img, err = webp.Decode(file) // Decode WebP format
	default:
		img, _, err = image.Decode(file) // default to generic decode for PNG, BMP, TIFF, etc.
	}

	if err != nil {
		log.Println("Error decoding image:", err)
		return "", err
	}

	// Prepare the output file path
	newFilePath := filePath[0:len(filePath)-len(ext)] + ".png"
	log.Println("Output file path:", newFilePath)

	// Create a new file to save the PNG image
	outFile, err := os.Create(newFilePath)
	if err != nil {
		log.Println("Error creating output file:", err)
		return "", err
	}
	defer outFile.Close()

	// Encode the image as PNG and save it
	err = png.Encode(outFile, img)
	if err != nil {
		log.Println("Error encoding image as PNG:", err)
		return "", err
	}

	// Log success and return the path to the newly created PNG image
	log.Println("Image successfully converted to PNG:", newFilePath)
	return newFilePath, nil
}
