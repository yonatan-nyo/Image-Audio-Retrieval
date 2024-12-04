package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func ConvertToMidi(audioPath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(audioPath))
	if ext == ".mid" {
		log.Println("File is already a MIDI file:", audioPath)
		return audioPath, nil
	}

	// URL of the external FastAPI service
	apiURL := "http://127.0.0.1:8000/convert-to-midi/"

	// Prepare the request payload using json.Marshal to handle escaping automatically
	payload := map[string]string{"file_path": audioPath}
	requestBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal request body: %v\n", err)
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(requestBody)))
	if err != nil {
		log.Printf("Failed to create HTTP request: %v\n", err)
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send request to API: %v\n", err)
		return "", fmt.Errorf("failed to send request to API: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v\n", err)
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Print the response body to see its content
	log.Printf("Response body: %s\n", string(respBody))

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("API returned non-200 status code: %d\n", resp.StatusCode)
		return "", fmt.Errorf("API returned non-200 status code: %d", resp.StatusCode)
	}

	// Parse the response
	var response struct {
		FullPath string `json:"full_path"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Printf("Failed to parse API response: %v\n", err)
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}

	// Log the returned full path
	log.Printf("MIDI conversion successful, full path: %s\n", response.FullPath)
	return response.FullPath, nil
}
