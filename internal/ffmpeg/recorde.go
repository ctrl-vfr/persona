// Package speak provides a method to play a file using the speaker.
package ffmpeg

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Recorder holds configuration for audio recording
type Recorder struct {
	Input            string
	SilenceThreshold int
	SilenceDuration  int
}

// Player holds configuration for audio playback
type Player struct {
	Output string
}

// FFmpeg manages ffmpeg operations
type FFmpeg struct {
	Recorder Recorder
}

// New creates a new FFmpeg instance with default values for optional parameters
func New(input string, silenceThreshold int, silenceDuration int) *FFmpeg {
	// Set default values if negative values are provided
	if silenceThreshold == 0 {
		silenceThreshold = -50
	}

	if silenceDuration == 0 {
		silenceDuration = 2
	}

	return &FFmpeg{
		Recorder: Recorder{
			Input:            input,
			SilenceThreshold: silenceThreshold,
			SilenceDuration:  silenceDuration,
		},
	}
}

// Record starts recording audio and stops when silence is detected
func (f *FFmpeg) Record() (string, error) {
	tempFile, err := os.CreateTemp("", "recording-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	// Close the file handle immediately so FFmpeg can use it
	tempFile.Close()
	// Build ffmpeg command
	cmd := exec.Command(
		"ffmpeg.exe",
		"-f", "dshow",
		"-y",
		"-i", fmt.Sprintf("audio=%s", f.Recorder.Input),
		"-af", fmt.Sprintf("silencedetect=n=%ddB:d=%d", f.Recorder.SilenceThreshold, f.Recorder.SilenceDuration),
		tempFile.Name(),
	)

	// Get stderr pipe to monitor silence detection and capture errors
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Capture all stderr output for error logging
	var stderrOutput strings.Builder
	silenceDetected := false

	// Monitor stderr for silence detection and capture all output
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()

		// Always capture stderr output for error reporting
		stderrOutput.WriteString(line)
		stderrOutput.WriteString("\n")

		// Stop recording when silence is detected
		if strings.Contains(line, "silencedetect @") && strings.Contains(line, "silence_start:") {
			silenceDetected = true
			if err := cmd.Process.Kill(); err != nil {
				fmt.Printf("Warning: failed to kill ffmpeg process: %v\n", err)
			}
			break
		}
	}

	// Wait for command to complete
	err = cmd.Wait()
	if err != nil {
		// If silence was detected and process was killed by us, that's expected behavior
		if silenceDetected {
			// This is normal - we killed the process after detecting silence
			return tempFile.Name(), nil
		}

		// Real error - include stderr output for debugging
		stderrText := stderrOutput.String()
		if stderrText != "" {
			return "", fmt.Errorf("ffmpeg process failed: %w\nFFmpeg stderr output:\n%s", err, stderrText)
		}
		return "", fmt.Errorf("ffmpeg process failed: %w", err)
	}

	return tempFile.Name(), nil
}
