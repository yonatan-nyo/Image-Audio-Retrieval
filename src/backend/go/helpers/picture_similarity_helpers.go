package helpers

import "log"

func CheckPictureSimilarity(hummingPicturePath, songPicturePath string) float64 {

	// count similarity
	similarity := 0.85 // Example score (in reality, this should be dynamically calculated)

	// You could add error handling here in case audio processing fails
	if similarity < 0 {
		log.Println("Error calculating similarity.")
		return 0
	}

	return similarity
}
