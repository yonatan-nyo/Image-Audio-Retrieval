package helpers

import (
	"bos/pablo/types"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
)

// ImageProcessor handles image preprocessing and PCA
type ImageProcessor struct {
	images        *types.Matrix
	imageNames    []string
	meanPixel     []float64
	pcaComponents [][]float64
	imageSize     image.Point
}

func CheckPictureSimilarity(uploadedPicturePath, albumPicturePath string) float64 {
	// Set standard image size for processing
	width, height := 120, 120

	// Create a new image processor with the uploads folder
	processor, err := NewImageProcessor("public/uploads", width, height)
	if err != nil {
		log.Printf("Error initializing image processor: %v", err)
		return 0.0
	}

	// Find similar images to the song picture
	similarImages, distances, err := processor.FindSimilarImages(uploadedPicturePath, 1)
	if err != nil {
		log.Printf("Error finding similar images: %v", err)
		return 0.0
	}

	// If no similar images found, return low similarity
	if len(distances) == 0 {
		return 0.0
	}

	// Convert distance to similarity score
	// Lower distance means higher similarity
	// We'll use an exponential decay function to convert distance to similarity
	maxSimilarityDistance := 10.0
	similarity := math.Max(0, 1-(distances[0]/maxSimilarityDistance))

	log.Printf("Similar Image: %s, Distance: %.4f, Similarity Score: %.4f",
		similarImages[0], distances[0], similarity)

	return similarity
}

// NewImageProcessor creates a new ImageProcessor
func NewImageProcessor(folderPath string, width, height int) (*ImageProcessor, error) {
	processor := &ImageProcessor{
		imageSize: image.Point{X: width, Y: height},
	}

	err := processor.loadImagesFromFolder(folderPath)
	if err != nil {
		return nil, err
	}

	err = processor.preprocessImages()
	if err != nil {
		return nil, err
	}

	return processor, nil
}

// loadImagesFromFolder reads images from a folder and converts them to grayscale
func (ip *ImageProcessor) loadImagesFromFolder(folderPath string) error {
	var imagesList [][]float64

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if file is an image
		ext := filepath.Ext(path)
		if ext == ".png" {
			img, err := loadImage(path)
			if err != nil {
				return err
			}

			grayImg := convertToGrayscale(img)
			resizedImg := resizeImage(grayImg, ip.imageSize)
			flattenedImg := flattenImage(resizedImg)

			imagesList = append(imagesList, flattenedImg)
			ip.imageNames = append(ip.imageNames, filepath.Base(path))
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error loading images: %v", err)
	}

	// Convert images list to Matrix
	ip.images = types.NewMatrix(imagesList)

	return nil
}

// preprocessImages standardizes the image dataset
func (ip *ImageProcessor) preprocessImages() error {
	if ip.images.Rows() == 0 {
		return fmt.Errorf("no images loaded")
	}
	// Compute mean pixel values
	ip.meanPixel = computeMeanPixel(ip.images).GetRow(0)

	// Center the data
	meanPixelMatrix := types.NewMatrix([][]float64{ip.meanPixel})
	centeredImages := centerData(ip.images, meanPixelMatrix)

	// Perform PCA
	components, _, err := randomizedSVD(centeredImages, 100)
	if err != nil {
		return err
	}

	ip.pcaComponents = components
	return nil
}

// findSimilarImages finds images similar to the query image
func (ip *ImageProcessor) FindSimilarImages(queryImagePath string, topK int) ([]string, []float64, error) {
	// Load and preprocess query image
	queryImg, err := loadImage(queryImagePath)
	if err != nil {
		return nil, nil, err
	}

	grayImg := convertToGrayscale(queryImg)
	resizedImg := resizeImage(grayImg, ip.imageSize)
	flattenedQuery := flattenImage(resizedImg)

	// Convert flattened query to matrix
	queryMatrix := types.NewMatrix([][]float64{flattenedQuery})

	// Center the query image
	meanPixelMatrix := types.NewMatrix([][]float64{ip.meanPixel})
	centeredQuery := subtractVector(queryMatrix, meanPixelMatrix)

	// Project query to PCA space
	pcaComponentsMatrix := types.NewMatrix(ip.pcaComponents)
	queryPCASpace := projectToPCASpace(centeredQuery, pcaComponentsMatrix)

	// Compute distances
	distances := make([]float64, ip.images.Rows())
	for i := 0; i < ip.images.Rows(); i++ {
		rowImg := ip.images.GetRow(i)
		meanPixelMatrix := types.NewMatrix([][]float64{ip.meanPixel})
		centeredImg := subtractVector(types.NewMatrix([][]float64{rowImg}), meanPixelMatrix)
		pcaComponentsMatrix := types.NewMatrix(ip.pcaComponents)
		imgPCASpace := projectToPCASpace(centeredImg, pcaComponentsMatrix)
		distances[i] = euclideanDistance(queryPCASpace, imgPCASpace)
	}

	// Sort and return top K results
	type result struct {
		index    int
		distance float64
	}

	results := make([]result, len(distances))
	for i, dist := range distances {
		results[i] = result{index: i, distance: dist}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].distance < results[j].distance
	})

	similarImages := make([]string, topK)
	similarDistances := make([]float64, topK)
	for i := 0; i < topK; i++ {
		similarImages[i] = ip.imageNames[results[i].index]
		similarDistances[i] = results[i].distance
	}

	return similarImages, similarDistances, nil
}

