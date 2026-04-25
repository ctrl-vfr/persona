package ffmpeg

import "runtime"

// ffmpegBinary returns the platform-appropriate ffmpeg executable name.
// We let exec.Command resolve it from PATH rather than hardcoding paths.
func ffmpegBinary() string {
	if runtime.GOOS == "windows" {
		return "ffmpeg.exe"
	}
	return "ffmpeg"
}

// inputFormat returns the ffmpeg input format flag for capturing audio
// devices on the host OS. dshow on Windows, avfoundation on macOS,
// alsa on Linux. Without this dispatch ffmpeg refuses to open the
// device on non-Windows hosts.
func inputFormat() string {
	switch runtime.GOOS {
	case "windows":
		return "dshow"
	case "darwin":
		return "avfoundation"
	default:
		return "alsa"
	}
}
