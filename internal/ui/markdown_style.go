package ui

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
)

// Glamour markdown theme bound to the Catppuccin palette defined in
// theme.go, so changing flavor there re-themes both the TUI chrome and
// the rendered markdown in one place. Accents (H1, Strong, blockquote
// bar, code-block frame) use Peach to match the pill system; links and
// enumerations use Sky (accentAlt).

func sp(s string) *string { return &s }
func bp(b bool) *bool     { return &b }
func up(u uint) *uint     { return &u }

// markdownStyleConfig returns a glamour StyleConfig built from the
// current Catppuccin flavor. Colours are read at call time so a runtime
// flavor swap reflects on the next renderer.
func markdownStyleConfig() ansi.StyleConfig {
	var (
		base     = flavor.Base().Hex
		text     = flavor.Text().Hex
		subtext  = flavor.Subtext0().Hex
		overlay0 = flavor.Overlay0().Hex
		overlay1 = flavor.Overlay1().Hex
		surface0 = flavor.Surface0().Hex
		surface1 = flavor.Surface1().Hex
		peach    = flavor.Peach().Hex
		sky      = flavor.Sky().Hex
		mauve    = flavor.Mauve().Hex
		lavender = flavor.Lavender().Hex
		green    = flavor.Green().Hex
		yellow   = flavor.Yellow().Hex
		red      = flavor.Red().Hex
		pink     = flavor.Pink().Hex
		teal     = flavor.Teal().Hex
		sapphire = flavor.Sapphire().Hex
		blue     = flavor.Blue().Hex
	)

	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockPrefix: "",
				BlockSuffix: "",
				Color:       sp(text),
			},
		},

		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  sp(peach),
				Italic: bp(true),
			},
			Indent:      up(1),
			IndentToken: sp("▍ "),
		},

		Paragraph: ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{}},

		List: ansi.StyleList{
			LevelIndent: 2,
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{Color: sp(text)},
			},
		},

		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockSuffix: "\n",
				Color:       sp(peach),
				Bold:        bp(true),
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix:          " ",
				Suffix:          " ",
				Color:           sp(base),
				BackgroundColor: sp(peach),
				Bold:            bp(true),
			},
		},
		H2: ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Prefix: "▍ "}},
		H3: ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Prefix: "▸ ", Color: sp(mauve)}},
		H4: ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Prefix: "• ", Color: sp(lavender)}},
		H5: ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Prefix: "  · ", Color: sp(subtext)}},
		H6: ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{Prefix: "  · ", Color: sp(overlay1)}},

		Strikethrough: ansi.StylePrimitive{CrossedOut: bp(true)},
		Emph:          ansi.StylePrimitive{Color: sp(lavender), Italic: bp(true)},
		Strong:        ansi.StylePrimitive{Color: sp(peach), Bold: bp(true)},

		HorizontalRule: ansi.StylePrimitive{
			Color:  sp(surface1),
			Format: "\n────────────────\n",
		},

		Item:        ansi.StylePrimitive{BlockPrefix: "• "},
		Enumeration: ansi.StylePrimitive{BlockPrefix: ". ", Color: sp(sky)},
		Task: ansi.StyleTask{
			StylePrimitive: ansi.StylePrimitive{},
			Ticked:         "[✓] ",
			Unticked:       "[ ] ",
		},

		Link:      ansi.StylePrimitive{Color: sp(sky), Underline: bp(true)},
		LinkText:  ansi.StylePrimitive{Color: sp(sapphire)},
		Image:     ansi.StylePrimitive{Color: sp(sky), Underline: bp(true)},
		ImageText: ansi.StylePrimitive{Color: sp(pink), Format: "Image: {{.text}} →"},

		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:           sp(green),
				BackgroundColor: sp(surface0),
				Prefix:          " ",
				Suffix:          " ",
			},
		},

		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{Color: sp(text)},
				Margin:         up(1),
			},
			Chroma: &ansi.Chroma{
				Text:                ansi.StylePrimitive{Color: sp(text)},
				Error:               ansi.StylePrimitive{Color: sp(text), BackgroundColor: sp(red)},
				Comment:             ansi.StylePrimitive{Color: sp(overlay0), Italic: bp(true)},
				CommentPreproc:      ansi.StylePrimitive{Color: sp(pink)},
				Keyword:             ansi.StylePrimitive{Color: sp(mauve)},
				KeywordReserved:     ansi.StylePrimitive{Color: sp(mauve)},
				KeywordNamespace:    ansi.StylePrimitive{Color: sp(pink)},
				KeywordType:         ansi.StylePrimitive{Color: sp(yellow)},
				Operator:            ansi.StylePrimitive{Color: sp(sky)},
				Punctuation:         ansi.StylePrimitive{Color: sp(overlay1)},
				Name:                ansi.StylePrimitive{Color: sp(text)},
				NameBuiltin:         ansi.StylePrimitive{Color: sp(peach)},
				NameTag:             ansi.StylePrimitive{Color: sp(mauve)},
				NameAttribute:       ansi.StylePrimitive{Color: sp(yellow)},
				NameClass:           ansi.StylePrimitive{Color: sp(yellow)},
				NameConstant:        ansi.StylePrimitive{Color: sp(peach)},
				NameDecorator:       ansi.StylePrimitive{Color: sp(peach)},
				NameFunction:        ansi.StylePrimitive{Color: sp(blue)},
				LiteralNumber:       ansi.StylePrimitive{Color: sp(peach)},
				LiteralString:       ansi.StylePrimitive{Color: sp(green)},
				LiteralStringEscape: ansi.StylePrimitive{Color: sp(pink)},
				GenericDeleted:      ansi.StylePrimitive{Color: sp(red)},
				GenericEmph:         ansi.StylePrimitive{Color: sp(lavender), Italic: bp(true)},
				GenericInserted:     ansi.StylePrimitive{Color: sp(green)},
				GenericStrong:       ansi.StylePrimitive{Color: sp(peach), Bold: bp(true)},
				GenericSubheading:   ansi.StylePrimitive{Color: sp(teal)},
				Background:          ansi.StylePrimitive{BackgroundColor: sp(surface0)},
			},
		},

		Table: ansi.StyleTable{
			StyleBlock:      ansi.StyleBlock{StylePrimitive: ansi.StylePrimitive{}},
			CenterSeparator: sp("┼"),
			ColumnSeparator: sp("│"),
			RowSeparator:    sp("─"),
		},
		DefinitionDescription: ansi.StylePrimitive{BlockPrefix: "\n🠶 "},
	}
}

// newMarkdownRenderer returns a glamour renderer themed with the
// current Catppuccin flavor. Width < 1 disables word wrapping.
func newMarkdownRenderer(width int) (*glamour.TermRenderer, error) {
	opts := []glamour.TermRendererOption{
		glamour.WithStyles(markdownStyleConfig()),
	}
	if width > 0 {
		opts = append(opts, glamour.WithWordWrap(width))
	}
	return glamour.NewTermRenderer(opts...)
}

// renderMarkdown formats msg using the Catppuccin markdown renderer.
// Falls back to the raw message on any error so we never lose content.
// Trailing whitespace from glamour is trimmed because it inflates the
// containing bubble height.
func renderMarkdown(msg string, width int) string {
	r, err := newMarkdownRenderer(width)
	if err != nil {
		return msg
	}
	out, err := r.Render(msg)
	if err != nil {
		return msg
	}
	return strings.TrimSpace(out)
}
