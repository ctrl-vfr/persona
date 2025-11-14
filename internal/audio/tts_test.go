package audio

import (
	"fmt"
	"testing"

	"github.com/ctrl-vfr/persona/internal/textsplit"
)

func TestFilterSuccessfulChunks(t *testing.T) {
	// TODO: Test de séparation des chunks réussis et échoués
	chunks := []AudioChunk{
		{Order: 0, FilePath: "chunk0.mp3", Error: nil},
		{Order: 1, FilePath: "", Error: fmt.Errorf("generation failed")},
		{Order: 2, FilePath: "chunk2.mp3", Error: nil},
		{Order: 3, FilePath: "", Error: fmt.Errorf("API timeout")},
		{Order: 4, FilePath: "chunk4.mp3", Error: nil},
	}

	successful, errors := FilterSuccessfulChunks(chunks)

	// REVIEW: Vérifier le nombre de chunks réussis
	expectedSuccessful := 3
	if len(successful) != expectedSuccessful {
		t.Errorf("Expected %d successful chunks, got %d", expectedSuccessful, len(successful))
	}

	// NOTE: Vérifier le nombre d'erreurs
	expectedErrors := 2
	if len(errors) != expectedErrors {
		t.Errorf("Expected %d errors, got %d", expectedErrors, len(errors))
	}

	// WARNING: Vérifier que les chunks réussis n'ont pas d'erreur
	for i, chunk := range successful {
		if chunk.Error != nil {
			t.Errorf("Successful chunk %d should not have error: %v", i, chunk.Error)
		}
		if chunk.FilePath == "" {
			t.Errorf("Successful chunk %d should have a file path", i)
		}
	}

	// FIXME: Vérifier le format des messages d'erreur
	for i, errMsg := range errors {
		if errMsg == "" {
			t.Errorf("Error message %d should not be empty", i)
		}
		t.Logf("Error %d: %s", i, errMsg)
	}
}

func TestExtractFilePaths(t *testing.T) {
	// NOTE: Test d'extraction des chemins de fichiers
	chunks := []AudioChunk{
		{Order: 0, FilePath: "/tmp/chunk0.mp3"},
		{Order: 1, FilePath: "/tmp/chunk1.mp3"},
		{Order: 2, FilePath: "/tmp/chunk2.mp3"},
	}

	filePaths := ExtractFilePaths(chunks)

	if len(filePaths) != len(chunks) {
		t.Errorf("Expected %d file paths, got %d", len(chunks), len(filePaths))
	}

	for i, path := range filePaths {
		expectedPath := chunks[i].FilePath
		if path != expectedPath {
			t.Errorf("File path %d: expected %s, got %s", i, expectedPath, path)
		}
	}

	// OPTIMIZE: Test avec chunks vides
	emptyChunks := []AudioChunk{}
	emptyPaths := ExtractFilePaths(emptyChunks)
	if len(emptyPaths) != 0 {
		t.Errorf("Expected 0 file paths for empty chunks, got %d", len(emptyPaths))
	}
}

func TestNewTTSGenerator(t *testing.T) {
	// REVIEW: Test de création du générateur TTS

	// TODO: Mock OpenAI client for testing
	// For now, we test the configuration logic

	tests := []struct {
		name            string
		maxWorkers      int
		expectedWorkers int
		desc            string
	}{
		{
			name:            "Default workers",
			maxWorkers:      0,
			expectedWorkers: 3,
			desc:            "Should default to 3 workers for invalid input",
		},
		{
			name:            "Negative workers",
			maxWorkers:      -1,
			expectedWorkers: 3,
			desc:            "Should default to 3 workers for negative input",
		},
		{
			name:            "Too many workers",
			maxWorkers:      10,
			expectedWorkers: 3,
			desc:            "Should limit to 3 workers to prevent API overload",
		},
		{
			name:            "Valid workers",
			maxWorkers:      2,
			expectedWorkers: 2,
			desc:            "Should use provided valid worker count",
		},
		{
			name:            "Max valid workers",
			maxWorkers:      5,
			expectedWorkers: 5,
			desc:            "Should allow up to 5 workers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Test: %s", tt.desc)

			// NOTE: Pass nil client for configuration testing
			tts := NewTTSGenerator(nil, "test instructions", tt.maxWorkers)

			if tts.maxWorkers != tt.expectedWorkers {
				t.Errorf("Expected %d workers, got %d", tt.expectedWorkers, tts.maxWorkers)
			}

			if tts.voiceInstructions != "test instructions" {
				t.Errorf("Expected voice instructions to be preserved")
			}
		})
	}
}

// TestGenerateParallelWithMockData tests the parallel generation logic without actual API calls
func TestGenerateParallelWithMockData(t *testing.T) {
	// WARNING: This test requires careful setup to avoid actual API calls

	// Test with empty chunks
	tts := NewTTSGenerator(nil, "test", 3)

	emptyChunks := []textsplit.TextChunk{}
	_, err := tts.GenerateParallel(emptyChunks)

	if err == nil {
		t.Error("Expected error for empty chunks")
	}

	expectedError := "no chunks to process"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

