// Package editor provides text width and grapheme segmentation helpers.
package editor

import (
	"unicode"

	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
	"golang.org/x/text/unicode/norm"
)

// WidthEngine normalizes and measures mixed-width graphemes deterministically.
type WidthEngine struct {
	ambiguousAsWide bool
	overrides       map[rune]int
}

func NewWidthEngine(ambiguousAsWide bool) *WidthEngine {
	return &WidthEngine{
		ambiguousAsWide: ambiguousAsWide,
		overrides: map[rune]int{
			'·': 1, // Keep middle dot narrow for predictable Korean/English columns.
			// Box drawing glyphs are treated as 2 columns (see GraphemeWidth range override).
		},
	}
}

func (w *WidthEngine) NormalizeNFC(text string) string {
	return norm.NFC.String(text)
}

func (w *WidthEngine) GraphemeWidth(cluster string) int {
	if cluster == "" {
		return 0
	}

	cond := runewidth.DefaultCondition
	cond.EastAsianWidth = w.ambiguousAsWide

	width := 0
	for _, r := range cluster {
		// Box-drawing and block-element glyphs are East Asian Ambiguous and often
		// render double-width in CJK terminal fonts. Treat them as 2 columns.
		if (r >= 0x2500 && r <= 0x257F) || (r >= 0x2580 && r <= 0x259F) {
			width += 2
			continue
		}
		if override, ok := w.overrides[r]; ok {
			width += override
			continue
		}
		if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) {
			continue
		}
		width += cond.RuneWidth(r)
	}

	if width <= 0 {
		return 1
	}
	return width
}

type Grapheme struct {
	Text  string `json:"text"`
	Width int    `json:"width"`
	Color string `json:"color,omitempty"`
	BgColor string `json:"bgColor,omitempty"`
}

func (w *WidthEngine) Segment(text string) []Grapheme {
	normalized := w.NormalizeNFC(text)
	g := uniseg.NewGraphemes(normalized)
	result := make([]Grapheme, 0, len(normalized))
	for g.Next() {
		cluster := g.Str()
		result = append(result, Grapheme{
			Text:  cluster,
			Width: w.GraphemeWidth(cluster),
		})
	}
	return result
}

func (w *WidthEngine) StringWidth(text string) int {
	total := 0
	for _, cluster := range w.Segment(text) {
		total += cluster.Width
	}
	return total
}
