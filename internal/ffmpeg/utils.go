package ffmpeg

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// listAudioDevices returns a list of available audio devices (input or output)
func ListAudioDevices() ([]string, error) {
	cmd := exec.Command("ffmpeg.exe", "-list_devices", "true", "-f", "dshow", "-i", "dummy")

	// ffmpeg writes device list to stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	var devices []string
	scanner := bufio.NewScanner(stderr)

	// Updated regex to match device lines - ffmpeg outputs device names in quotes
	deviceRegex := regexp.MustCompile(`"([^"]+)"\s+\(audio\)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Look for device names with the specified type
		matches := deviceRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			deviceName := strings.TrimSpace(matches[1])
			if deviceName != "" {
				devices = append(devices, deviceName)
			}
		}
	}

	// Wait for command to finish (it will error out, which is expected)
	err = cmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg process failed: %w", err)
	}

	return devices, nil
}

// ConcatenateAudioFiles combines multiple audio files using FFmpeg
func ConcatenateAudioFiles(inputFiles []string, outputFile string) error {
	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files provided")
	}

	// OPTIMIZE: Single file case - just copy instead of using FFmpeg
	if len(inputFiles) == 1 {
		return copyFile(inputFiles[0], outputFile)
	}

	// NOTE: Create a file list for FFmpeg concat demuxer
	listFile, err := os.CreateTemp("", "ffmpeg-concat-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create concat list file: %w", err)
	}
	defer os.Remove(listFile.Name()) // FIXME: Ensure cleanup

	// Write file list in FFmpeg concat format
	var content strings.Builder
	for _, file := range inputFiles {
		absPath, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", file, err)
		}
		// WARNING: Escape single quotes for FFmpeg safety
		escapedPath := strings.ReplaceAll(absPath, "'", "'\"'\"'")
		fmt.Fprintf(&content, "file '%s'\n", escapedPath)
	}

	err = os.WriteFile(listFile.Name(), []byte(content.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write concat list: %w", err)
	}

	// Use FFmpeg concat demuxer for efficient concatenation
	return runFFmpegConcat(listFile.Name(), outputFile)
}

// copyFile copies a single file when concatenation isn't needed
func copyFile(src, dst string) error {
	// NOTE: Simple file copy for single-file case
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// runFFmpegConcat executes FFmpeg to concatenate audio files
func runFFmpegConcat(listFile, outputFile string) error {
	cmd := exec.Command(
		"ffmpeg.exe",
		"-f", "concat", // Use concat demuxer
		"-safe", "0", // Allow unsafe file paths
		"-i", listFile, // Input file list
		"-c", "copy", // Copy streams without re-encoding
		"-y", // Overwrite output file
		outputFile,
	)

	// REVIEW: Consider capturing stderr for better error reporting
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("FFmpeg concatenation failed: %w", err)
	}

	return nil
}
