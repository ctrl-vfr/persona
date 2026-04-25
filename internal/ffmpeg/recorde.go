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

// Record starts recording audio and stops when silence is detected.
// The returned path is the WAV file written by ffmpeg; the caller is
// responsible for removing it.
func (f *FFmpeg) Record() (string, error) {
	if f.Recorder.Input == "" {
		return "", fmt.Errorf("no input device configured. Use 'persona config set-input-device <device>' to configure")
	}

	tempFile, err := os.CreateTemp("", "recording-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	// Close immediately: ffmpeg opens its own handle to write to this
	// path. Holding the descriptor open here would leak one fd per call.
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	cmd := exec.Command(
		ffmpegBinary(),
		"-f", inputFormat(),
		"-y",
		"-i", fmt.Sprintf("audio=%s", f.Recorder.Input),
		"-af", fmt.Sprintf("silencedetect=n=%ddB:d=%d", f.Recorder.SilenceThreshold, f.Recorder.SilenceDuration),
		tempPath,
	)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	var stderrOutput strings.Builder
	silenceDetected := false

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		stderrOutput.WriteString(line)
		stderrOutput.WriteString("\n")

		if strings.Contains(line, "silencedetect @") && strings.Contains(line, "silence_start:") {
			silenceDetected = true
			if err := cmd.Process.Kill(); err != nil {
				stderrOutput.WriteString(fmt.Sprintf("warn: kill ffmpeg: %v\n", err))
			}
			break
		}
	}

	// Always Wait so we don't leak a zombie process. If we killed it
	// after silence detection the exit error is expected and ignored.
	waitErr := cmd.Wait()
	if waitErr != nil && !silenceDetected {
		_ = os.Remove(tempPath)
		stderrText := stderrOutput.String()
		if stderrText != "" {
			return "", fmt.Errorf("ffmpeg process failed: %w\nFFmpeg stderr output:\n%s", waitErr, stderrText)
		}
		return "", fmt.Errorf("ffmpeg process failed: %w", waitErr)
	}

	// Sanity: an empty file means ffmpeg produced no audio at all.
	fileInfo, err := os.Stat(tempPath)
	if err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("failed to verify recording file: %w", err)
	}
	if fileInfo.Size() == 0 {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("recording file is empty - no audio was captured")
	}

	return tempPath, nil
}
