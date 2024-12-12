package helpers

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"

	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/midimessage/realtime"
	"gitlab.com/gomidi/midi/midireader"
)

// Helper function to read a MIDI file and extract note data
func extractMidiNotes(filePath string) ([]int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening MIDI file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	var notes []int
	rthandler := func(m realtime.Message) {
		//do nothing
	}

	rd := midireader.New(file, rthandler)

	var m midi.Message
	var errRead error

	for {
		m, errRead = rd.Read()
		if errRead != nil {
			if errRead == io.EOF {
				break
			}
			log.Fatalf("Error reading MIDI file: %v\n", errRead)
			return nil, errRead
		}

		switch v := m.(type) {
		case channel.NoteOn:
			notes = append(notes, int(v.Key()))
		}
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
	hummingNotes, err := extractMidiNotes(hummingAudioPathMidi)
	if err != nil {
		log.Fatalf("Failed to extract notes from humming MIDI file: %v", err)
		return 0
	}

	songNotes, err := extractMidiNotes(songAudioPathMidi)
	if err != nil {
		log.Fatalf("Failed to extract notes from song MIDI file: %v", err)
		return 0
	}

	hummingATB := computeATB(hummingNotes)
	hummingRTB := computeRTB(hummingNotes)
	hummingFTB := computeFTB(hummingNotes)

	songATB := computeATB(songNotes)
	songRTB := computeRTB(songNotes)
	songFTB := computeFTB(songNotes)

	atbSim := cosineSimilarity(hummingATB, songATB)
	rtbSim := cosineSimilarity(hummingRTB, songRTB)
	ftbSim := cosineSimilarity(hummingFTB, songFTB)

	for i := 0; i < len(songATB); i++ {
		fmt.Printf("humming[%d]: %.2f, song[%d]: %.2f\n", i, hummingATB[i], i, songATB[i])
	}

	overallSimilarity := (0.05*atbSim + 0.55*rtbSim + 0.40*ftbSim)

	return overallSimilarity
}
