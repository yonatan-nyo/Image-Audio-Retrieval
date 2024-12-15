// picture_similarity_helpers.go contains helper functions for image preprocessing and PCA
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
)

func CheckPictureSimilarity(uploadPictureFlattened []float64, albumPictureFlattened []float64) float64 {
	// Load images into matrix
	uploadPictureMatrix := types.NewMatrix([][]float64{uploadPictureFlattened})
	albumPictureMatrix := types.NewMatrix([][]float64{albumPictureFlattened})

	// Compute mean pixel
	meanPixel := computeMeanPixel(types.NewMatrix([][]float64{uploadPictureFlattened, albumPictureFlattened}))

	// Center the data
	meanPixelMatrix := types.NewMatrix([][]float64{meanPixel.GetRow(0)})
	uploadPictureMatrix = subtractVector(uploadPictureMatrix, meanPixelMatrix)
	albumPictureMatrix = subtractVector(albumPictureMatrix, meanPixelMatrix)

	// Perform PCA
	components, _, err := randomizedSVD(types.NewMatrix([][]float64{uploadPictureFlattened, albumPictureFlattened}), 10)
	if err != nil {
		log.Fatalf("Error performing PCA: %v", err)
		return 0.0
	}

	pcaComponents := types.NewMatrix(components)

	// Project images to PCA space
	uploadPictureProjected := projectToPCASpace(uploadPictureMatrix, pcaComponents)
	albumPictureProjected := projectToPCASpace(albumPictureMatrix, pcaComponents)

	// Compute Euclidean distance
	distance := euclideanDistance(uploadPictureProjected, albumPictureProjected)

	maxSimilarity := 10.0
	similarity := math.Max(0, 1-(distance/maxSimilarity))

	log.Printf("Image Distance: %.4f, Similarity Score: %.4f", distance, similarity)

	return similarity
}

func PreprocessImage(imagePath string, width, height int) ([]float64, error) {
	// Load the image
	pictureImg, err := loadImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("error loading image from path %s: %w", imagePath, err)
	}
	// grayscale, resize, and flatten the image
	pictureFlattened := flattenImage(resizeImage(convertToGrayscale(pictureImg), image.Point{X: width, Y: height}))

	return pictureFlattened, nil
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
