package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGetVersionInfo(t *testing.T) {
	info := getVersionInfo()

	// Check that all fields are present
	if info.Version == "" {
		t.Error("Version should not be empty")
	}
	if info.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
	if info.Platform == "" {
		t.Error("Platform should not be empty")
	}

	// Check that Go version starts with "go"
	if !strings.HasPrefix(info.GoVersion, "go") {
		t.Errorf("GoVersion should start with 'go', got '%s'", info.GoVersion)
	}

	// Check that platform contains "/"
	if !strings.Contains(info.Platform, "/") {
		t.Errorf("Platform should contain '/', got '%s'", info.Platform)
	}
}

func TestGetVersionString(t *testing.T) {
	// Test with development version
	originalVersion := version
	version = "dev"
	defer func() { version = originalVersion }()

	versionString := getVersionString()

	if !strings.Contains(versionString, "persona") {
		t.Error("Version string should contain 'persona'")
	}
	if !strings.Contains(versionString, "dev") {
		t.Error("Version string should contain 'dev' for development version")
	}

	// Test with release version
	version = "v1.0.0"
	versionString = getVersionString()

	if !strings.Contains(versionString, "v1.0.0") {
		t.Error("Version string should contain version number")
	}
	if !strings.Contains(versionString, "persona") {
		t.Error("Version string should contain 'persona'")
	}
}

func TestVersionInfoJSON(t *testing.T) {
	info := getVersionInfo()

	// Test JSON marshalling
	jsonData, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal version info to JSON: %v", err)
	}

	// Test JSON unmarshalling
	var unmarshaledInfo VersionInfo
	err = json.Unmarshal(jsonData, &unmarshaledInfo)
	if err != nil {
		t.Fatalf("Failed to unmarshal version info from JSON: %v", err)
	}

	// Compare fields
	if unmarshaledInfo.Version != info.Version {
		t.Errorf("Version mismatch after JSON roundtrip: expected '%s', got '%s'", info.Version, unmarshaledInfo.Version)
	}
	if unmarshaledInfo.GoVersion != info.GoVersion {
		t.Errorf("GoVersion mismatch after JSON roundtrip: expected '%s', got '%s'", info.GoVersion, unmarshaledInfo.GoVersion)
	}
	if unmarshaledInfo.Platform != info.Platform {
		t.Errorf("Platform mismatch after JSON roundtrip: expected '%s', got '%s'", info.Platform, unmarshaledInfo.Platform)
	}
}

func TestVersionVariables(t *testing.T) {
	// Save original values
	originalVersion := version
	originalBuildTime := buildTime
	originalGitCommit := gitCommit
	originalGitBranch := gitBranch

	// Test setting version variables
	version = "v2.0.0"
	buildTime = "2024-01-01T12:00:00Z"
	gitCommit = "abc123def"
	gitBranch = "main"

	defer func() {
		// Restore original values
		version = originalVersion
		buildTime = originalBuildTime
		gitCommit = originalGitCommit
		gitBranch = originalGitBranch
	}()

	info := getVersionInfo()

	if info.Version != "v2.0.0" {
		t.Errorf("Expected version 'v2.0.0', got '%s'", info.Version)
	}
	if info.BuildTime != "2024-01-01T12:00:00Z" {
		t.Errorf("Expected build time '2024-01-01T12:00:00Z', got '%s'", info.BuildTime)
	}
	if info.GitCommit != "abc123def" {
		t.Errorf("Expected git commit 'abc123def', got '%s'", info.GitCommit)
	}
	if info.GitBranch != "main" {
		t.Errorf("Expected git branch 'main', got '%s'", info.GitBranch)
	}
}

func TestVersionStringFormats(t *testing.T) {
	// Save original values
	originalVersion := version
	originalGitCommit := gitCommit

	defer func() {
		version = originalVersion
		gitCommit = originalGitCommit
	}()

	testCases := []struct {
		name             string
		version          string
		gitCommit        string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:          "Development version",
			version:       "dev",
			gitCommit:     "abc123",
			shouldContain: []string{"persona", "dev", "abc123"},
		},
		{
			name:             "Release version",
			version:          "v1.2.3",
			gitCommit:        "def456",
			shouldContain:    []string{"persona", "v1.2.3"},
			shouldNotContain: []string{"def456"}, // commit not shown in release
		},
		{
			name:          "Beta version",
			version:       "v2.0.0-beta",
			gitCommit:     "ghi789",
			shouldContain: []string{"persona", "v2.0.0-beta"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			version = tc.version
			gitCommit = tc.gitCommit

			versionString := getVersionString()

			for _, expected := range tc.shouldContain {
				if !strings.Contains(versionString, expected) {
					t.Errorf("Version string should contain '%s', got '%s'", expected, versionString)
				}
			}

			for _, notExpected := range tc.shouldNotContain {
				if strings.Contains(versionString, notExpected) {
					t.Errorf("Version string should not contain '%s', got '%s'", notExpected, versionString)
				}
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkGetVersionInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getVersionInfo()
	}
}

func BenchmarkGetVersionString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getVersionString()
	}
}

func BenchmarkVersionInfoJSON(b *testing.B) {
	info := getVersionInfo()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(info)
	}
}
