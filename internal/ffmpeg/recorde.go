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

	if f.Recorder.Input == "" {
		return "", fmt.Errorf("no input device configured. Use 'persona config set-input-device <device>' to configure")
	}

	cmd := exec.Command(
		"ffmpeg.exe",
		"-f", "dshow",
		"-y",
		"-i", fmt.Sprintf("audio=%s", f.Recorder.Input),
		"-af", fmt.Sprintf("silencedetect=n=%ddB:d=%d", f.Recorder.SilenceThreshold, f.Recorder.SilenceDuration),
		tempFile.Name(),
	)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	processWasKilled := false
	var stderrOutput strings.Builder

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		stderrOutput.WriteString(line + "\n")

		if strings.Contains(line, "silencedetect @") && strings.Contains(line, "silence_start:") {
			if err := cmd.Process.Kill(); err != nil {
				fmt.Printf("Warning: failed to kill ffmpeg process: %v\n", err)
			} else {
				processWasKilled = true
			}
			break
		}
	}

	err = cmd.Wait()

	if err != nil && !processWasKilled {
		return "", fmt.Errorf("ffmpeg process failed: %w\nFFmpeg output:\n%s", err, stderrOutput.String())
	}

	fileInfo, err := os.Stat(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to verify recording file: %w", err)
	}

	if fileInfo.Size() == 0 {
		return "", fmt.Errorf("recording file is empty - no audio was captured")
	}

	return tempFile.Name(), nil
}
