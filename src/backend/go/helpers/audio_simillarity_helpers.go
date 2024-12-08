package helpers

import (
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

// Normalize the extracted note sequence based on pitch (to a fixed range)
func normalizeNotes(notes []int) ([]float64, float64, float64) {
	var minNote, maxNote int
	for _, note := range notes {
		if note < minNote {
			minNote = note
		}
		if note > maxNote {
			maxNote = note
		}
	}

	// Normalize the notes to the range [0, 1]
	normalized := make([]float64, len(notes))
	for i, note := range notes {
		normalized[i] = float64(note-minNote) / float64(maxNote-minNote)
	}

	return normalized, float64(minNote), float64(maxNote)
}

// Calculate cosine similarity between two normalized note sequences
func calculateCosineSimilarity(notes1, notes2 []float64) float64 {
	// Ensure both sequences are the same length by padding or truncating
	minLength := len(notes1)
	if len(notes2) < minLength {
		minLength = len(notes2)
	}

	// Optionally, you can pad the shorter array with zeroes if desired
	// For now, we'll truncate the longer array to the minimum length
	notes1 = notes1[:minLength]
	notes2 = notes2[:minLength]

	var dotProduct, norm1, norm2 float64

	for i := range notes1 {
		dotProduct += notes1[i] * notes2[i]
		norm1 += notes1[i] * notes1[i]
		norm2 += notes2[i] * notes2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// Main function to compare humming and song audio and calculate similarity
func CheckAudioSimilarity(hummingAudioPathMidi, songAudioPathMidi string) float64 {
	log.Println("Checking audio similarity...")

	// Extract notes from the humming MIDI file
	hummingNotes, err := extractMidiNotes(hummingAudioPathMidi)
	if err != nil {
		log.Fatalf("Error extracting humming notes: %v\n", err)
		return 0
	}

	// Extract notes from the song MIDI file
	songNotes, err := extractMidiNotes(songAudioPathMidi)
	if err != nil {
		log.Fatalf("Error extracting song notes: %v\n", err)
		return 0
	}

	// Normalize the notes for both the humming and song
	hummingNormalized, _, _ := normalizeNotes(hummingNotes)
	songNormalized, _, _ := normalizeNotes(songNotes)

	// Calculate the cosine similarity between the two normalized note sequences
	similarity := calculateCosineSimilarity(hummingNormalized, songNormalized)

	// Log the calculated similarity score

	log.Printf("%s %s Calculated similarity score: %f\n", hummingAudioPathMidi, songAudioPathMidi, similarity)

	// Return the similarity score
	return similarity
}
