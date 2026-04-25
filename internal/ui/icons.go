package ui

// Nerd-Font glyph constants. Single source of truth so a glyph swap is
// a one-file change. We use Go unicode escapes (\uXXXX / \UXXXXXXXX)
// instead of literal glyphs because tooling that touches this file
// (formatters, copy-paste through web-based editors, terminals without
// the font installed) tends to silently strip Private Use Area
// codepoints, leaving empty string literals.
const (
	// Branding & navigation
	iconBrand     = ""     // nf-fa-microphone     — persona brand
	iconAssistant = "\U000f16a3" // nf-md-robot          — assistant / persona avatar
	iconUser      = ""     // nf-fa-user           — user message
	iconClock     = ""     // nf-fa-clock-o        — timestamps

	// Conversational status
	iconRecord     = ""     // nf-fa-microphone   — recording
	iconTranscribe = ""     // nf-fa-pencil       — transcribing
	iconThink      = "\U000f07f6" // nf-md-thought-bubble — thinking
	iconSpeak      = ""     // nf-fa-volume-up    — generating audio
	iconPlaying    = "\U000f057e" // nf-md-volume-high  — playing
	iconMute       = "\U000f075f" // nf-md-volume-mute  — silent mode
	iconReset      = "\U000f0450" // nf-md-restart      — cleanup / reset

	// Banners
	iconError   = "" // nf-fa-exclamation-circle
	iconSuccess = "" // nf-fa-check-circle
	iconInfo    = "" // nf-fa-info-circle
	iconWarning = "" // nf-fa-exclamation-triangle

	// Powerline rounded chevrons — wrap a pill so it renders as a capsule
	// over the container background. The cap's foreground carries the
	// pill's fill colour; its background stays default so the surrounding
	// surface shows through.
	chevronLeft  = "" // U+E0B6 half-circle-thick left
	chevronRight = "" // U+E0B4 half-circle-thick right
)

// personaGlyphs maps the built-in persona names to a distinctive Nerd
// Font glyph so the selector list is scannable at a glance. User-created
// personas fall back to iconAssistant.
var personaGlyphs = map[string]string{
	"marceline": "", // nf-fa-music     — bass-playing vampire (Adventure Time)
	"freud":     "", // nf-fa-book      — psychoanalyst, books
	"coach":     "", // nf-fa-bolt      — energy / motivation
	"kevin":     "", // nf-fa-gamepad   — young gamer-hacker
	"merlin":    "", // nf-fa-magic     — sparkles / wizardry
	"racoon":    "", // nf-fa-paw       — animal paw print
	"persona":   "", // nf-fa-microphone — default built-in fallback
}

// personaIcon returns a glyph that visually identifies the persona in
// lists, titles and message bubbles. Unknown names receive the generic
// assistant glyph so custom personas stay visually consistent.
func personaIcon(name string) string {
	if g, ok := personaGlyphs[name]; ok {
		return g
	}
	return iconAssistant
}
