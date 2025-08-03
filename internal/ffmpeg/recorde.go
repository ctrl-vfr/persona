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
	// Build ffmpeg command
	cmd := exec.Command(
		"ffmpeg.exe",
		"-f", "dshow",
		"-y",
		"-i", fmt.Sprintf("audio=%s", f.Recorder.Input),
		"-af", fmt.Sprintf("silencedetect=n=%ddB:d=%d", f.Recorder.SilenceThreshold, f.Recorder.SilenceDuration),
		tempFile.Name(),
	)

	// Get stderr pipe to monitor silence detection
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Monitor stderr for silence detection
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		// Stop recording when silence is detected
		if strings.Contains(line, "silencedetect @") && strings.Contains(line, "silence_start:") {
			if err := cmd.Process.Kill(); err != nil {
				fmt.Printf("Warning: failed to kill ffmpeg process: %v\n", err)
			}
			break
		}
	}

	// Wait for command to complete
	err = cmd.Wait()
	if err != nil {
		return "", fmt.Errorf("ffmpeg process failed: %w", err)
	}

	return tempFile.Name(), nil
}
