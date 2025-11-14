// Package audio provides parallel TTS generation capabilities
package audio

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ctrl-vfr/persona/internal/openai"
	"github.com/ctrl-vfr/persona/internal/textsplit"
)

// AudioChunk represents generated audio with its order and potential error
type AudioChunk struct {
	FilePath string
	Order    int
	Error    error
}

// TTSGenerator handles parallel text-to-speech generation
type TTSGenerator struct {
	aiClient          *openai.OpenAI
	voiceInstructions string
	maxWorkers        int
}

// NewTTSGenerator creates a new TTS generator with specified configuration
func NewTTSGenerator(aiClient *openai.OpenAI, voiceInstructions string, maxWorkers int) *TTSGenerator {
	// WARNING: Limit concurrent workers to prevent overwhelming the API
	if maxWorkers <= 0 || maxWorkers > 5 {
		maxWorkers = 3 // REVIEW: Conservative default, could be configurable
	}

	return &TTSGenerator{
		aiClient:          aiClient,
		voiceInstructions: voiceInstructions,
		maxWorkers:        maxWorkers,
	}
}

// GenerateParallel generates audio for all text chunks in parallel
func (tts *TTSGenerator) GenerateParallel(chunks []textsplit.TextChunk) ([]AudioChunk, error) {
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks to process")
	}

	// NOTE: Use channels for work distribution
	workChan := make(chan textsplit.TextChunk, len(chunks))
	resultChan := make(chan AudioChunk, len(chunks))

	var wg sync.WaitGroup

	// TODO: Add context support for cancellation
	// Start worker goroutines
	for i := 0; i < tts.maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for chunk := range workChan {
				audioChunk := tts.processChunk(chunk)
				resultChan <- audioChunk
			}
		}(i)
	}

	// Send chunks to workers
	go func() {
		for _, chunk := range chunks {
			workChan <- chunk
		}
		close(workChan)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	results := make([]AudioChunk, 0, len(chunks))
	for audioChunk := range resultChan {
		results = append(results, audioChunk)
	}

	return results, nil
}

// processChunk processes a single text chunk into audio
func (tts *TTSGenerator) processChunk(chunk textsplit.TextChunk) AudioChunk {
	// NOTE: Generate audio for this specific chunk
	audioResponseData, err := tts.aiClient.GenerateAudio(chunk.Text, tts.voiceInstructions)
	if err != nil {
		return AudioChunk{
			Order: chunk.Order,
			Error: fmt.Errorf("failed to generate audio for chunk %d: %w", chunk.Order, err),
		}
	}

	// Read audio data from response
	audioBytes, err := io.ReadAll(audioResponseData)
	if err != nil {
		return AudioChunk{
			Order: chunk.Order,
			Error: fmt.Errorf("failed to read audio data for chunk %d: %w", chunk.Order, err),
		}
	}

	// TODO: Make temp file prefix configurable
	// Create temporary file for this chunk
	tempFile, err := os.CreateTemp("", fmt.Sprintf("persona-chunk-%d-*.mp3", chunk.Order))
	if err != nil {
		return AudioChunk{
			Order: chunk.Order,
			Error: fmt.Errorf("failed to create temp file for chunk %d: %w", chunk.Order, err),
		}
	}

	// Write audio data to file
	err = os.WriteFile(tempFile.Name(), audioBytes, 0644)
	if err != nil {
		os.Remove(tempFile.Name())
		return AudioChunk{
			Order: chunk.Order,
			Error: fmt.Errorf("failed to write audio file for chunk %d: %w", chunk.Order, err),
		}
	}

	return AudioChunk{
		FilePath: tempFile.Name(),
		Order:    chunk.Order,
		Error:    nil,
	}
}

// CleanupAudioChunks removes temporary audio files
func CleanupAudioChunks(chunks []AudioChunk) {
	// NOTE: Clean up temporary files to prevent disk space issues
	for _, chunk := range chunks {
		if chunk.FilePath != "" {
			err := os.Remove(chunk.FilePath)
			if err != nil {
				// WARNING: Log but don't fail on cleanup errors
				fmt.Printf("Warning: failed to cleanup audio file %s: %v\n", chunk.FilePath, err)
			}
		}
	}
}

// FilterSuccessfulChunks separates successful chunks from failed ones
func FilterSuccessfulChunks(chunks []AudioChunk) (successful []AudioChunk, errors []string) {
	for _, chunk := range chunks {
		if chunk.Error != nil {
			errors = append(errors, fmt.Sprintf("Chunk %d: %s", chunk.Order, chunk.Error.Error()))
		} else {
			successful = append(successful, chunk)
		}
	}
	return successful, errors
}

// ExtractFilePaths extracts file paths from successful audio chunks
func ExtractFilePaths(chunks []AudioChunk) []string {
	filePaths := make([]string, len(chunks))
	for i, chunk := range chunks {
		filePaths[i] = chunk.FilePath
	}
	return filePaths
}

