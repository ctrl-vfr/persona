// Package audio provides parallel TTS generation capabilities.
package audio

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/ctrl-vfr/persona/internal/openai"
	"github.com/ctrl-vfr/persona/internal/textsplit"
)

// AudioChunk represents generated audio with its order and potential error.
type AudioChunk struct {
	FilePath string
	Order    int
	Error    error
}

// TTSGenerator handles parallel text-to-speech generation.
type TTSGenerator struct {
	aiClient          *openai.OpenAI
	voiceInstructions string
	maxWorkers        int
}

// NewTTSGenerator creates a new TTS generator with specified configuration.
// maxWorkers is clamped to [1, 5] to avoid hammering the OpenAI API.
func NewTTSGenerator(aiClient *openai.OpenAI, voiceInstructions string, maxWorkers int) *TTSGenerator {
	if maxWorkers <= 0 || maxWorkers > 5 {
		maxWorkers = 3
	}
	return &TTSGenerator{
		aiClient:          aiClient,
		voiceInstructions: voiceInstructions,
		maxWorkers:        maxWorkers,
	}
}

// GenerateParallel generates audio for all text chunks in parallel.
// ctx is propagated to every OpenAI request so a single cancellation
// (Ctrl+C, command timeout) stops the whole batch.
func (tts *TTSGenerator) GenerateParallel(ctx context.Context, chunks []textsplit.TextChunk) ([]AudioChunk, error) {
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks to process")
	}

	workChan := make(chan textsplit.TextChunk, len(chunks))
	resultChan := make(chan AudioChunk, len(chunks))

	var wg sync.WaitGroup
	for i := 0; i < tts.maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range workChan {
				resultChan <- tts.processChunk(ctx, chunk)
			}
		}()
	}

	go func() {
		for _, chunk := range chunks {
			workChan <- chunk
		}
		close(workChan)
	}()

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

// processChunk processes a single text chunk into audio.
func (tts *TTSGenerator) processChunk(ctx context.Context, chunk textsplit.TextChunk) AudioChunk {
	audioBytes, err := tts.aiClient.GenerateAudio(ctx, chunk.Text, tts.voiceInstructions)
	if err != nil {
		return AudioChunk{
			Order: chunk.Order,
			Error: fmt.Errorf("failed to generate audio for chunk %d: %w", chunk.Order, err),
		}
	}

	tempFile, err := os.CreateTemp("", fmt.Sprintf("persona-chunk-%d-*.mp3", chunk.Order))
	if err != nil {
		return AudioChunk{
			Order: chunk.Order,
			Error: fmt.Errorf("failed to create temp file for chunk %d: %w", chunk.Order, err),
		}
	}
	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return AudioChunk{
			Order: chunk.Order,
			Error: fmt.Errorf("failed to close temp file for chunk %d: %w", chunk.Order, err),
		}
	}

	if err := os.WriteFile(tempPath, audioBytes, 0o600); err != nil {
		_ = os.Remove(tempPath)
		return AudioChunk{
			Order: chunk.Order,
			Error: fmt.Errorf("failed to write audio file for chunk %d: %w", chunk.Order, err),
		}
	}

	return AudioChunk{
		FilePath: tempPath,
		Order:    chunk.Order,
		Error:    nil,
	}
}

// CleanupAudioChunks removes temporary audio files. Errors are logged
// to stderr but never fail the caller.
func CleanupAudioChunks(chunks []AudioChunk) {
	for _, chunk := range chunks {
		if chunk.FilePath != "" {
			if err := os.Remove(chunk.FilePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to cleanup audio file %s: %v\n", chunk.FilePath, err)
			}
		}
	}
}

// FilterSuccessfulChunks separates successful chunks from failed ones.
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

// ExtractFilePaths extracts file paths from successful audio chunks.
func ExtractFilePaths(chunks []AudioChunk) []string {
	filePaths := make([]string, len(chunks))
	for i, chunk := range chunks {
		filePaths[i] = chunk.FilePath
	}
	return filePaths
}
