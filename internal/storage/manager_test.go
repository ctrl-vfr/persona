package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestManager returns a Manager rooted at a temporary directory and
// ensures InitializeStructure ran (so ~/.persona equivalent exists in
// the temp dir).
func newTestManager(t *testing.T) *Manager {
	t.Helper()
	m := &Manager{BasePath: t.TempDir()}
	if err := m.InitializeStructure(); err != nil {
		t.Fatalf("init structure: %v", err)
	}
	return m
}

func TestValidatePersonaName(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		wantOK bool
	}{
		{"simple", "marceline", true},
		{"with-dash", "my-persona", true},
		{"with_underscore", "my_persona_2", true},
		{"empty", "", false},
		{"path-traversal", "../etc/passwd", false},
		{"slash", "foo/bar", false},
		{"backslash", `foo\bar`, false},
		{"dot", ".", false},
		{"double-dot", "..", false},
		{"null-byte", "foo\x00bar", false},
		{"too-long", strings.Repeat("a", 65), false},
		{"max-length", strings.Repeat("a", 64), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePersonaName(tc.input)
			if (err == nil) != tc.wantOK {
				t.Errorf("ValidatePersonaName(%q): wantOK=%v, err=%v", tc.input, tc.wantOK, err)
			}
		})
	}
}

func TestManager_GetPersona_RejectsTraversal(t *testing.T) {
	m := newTestManager(t)
	_, err := m.GetPersona("../etc/passwd")
	if err == nil {
		t.Fatal("expected error for path traversal, got nil")
	}
}

func TestManager_InitializeStructure_CreatesBuiltins(t *testing.T) {
	m := newTestManager(t)
	for _, b := range GetBuiltinPersonas() {
		if !m.PersonaExists(b.Name) {
			t.Errorf("built-in persona %q missing after init", b.Name)
		}
	}
	// config.yaml exists with restricted perms (only on Unix; Windows
	// does not honor 0600 the same way).
	if _, err := os.Stat(m.GetConfigPath()); err != nil {
		t.Errorf("config file missing: %v", err)
	}
}

func TestManager_PersonaPath(t *testing.T) {
	m := &Manager{BasePath: "/tmp/persona-test"}
	personaPath, historyPath := m.GetPersonaPath("kevin")
	if !filepath.IsAbs(personaPath) || !filepath.IsAbs(historyPath) {
		t.Errorf("expected absolute paths, got %q / %q", personaPath, historyPath)
	}
	if !strings.HasSuffix(personaPath, filepath.Join("personas", "kevin", "persona.yaml")) {
		t.Errorf("unexpected persona path: %s", personaPath)
	}

	// Invalid name returns empty paths so any subsequent os.* call fails.
	bp, bh := m.GetPersonaPath("../escape")
	if bp != "" || bh != "" {
		t.Errorf("expected empty paths for invalid name, got %q %q", bp, bh)
	}
}