// Utility functions for image processing and math
func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

func convertToGrayscale(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			lum := 0.2989*float64(r) + 0.5870*float64(g) + 0.1140*float64(b)
			gray.Set(x, y, color.Gray{Y: uint8(lum / 256)})
		}
	}
	return gray
}

func resizeImage(img *image.Gray, size image.Point) *image.Gray {
	resized := image.NewGray(image.Rect(0, 0, size.X, size.Y))
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			srcX := x * img.Bounds().Dx() / size.X
			srcY := y * img.Bounds().Dy() / size.Y
			resized.Set(x, y, img.At(srcX, srcY))
		}
	}
	return resized
}

func flattenImage(img *image.Gray) []float64 {
	bounds := img.Bounds()
	flattened := make([]float64, bounds.Dx()*bounds.Dy())
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			flattened[idx] = float64(img.GrayAt(x, y).Y)
			idx++
		}
	}
	return flattened
}

func computeMeanPixel(images *types.Matrix) *types.Matrix {
	if images.Rows() == 0 {
		return nil
	}

	pixelCount := images.Cols()
	meanPixel := make([]float64, pixelCount)

	for i := 0; i < images.Rows(); i++ {
		row := images.GetRow(i)
		for j := 0; j < pixelCount; j++ {
			meanPixel[j] += row[j]
		}
	}

	for j := 0; j < pixelCount; j++ {
		meanPixel[j] /= float64(images.Rows())
	}

	return types.NewMatrix([][]float64{meanPixel})
}

func centerData(images *types.Matrix, meanPixel *types.Matrix) *types.Matrix {
	centeredImages := make([][]float64, images.Rows())
	meanRow := meanPixel.GetRow(0)

	for i := 0; i < images.Rows(); i++ {
		row := images.GetRow(i)
		centeredRow := make([]float64, len(row))
		for j := range row {
			centeredRow[j] = row[j] - meanRow[j]
		}
		centeredImages[i] = centeredRow
	}

	return types.NewMatrix(centeredImages)
}

func subtractVector(vec *types.Matrix, mean *types.Matrix) *types.Matrix {
	if vec.Rows() == 0 || mean.Rows() == 0 {
		return vec
	}

	meanRow := mean.GetRow(0)
	result := make([][]float64, vec.Rows())

	for i := 0; i < vec.Rows(); i++ {
		row := vec.GetRow(i)
		resultRow := make([]float64, len(row))
		for j := range row {
			resultRow[j] = row[j] - meanRow[j]
		}
		result[i] = resultRow
	}

	return types.NewMatrix(result)
}

func projectToPCASpace(vec *types.Matrix, components *types.Matrix) []float64 {
	if vec.Rows() == 0 || components.Rows() == 0 {
		return nil
	}

	// Ensure the projection works for a matrix, not just a slice
	queryVec := vec.GetRow(0)
	projectedVec := make([]float64, components.Cols())

	for j := 0; j < components.Cols(); j++ {
		for i := 0; i < components.Rows(); i++ {
			projectedVec[j] += queryVec[i] * components.Get(i, j)
		}
	}

	return projectedVec
}

func euclideanDistance(vec1, vec2 []float64) float64 {
	var sumSquaredDiff float64
	for i := range vec1 {
		diff := vec1[i] - vec2[i]
		sumSquaredDiff += diff * diff
	}
	return math.Sqrt(sumSquaredDiff)
}

// Simplified SVD implementation using power iteration method
func randomizedSVD(data *types.Matrix, k int) ([][]float64, []float64, error) {
	if data.Rows() == 0 || data.Cols() == 0 {
		return nil, nil, fmt.Errorf("empty data matrix")
	}

	rows, cols := data.Rows(), data.Cols()
	k = min(k, min(rows, cols))

	// Random initialization
	components := make([][]float64, k)
	singularValues := make([]float64, k)

	for i := 0; i < k; i++ {
		components[i] = make([]float64, cols)
		// Random initialization
		for j := range components[i] {
			components[i][j] = math.Pow(-1, float64(i+j))
		}

		// Power iteration
		for iter := 0; iter < 5; iter++ {
			newComponent := make([]float64, cols)
			for l := 0; l < rows; l++ {
				for j := 0; j < cols; j++ {
					newComponent[j] += data.Get(l, j) * components[i][j]
				}
			}

			// Orthogonalize against previous components
			for p := 0; p < i; p++ {
				proj := dotProduct(newComponent, components[p])
				for j := range newComponent {
					newComponent[j] -= proj * components[p][j]
				}
			}

			// Normalize
			norm := vectorNorm(newComponent)
			for j := range newComponent {
				components[i][j] = newComponent[j] / norm
			}
		}

		// Compute singular value
		singularValues[i] = computeSingularValue(data.ToSlice(), components[i])
	}

	return components, singularValues, nil
}

func dotProduct(vec1, vec2 []float64) float64 {
	var sum float64
	for i := range vec1 {
		sum += vec1[i] * vec2[i]
	}
	return sum
}

func vectorNorm(vec []float64) float64 {
	var sumSquared float64
	for _, v := range vec {
		sumSquared += v * v
	}
	return math.Sqrt(sumSquared)
}

func computeSingularValue(data [][]float64, component []float64) float64 {
	var maxVal float64
	for _, row := range data {
		val := dotProduct(row, component)
		maxVal = math.Max(maxVal, math.Abs(val))
	}
	return maxVal
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
