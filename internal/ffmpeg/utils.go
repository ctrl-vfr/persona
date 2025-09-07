package ffmpeg

import (
	"bufio"
	"fmt"
	"os/exec"
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
	var stderrOutput strings.Builder
	scanner := bufio.NewScanner(stderr)

	// Updated regex to match device lines - ffmpeg outputs device names in quotes
	deviceRegex := regexp.MustCompile(`"([^"]+)"\s+\(audio\)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Capture all stderr output for error reporting
		stderrOutput.WriteString(line)
		stderrOutput.WriteString("\n")

		// Look for device names with the specified type
		matches := deviceRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			deviceName := strings.TrimSpace(matches[1])
			if deviceName != "" {
				devices = append(devices, deviceName)
			}
		}
	}

	// Wait for command to finish (it will error out, which is expected for this command)
	err = cmd.Wait()
	if err != nil {
		// For device listing, FFmpeg always returns an error code, but that's expected
		// Only return error if we couldn't parse any devices and there's indication of a real problem
		if len(devices) == 0 {
			stderrText := stderrOutput.String()
			if strings.Contains(stderrText, "Unknown input format") || strings.Contains(stderrText, "No such file or directory") {
				return nil, fmt.Errorf("ffmpeg device listing failed: %w\nFFmpeg stderr output:\n%s", err, stderrText)
			}
		}
	}

	return devices, nil
}
