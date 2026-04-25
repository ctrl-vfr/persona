package ffmpeg

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// ListAudioDevices returns the audio input devices ffmpeg can see on
// the current host. Output is parsed from ffmpeg's stderr because that
// is where it prints device enumeration; the command itself always
// exits non-zero (it's a "no actual input" call).
//
// The parser is platform-specific because each input format prints
// devices in its own format:
//   - dshow (Windows):       [dshow @ ...]  "Microphone (Realtek)"
//   - avfoundation (macOS):  [AVFoundation indev @ ...] [0] MacBook Pro Microphone
//   - alsa (Linux):          ffmpeg -list_devices is not supported.
func ListAudioDevices() ([]string, error) {
	cmd := exec.Command(ffmpegBinary(), "-list_devices", "true", "-f", inputFormat(), "-i", "")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	devices, stderrText := parseDevices(stderr, runtime.GOOS)

	// Always Wait to reap the process. ffmpeg exits non-zero here.
	if err := cmd.Wait(); err != nil && len(devices) == 0 {
		if strings.Contains(stderrText, "Unknown input format") || strings.Contains(stderrText, "No such file or directory") {
			return nil, fmt.Errorf("ffmpeg device listing failed: %w\nFFmpeg stderr output:\n%s", err, stderrText)
		}
	}

	return devices, nil
}

// parseDevices reads ffmpeg's stderr stream and extracts audio input
// device names according to the host OS output format. Returns the
// device list and the raw stderr (for error reporting upstream).
func parseDevices(r io.Reader, goos string) ([]string, string) {
	var (
		devices   []string
		raw       strings.Builder
		scanner   = bufio.NewScanner(r)
		dshowRe   = regexp.MustCompile(`"([^"]+)"`)
		avfoundRe = regexp.MustCompile(`\[\d+\]\s+(.+)$`)
		// avfoundation prints "AVFoundation video devices:" then
		// "AVFoundation audio devices:". We only collect lines after
		// the audio header.
		inAudioSection bool
	)

	for scanner.Scan() {
		line := scanner.Text()
		raw.WriteString(line)
		raw.WriteString("\n")

		switch goos {
		case "windows":
			if strings.Contains(line, "(audio)") {
				if m := dshowRe.FindStringSubmatch(line); len(m) > 1 {
					if name := strings.TrimSpace(m[1]); name != "" {
						devices = append(devices, name)
					}
				}
			}
		case "darwin":
			lower := strings.ToLower(line)
			if strings.Contains(lower, "avfoundation audio devices") {
				inAudioSection = true
				continue
			}
			if strings.Contains(lower, "avfoundation video devices") {
				inAudioSection = false
				continue
			}
			if !inAudioSection {
				continue
			}
			if m := avfoundRe.FindStringSubmatch(line); len(m) > 1 {
				if name := strings.TrimSpace(m[1]); name != "" {
					devices = append(devices, name)
				}
			}
		}
	}
	return devices, raw.String()
}

// ConcatenateAudioFiles combines multiple audio files using FFmpeg's
// concat demuxer.
func ConcatenateAudioFiles(inputFiles []string, outputFile string) error {
	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files provided")
	}
	if len(inputFiles) == 1 {
		return copyFile(inputFiles[0], outputFile)
	}

	listFile, err := os.CreateTemp("", "ffmpeg-concat-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create concat list file: %w", err)
	}
	defer func() { _ = os.Remove(listFile.Name()) }()

	var content strings.Builder
	for _, file := range inputFiles {
		absPath, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", file, err)
		}
		// Escape single quotes for FFmpeg's concat list format.
		escapedPath := strings.ReplaceAll(absPath, "'", "'\"'\"'")
		fmt.Fprintf(&content, "file '%s'\n", escapedPath)
	}

	if err := os.WriteFile(listFile.Name(), []byte(content.String()), 0o600); err != nil {
		return fmt.Errorf("failed to write concat list: %w", err)
	}

	return runFFmpegConcat(listFile.Name(), outputFile)
}

// copyFile copies a single file when concatenation isn't needed.
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = source.Close() }()

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = destination.Close() }()

	if _, err := io.Copy(destination, source); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}
	return nil
}

// runFFmpegConcat executes FFmpeg to concatenate audio files.
func runFFmpegConcat(listFile, outputFile string) error {
	cmd := exec.Command(
		ffmpegBinary(),
		"-f", "concat",
		"-safe", "0",
		"-i", listFile,
		"-c", "copy",
		"-y",
		outputFile,
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg concatenation failed: %w", err)
	}
	return nil
}
