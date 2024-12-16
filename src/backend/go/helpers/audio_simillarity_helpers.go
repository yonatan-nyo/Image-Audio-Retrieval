package helpers

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func LoadNotesArrayFromJSON(filePath string) ([]int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data struct {
		Data string `json:"data"`
	}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	stringNotes := strings.Fields(data.Data)

	notes := make([]int, len(stringNotes))
	for i, str := range stringNotes {
		note, err := strconv.Atoi(str)
		if err != nil {
			return nil, fmt.Errorf("error parsing note '%s': %v", str, err)
		}
		notes[i] = note
	}

	return notes, nil
}

func normalizePitch(notes []int) []float64 {
	var mean, stdDev float64
	for _, note := range notes {
		mean += float64(note)
	}
	mean /= float64(len(notes))

	for _, note := range notes {
		stdDev += math.Pow(float64(note)-mean, 2)
	}
	stdDev = math.Sqrt(stdDev / float64(len(notes)))

	normalized := make([]float64, len(notes))
	for i, note := range notes {
		normalized[i] = (float64(note) - mean) / stdDev
	}
	return normalized
}

func computeATB(notes []int) []float64 {
	atb := make([]float64, 128)
	for _, note := range notes {
		atb[note]++
	}

	// total := 0.0
	// for _, count := range atb {
	// 	total += count
	// }
	// for i := range atb {
	// 	atb[i] /= total
	// }
	return atb
}

func computeRTB(notes []int) []float64 {
	rtb := make([]float64, 255)
	for i := 1; i < len(notes); i++ {
		diff := notes[i] - notes[i-1] + 127
		rtb[diff]++
	}

	// total := 0.0
	// for _, count := range rtb {
	// 	total += count
	// }
	// for i := range rtb {
	// 	rtb[i] /= total
	// }
	return rtb
}

func computeFTB(notes []int) []float64 {
	ftb := make([]float64, 255)
	if len(notes) == 0 {
		return ftb
	}

	first := notes[0]
	for _, note := range notes {
		diff := note - first + 127
		ftb[diff]++
	}

	// total := 0.0
	// for _, count := range ftb {
	// 	total += count
	// }
	// for i := range ftb {
	// 	ftb[i] /= total
	// }
	return ftb
}

func cosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		panic("Vectors must have the same length")
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		normA += vec1[i] * vec1[i]
		normB += vec2[i] * vec2[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func CheckAudioSimilarity(hummingAudioPathMidi, songAudioPathMidi string) float64 {
	hummingNotes, err := LoadNotesArrayFromJSON(hummingAudioPathMidi)
	if err != nil {
		return -1
	}

	songNotes, err := LoadNotesArrayFromJSON(songAudioPathMidi)
	if err != nil {
		return -1
	}

	hummingATB := computeATB(hummingNotes)
	hummingRTB := computeRTB(hummingNotes)
	hummingFTB := computeFTB(hummingNotes)

	if len(songNotes) <= len(hummingNotes) {
		songATB := computeATB(songNotes)
		songRTB := computeRTB(songNotes)
		songFTB := computeFTB(songNotes)

		atbSim := cosineSimilarity(hummingATB, songATB)
		rtbSim := cosineSimilarity(hummingRTB, songRTB)
		ftbSim := cosineSimilarity(hummingFTB, songFTB)

		return 0.05*atbSim + 0.55*rtbSim + 0.40*ftbSim
	}

	windowSize := len(hummingNotes)

	songATB := computeATB(songNotes[0:windowSize])
	songRTB := computeRTB(songNotes[0:windowSize])
	songFTB := computeFTB(songNotes[0:windowSize])

	atbSim := cosineSimilarity(hummingATB, songATB)
	rtbSim := cosineSimilarity(hummingRTB, songRTB)
	ftbSim := cosineSimilarity(hummingFTB, songFTB)

	maxSimilarity := 0.05*atbSim + 0.55*rtbSim + 0.40*ftbSim

	for it := windowSize; it <= len(songNotes); it++ {

		songATB[songNotes[it-windowSize]]--
		songATB[songNotes[it]]++

		songRTB[songNotes[it-windowSize+1]-songNotes[it-windowSize]+127]--
		songRTB[songNotes[it]-songNotes[it-1]+127]++

		songFTB[songNotes[it-windowSize]-songNotes[0]+127]--
		songFTB[songNotes[it]-songNotes[0]+127]--

		atbSim := cosineSimilarity(hummingATB, songATB)
		rtbSim := cosineSimilarity(hummingRTB, songRTB)
		ftbSim := cosineSimilarity(hummingFTB, songFTB)

		similarity := 0.05*atbSim + 0.55*rtbSim + 0.40*ftbSim

		// Update maximum similarity score
		if similarity > maxSimilarity {
			maxSimilarity = similarity
		}
	}

	return maxSimilarity
}
