package editor

// CellStyle is reserved for the ANSI color phase.
type CellStyle struct {
	FG        string `json:"fg"`
	BG        string `json:"bg"`
	Bold      bool   `json:"bold"`
	Underline bool   `json:"underline"`
}

// DrawMode is reserved for graphical character editing.
type DrawMode int

const (
	DrawModeNone DrawMode = iota
	DrawModeBox
	DrawModeLine
)

// FeatureFlags provide a stable toggle point for later phases.
type FeatureFlags struct {
	ANSIColors bool `json:"ansiColors"`
	Drawing    bool `json:"drawing"`
}
