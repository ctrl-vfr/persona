package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// peachRGB / skyRGB are the canonical Catppuccin Mocha values for the
// two semantic accents, expressed as the truecolor ANSI substring
// (R;G;B) lipgloss emits when forced to TrueColor profile.
const (
	peachRGB = "250;179;135" // #fab387
	skyRGB   = "137;220;235" // #89dceb
	redRGB   = "243;139;168" // #f38ba8
	greenRGB = "166;227;161" // #a6e3a1
)

func init() {
	// Hors TTY (`go test`), lipgloss sélectionne par défaut un profil
	// qui dégrade les couleurs en ANSI 16. On force TrueColor pour que
	// les tests puissent vérifier les `R;G;B` émis.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func TestChevronPill_WrapsContent(t *testing.T) {
	out := chevronPill(colPeach, colBase, true, "hello")
	if !strings.Contains(out, "hello") {
		t.Errorf("expected pill to contain text, got %q", out)
	}
	if !strings.Contains(out, chevronLeft) {
		t.Errorf("expected pill to contain left chevron")
	}
	if !strings.Contains(out, chevronRight) {
		t.Errorf("expected pill to contain right chevron")
	}
}

func TestChevronPill_EmptyContent(t *testing.T) {
	// Empty content still renders the chevrons + a space-padded body.
	out := chevronPill(colPeach, colBase, false, "")
	if out == "" {
		t.Error("chevronPill with empty content should still produce wrapped chevrons")
	}
}

func TestAccentPill_PeachAndIcon(t *testing.T) {
	out := accentPill(iconRecord, "rec")
	if !strings.Contains(strings.ToLower(out), peachRGB) {
		t.Errorf("expected peach ANSI %q in output, got %q", peachRGB, out)
	}
	if !strings.Contains(out, iconRecord) {
		t.Errorf("expected icon %q in output", iconRecord)
	}
	if !strings.Contains(out, "rec") {
		t.Errorf("expected text in output")
	}
}

func TestAccentAltPill_Sky(t *testing.T) {
	out := accentAltPill(iconBrand, "persona")
	if !strings.Contains(strings.ToLower(out), skyRGB) {
		t.Errorf("expected sky ANSI %q in output, got %q", skyRGB, out)
	}
}

func TestErrorPill_Red(t *testing.T) {
	out := errorPill(iconError, "boom")
	if !strings.Contains(strings.ToLower(out), redRGB) {
		t.Errorf("expected red ANSI %q in output, got %q", redRGB, out)
	}
}

func TestSuccessPill_Green(t *testing.T) {
	out := successPill(iconSuccess, "done")
	if !strings.Contains(strings.ToLower(out), greenRGB) {
		t.Errorf("expected green ANSI %q in output, got %q", greenRGB, out)
	}
}

func TestRenderError_UsesPill(t *testing.T) {
	out := RenderError("network down")
	if !strings.Contains(strings.ToLower(out), redRGB) {
		t.Errorf("RenderError should produce a red pill, got %q", out)
	}
	if !strings.Contains(out, "network down") {
		t.Errorf("RenderError should preserve message")
	}
}

func TestRenderInfo_UsesSky(t *testing.T) {
	out := RenderInfo("hello")
	if !strings.Contains(strings.ToLower(out), skyRGB) {
		t.Errorf("RenderInfo should be sky-tinted, got %q", out)
	}
}

func TestPillContent_OmitsLeadingSpaceWhenNoIcon(t *testing.T) {
	if got := pillContent("", "x"); got != "x" {
		t.Errorf("expected %q, got %q", "x", got)
	}
	// Two-space gap is intentional: Nerd Font glyphs often render
	// narrower than a cell, and a single space looks cramped.
	if got := pillContent("i", "x"); got != "i  x" {
		t.Errorf("expected %q, got %q", "i  x", got)
	}
}

func TestRenderMarkdown_FallbackOnEmpty(t *testing.T) {
	// Glamour renders an empty doc to whitespace. Our trim should give
	// back an empty string, never nil/panic.
	if out := renderMarkdown("", 80); out != "" {
		t.Errorf("expected empty render of empty input, got %q", out)
	}
}

func TestRenderMarkdown_PreservesPlainText(t *testing.T) {
	// Plain text round-trips through glamour: it splits words across
	// styled chunks, so we check each word is present rather than the
	// exact substring.
	out := renderMarkdown("hello world", 80)
	if !strings.Contains(out, "hello") || !strings.Contains(out, "world") {
		t.Errorf("expected both words in output, got %q", out)
	}
}
