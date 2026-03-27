package app

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/runableapp/simplescreeneditor/internal/editor"
	"github.com/runableapp/simplescreeneditor/internal/platform"
)

type State struct {
	Rows     int                 `json:"rows"`
	Cols     int                 `json:"cols"`
	Cursor   editor.Cursor       `json:"cursor"`
	Dirty    bool                `json:"dirty"`
	Filename string              `json:"filename"`
	Lines    [][]editor.RowToken `json:"lines"`
}

type EditorApp struct {
	mu        sync.Mutex
	buffer    *editor.Buffer
	cursor    editor.Cursor
	clipboard platform.Clipboard
	fileName  string
	dirty     bool
	activeANSIFGColor string
	activeANSIBGColor string
}

const (
	connNorth = 1
	connEast  = 2
	connSouth = 4
	connWest  = 8
)

type lineStyle struct {
	maskToChar map[int]string
	charToMask map[string]int
}

var lineStyles = map[string]lineStyle{
	"single": {
		maskToChar: map[int]string{
			0:  " ",
			1:  "│",
			2:  "─",
			3:  "└",
			4:  "│",
			5:  "│",
			6:  "┌",
			7:  "├",
			8:  "─",
			9:  "┘",
			10: "─",
			11: "┴",
			12: "┐",
			13: "┤",
			14: "┬",
			15: "┼",
		},
		charToMask: map[string]int{
			"│": connNorth | connSouth,
			"─": connEast | connWest,
			"└": connNorth | connEast,
			"┘": connNorth | connWest,
			"┌": connSouth | connEast,
			"┐": connSouth | connWest,
			"├": connNorth | connEast | connSouth,
			"┤": connNorth | connSouth | connWest,
			"┬": connEast | connSouth | connWest,
			"┴": connNorth | connEast | connWest,
			"┼": connNorth | connEast | connSouth | connWest,
		},
	},
	"double": {
		maskToChar: map[int]string{
			0:  " ",
			1:  "║",
			2:  "═",
			3:  "╚",
			4:  "║",
			5:  "║",
			6:  "╔",
			7:  "╠",
			8:  "═",
			9:  "╝",
			10: "═",
			11: "╩",
			12: "╗",
			13: "╣",
			14: "╦",
			15: "╬",
		},
		charToMask: map[string]int{
			"║": connNorth | connSouth,
			"═": connEast | connWest,
			"╚": connNorth | connEast,
			"╝": connNorth | connWest,
			"╔": connSouth | connEast,
			"╗": connSouth | connWest,
			"╠": connNorth | connEast | connSouth,
			"╣": connNorth | connSouth | connWest,
			"╦": connEast | connSouth | connWest,
			"╩": connNorth | connEast | connWest,
			"╬": connNorth | connEast | connSouth | connWest,
		},
	},
	"thick": {
		maskToChar: map[int]string{
			0:  " ",
			1:  "┃",
			2:  "━",
			3:  "┗",
			4:  "┃",
			5:  "┃",
			6:  "┏",
			7:  "┣",
			8:  "━",
			9:  "┛",
			10: "━",
			11: "┻",
			12: "┓",
			13: "┫",
			14: "┳",
			15: "╋",
		},
		charToMask: map[string]int{
			"┃": connNorth | connSouth,
			"━": connEast | connWest,
			"┗": connNorth | connEast,
			"┛": connNorth | connWest,
			"┏": connSouth | connEast,
			"┓": connSouth | connWest,
			"┣": connNorth | connEast | connSouth,
			"┫": connNorth | connSouth | connWest,
			"┳": connEast | connSouth | connWest,
			"┻": connNorth | connEast | connWest,
			"╋": connNorth | connEast | connSouth | connWest,
		},
	},
}

func New(clipboard platform.Clipboard) *EditorApp {
	return &EditorApp{
		buffer:    editor.NewBuffer(editor.NewWidthEngine(true)),
		cursor:    editor.Cursor{Row: 0, Col: 0},
		clipboard: clipboard,
	}
}

func (a *EditorApp) Snapshot() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	lines := make([][]editor.RowToken, 0, editor.Rows)
	for row := 0; row < editor.Rows; row++ {
		tokens, _ := a.buffer.RenderTokens(row)
		lines = append(lines, tokens)
	}

	return State{
		Rows:     editor.Rows,
		Cols:     editor.Columns,
		Cursor:   a.cursor,
		Dirty:    a.dirty,
		Filename: a.fileName,
		Lines:    lines,
	}
}

func (a *EditorApp) InsertText(input string) State {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, g := range editor.NewWidthEngine(true).Segment(input) {
		startCol := a.cursor.Col
		nextCol, err := a.buffer.InsertText(a.cursor.Row, startCol, g.Text)
		if err != nil {
			continue
		}
		if a.activeANSIFGColor != "" {
			_ = a.buffer.SetColorRange(a.cursor.Row, startCol, nextCol, a.activeANSIFGColor)
		}
		if a.activeANSIBGColor != "" {
			_ = a.buffer.SetBGColorRange(a.cursor.Row, startCol, nextCol, a.activeANSIBGColor)
		}
		a.cursor.Col = nextCol
		a.dirty = true
	}
	return a.snapshotLocked()
}

func (a *EditorApp) Backspace() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	nextCol, err := a.buffer.Backspace(a.cursor.Row, a.cursor.Col)
	if err == nil {
		a.cursor.Col = nextCol
		a.dirty = true
	}
	return a.snapshotLocked()
}

func (a *EditorApp) Delete() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.buffer.Delete(a.cursor.Row, a.cursor.Col); err == nil {
		a.dirty = true
	}
	return a.snapshotLocked()
}

func (a *EditorApp) ClearScreen() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.buffer = editor.NewBuffer(editor.NewWidthEngine(true))
	a.cursor = editor.Cursor{Row: 0, Col: 0}
	a.dirty = true
	return a.snapshotLocked()
}

func (a *EditorApp) Enter() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	nextRow, nextCol, err := a.buffer.InsertNewLine(a.cursor.Row, a.cursor.Col)
	if err == nil {
		a.cursor.Row = nextRow
		a.cursor.Col = nextCol
		a.dirty = true
	}
	return a.snapshotLocked()
}

func (a *EditorApp) MoveLeft() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	row, col, err := a.buffer.MoveLeft(a.cursor.Row, a.cursor.Col)
	if err == nil {
		a.cursor.Row, a.cursor.Col = row, col
	}
	return a.snapshotLocked()
}

func (a *EditorApp) MoveRight() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	row, col, err := a.buffer.MoveRight(a.cursor.Row, a.cursor.Col)
	if err == nil {
		a.cursor.Row, a.cursor.Col = row, col
	}
	return a.snapshotLocked()
}

func (a *EditorApp) MoveUp() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cursor.Row > 0 {
		a.cursor.Row--
		if snapped, err := a.buffer.SnapColumn(a.cursor.Row, a.cursor.Col); err == nil {
			a.cursor.Col = snapped
		}
	}
	return a.snapshotLocked()
}

func (a *EditorApp) MoveDown() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cursor.Row < editor.Rows-1 {
		a.cursor.Row++
		if snapped, err := a.buffer.SnapColumn(a.cursor.Row, a.cursor.Col); err == nil {
			a.cursor.Col = snapped
		}
	}
	return a.snapshotLocked()
}

func (a *EditorApp) SetCursor(row, col int) State {
	a.mu.Lock()
	defer a.mu.Unlock()

	if row < 0 {
		row = 0
	}
	if row >= editor.Rows {
		row = editor.Rows - 1
	}
	if col < 0 {
		col = 0
	}
	if col >= editor.Columns {
		col = editor.Columns - 1
	}
	a.cursor.Row = row
	if snapped, err := a.buffer.SnapColumn(row, col); err == nil {
		a.cursor.Col = snapped
	} else {
		a.cursor.Col = 0
	}
	return a.snapshotLocked()
}

func (a *EditorApp) MoveHome() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	if col, err := a.buffer.LineStart(a.cursor.Row); err == nil {
		a.cursor.Col = col
	}
	return a.snapshotLocked()
}

func (a *EditorApp) MoveEnd() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	if col, err := a.buffer.LineEnd(a.cursor.Row); err == nil {
		a.cursor.Col = col
	}
	return a.snapshotLocked()
}

func (a *EditorApp) MoveTop() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.cursor.Row = 0
	a.cursor.Col = 0
	return a.snapshotLocked()
}

func (a *EditorApp) MoveBottom() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.cursor.Row = editor.Rows - 1
	if col, err := a.buffer.LineEnd(a.cursor.Row); err == nil {
		a.cursor.Col = col
	} else {
		a.cursor.Col = 0
	}
	return a.snapshotLocked()
}

func (a *EditorApp) DrawLineLeft(style string) State {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.drawLineStepLocked(style, 0, -1)
	return a.snapshotLocked()
}

func (a *EditorApp) DrawLineRight(style string) State {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.drawLineStepLocked(style, 0, 1)
	return a.snapshotLocked()
}

func (a *EditorApp) DrawLineUp(style string) State {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.drawLineStepLocked(style, -1, 0)
	return a.snapshotLocked()
}

func (a *EditorApp) DrawLineDown(style string) State {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.drawLineStepLocked(style, 1, 0)
	return a.snapshotLocked()
}

func (a *EditorApp) SetActiveANSIColor(color string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.activeANSIFGColor = color
}

func (a *EditorApp) SetActiveANSIFGColor(color string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.activeANSIFGColor = color
}

func (a *EditorApp) SetActiveANSIBGColor(color string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.activeANSIBGColor = color
}

func (a *EditorApp) ClearANSIColors() State {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.buffer.ClearColors()
	a.dirty = true
	return a.snapshotLocked()
}

func (a *EditorApp) SetRegionColor(startRow, startCol, endRow, endCol int, color string) State {
	a.mu.Lock()
	defer a.mu.Unlock()
	r1, c1, r2, c2 := normalizeRect(startRow, startCol, endRow, endCol)
	for row := r1; row <= r2; row++ {
		_ = a.buffer.SetColorRange(row, c1, c2+1, color)
	}
	a.dirty = true
	return a.snapshotLocked()
}

func (a *EditorApp) SetRegionBGColor(startRow, startCol, endRow, endCol int, color string) State {
	a.mu.Lock()
	defer a.mu.Unlock()
	r1, c1, r2, c2 := normalizeRect(startRow, startCol, endRow, endCol)
	for row := r1; row <= r2; row++ {
		_ = a.buffer.SetBGColorRange(row, c1, c2+1, color)
	}
	a.dirty = true
	return a.snapshotLocked()
}

func (a *EditorApp) FillRegion(startRow, startCol, endRow, endCol int, text string) State {
	a.mu.Lock()
	defer a.mu.Unlock()
	r1, c1, r2, c2 := normalizeRect(startRow, startCol, endRow, endCol)
	for row := r1; row <= r2; row++ {
		_ = a.buffer.FillCells(row, c1, c2+1, text, a.activeANSIFGColor, a.activeANSIBGColor)
	}
	a.dirty = true
	return a.snapshotLocked()
}

func (a *EditorApp) ClearRegion(startRow, startCol, endRow, endCol int) State {
	a.mu.Lock()
	defer a.mu.Unlock()
	r1, c1, r2, c2 := normalizeRect(startRow, startCol, endRow, endCol)
	for row := r1; row <= r2; row++ {
		_ = a.buffer.FillCells(row, c1, c2+1, " ", "", "")
	}
	a.dirty = true
	return a.snapshotLocked()
}

func (a *EditorApp) CopyBuffer() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.clipboard.WriteText(strings.Join(a.buffer.LinesAsANSIText(), "\n"))
}

func (a *EditorApp) CopyRegion(startRow, startCol, endRow, endCol int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	r1, c1, r2, c2 := normalizeRect(startRow, startCol, endRow, endCol)
	var out strings.Builder
	for row := r1; row <= r2; row++ {
		tokens, err := a.buffer.RenderTokens(row)
		if err != nil {
			continue
		}
		cellText := make([]string, editor.Columns)
		cellColor := make([]string, editor.Columns)
		cellBGColor := make([]string, editor.Columns)
		for _, token := range tokens {
			if token.Col < 0 || token.Col >= editor.Columns {
				continue
			}
			cellText[token.Col] = token.Text
			cellColor[token.Col] = token.Color
			cellBGColor[token.Col] = token.BgColor
			for i := 1; i < token.Width && token.Col+i < editor.Columns; i++ {
				cellText[token.Col+i] = " "
				cellColor[token.Col+i] = token.Color
				cellBGColor[token.Col+i] = token.BgColor
			}
		}
		activeColor := ""
		activeBGColor := ""
		for col := c1; col <= c2; col++ {
			color := cellColor[col]
			bgColor := cellBGColor[col]
			if color != activeColor {
				if color == "" {
					out.WriteString("\x1b[39m")
					activeColor = ""
				} else if code, ok := ansiColorCodeForApp(color); ok {
					out.WriteString("\x1b[")
					out.WriteString(code)
					out.WriteString("m")
					activeColor = color
				}
			}
			if bgColor != activeBGColor {
				if bgColor == "" {
					out.WriteString("\x1b[49m")
					activeBGColor = ""
				} else if code, ok := ansiBGColorCodeForApp(bgColor); ok {
					out.WriteString("\x1b[")
					out.WriteString(code)
					out.WriteString("m")
					activeBGColor = bgColor
				}
			}
			ch := cellText[col]
			if ch == "" {
				ch = " "
			}
			out.WriteString(ch)
		}
		if activeColor != "" || activeBGColor != "" {
			out.WriteString("\x1b[0m")
		}
		if row < r2 {
			out.WriteByte('\n')
		}
	}
	return a.clipboard.WriteText(out.String())
}

func (a *EditorApp) PasteFromClipboard() State {
	a.mu.Lock()
	defer a.mu.Unlock()

	text, err := a.clipboard.ReadText()
	if err == nil {
		a.overwriteANSITextLocked(text)
	}
	return a.snapshotLocked()
}

func (a *EditorApp) SaveFile(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if path == "" {
		path = a.fileName
	}
	if path == "" {
		return os.ErrInvalid
	}

	data := strings.Join(a.buffer.LinesAsText(), "\n")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		return err
	}
	a.fileName = path
	a.dirty = false
	return nil
}

func (a *EditorApp) SaveFileANSI(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if path == "" {
		path = a.fileName
	}
	if path == "" {
		return os.ErrInvalid
	}

	data := strings.Join(a.buffer.LinesAsANSIText(), "\n")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		return err
	}
	a.fileName = path
	a.dirty = false
	return nil
}

func (a *EditorApp) OpenFile(path string) (State, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	raw, err := os.ReadFile(path)
	if err != nil {
		return State{}, err
	}
	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	a.buffer.SetLines(lines)
	a.fileName = path
	a.dirty = false
	a.cursor = editor.Cursor{Row: 0, Col: 0}
	return a.snapshotLocked(), nil
}

func (a *EditorApp) CurrentFilename() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.fileName
}

func (a *EditorApp) snapshotLocked() State {
	lines := make([][]editor.RowToken, 0, editor.Rows)
	for row := 0; row < editor.Rows; row++ {
		tokens, _ := a.buffer.RenderTokens(row)
		lines = append(lines, tokens)
	}
	return State{
		Rows:     editor.Rows,
		Cols:     editor.Columns,
		Cursor:   a.cursor,
		Dirty:    a.dirty,
		Filename: a.fileName,
		Lines:    lines,
	}
}

func (a *EditorApp) insertTextLocked(input string) {
	widthEngine := editor.NewWidthEngine(true)
	for _, segment := range widthEngine.Segment(input) {
		if segment.Text == "\n" {
			nextRow, nextCol, err := a.buffer.InsertNewLine(a.cursor.Row, a.cursor.Col)
			if err == nil {
				a.cursor.Row = nextRow
				a.cursor.Col = nextCol
				a.dirty = true
			}
			continue
		}
		startCol := a.cursor.Col
		nextCol, err := a.buffer.InsertText(a.cursor.Row, startCol, segment.Text)
		if err != nil {
			continue
		}
		if a.activeANSIFGColor != "" {
			_ = a.buffer.SetColorRange(a.cursor.Row, startCol, nextCol, a.activeANSIFGColor)
		}
		if a.activeANSIBGColor != "" {
			_ = a.buffer.SetBGColorRange(a.cursor.Row, startCol, nextCol, a.activeANSIBGColor)
		}
		a.cursor.Col = nextCol
		a.dirty = true
	}
}

func (a *EditorApp) insertANSITextLocked(input string) {
	type piece struct {
		text    string
		fgColor string
		bgColor string
	}
	defaultFGColor := a.activeANSIFGColor
	defaultBGColor := a.activeANSIBGColor
	currentFGColor := defaultFGColor
	currentBGColor := defaultBGColor
	pieces := make([]piece, 0)
	var buf strings.Builder
	flush := func() {
		if buf.Len() == 0 {
			return
		}
		pieces = append(pieces, piece{text: buf.String(), fgColor: currentFGColor, bgColor: currentBGColor})
		buf.Reset()
	}

	for i := 0; i < len(input); {
		if input[i] == 0x1b && i+1 < len(input) && input[i+1] == '[' {
			j := i + 2
			for j < len(input) && input[j] != 'm' {
				j++
			}
			if j < len(input) && input[j] == 'm' {
				flush()
				params := input[i+2 : j]
				if params == "" {
					currentFGColor = ""
					currentBGColor = ""
				} else {
					for _, part := range strings.Split(params, ";") {
						if part == "" {
							part = "0"
						}
						code, err := strconv.Atoi(part)
						if err != nil {
							continue
						}
						if code == 0 {
							currentFGColor = ""
							currentBGColor = ""
							continue
						}
						if mapped, ok := ansiCodeToFGColorForApp(code); ok {
							currentFGColor = mapped
						}
						if mapped, ok := ansiCodeToBGColorForApp(code); ok {
							currentBGColor = mapped
						}
						if code == 39 {
							currentFGColor = ""
						}
						if code == 49 {
							currentBGColor = ""
						}
					}
				}
				i = j + 1
				continue
			}
		}
		buf.WriteByte(input[i])
		i++
	}
	flush()

	widthEngine := editor.NewWidthEngine(true)
	for _, p := range pieces {
		for _, segment := range widthEngine.Segment(p.text) {
			if segment.Text == "\r" {
				continue
			}
			if segment.Text == "\n" {
				nextRow, nextCol, err := a.buffer.InsertNewLine(a.cursor.Row, a.cursor.Col)
				if err == nil {
					a.cursor.Row = nextRow
					a.cursor.Col = nextCol
					a.dirty = true
				}
				continue
			}
			startCol := a.cursor.Col
			nextCol, err := a.buffer.InsertText(a.cursor.Row, startCol, segment.Text)
			if err != nil {
				continue
			}
			if p.fgColor != "" {
				_ = a.buffer.SetColorRange(a.cursor.Row, startCol, nextCol, p.fgColor)
			}
			if p.bgColor != "" {
				_ = a.buffer.SetBGColorRange(a.cursor.Row, startCol, nextCol, p.bgColor)
			}
			a.cursor.Col = nextCol
			a.dirty = true
		}
	}
}

func (a *EditorApp) overwriteANSITextLocked(input string) {
	type piece struct {
		text    string
		fgColor string
		bgColor string
	}
	defaultFGColor := a.activeANSIFGColor
	defaultBGColor := a.activeANSIBGColor
	currentFGColor := defaultFGColor
	currentBGColor := defaultBGColor
	pieces := make([]piece, 0)
	var buf strings.Builder
	flush := func() {
		if buf.Len() == 0 {
			return
		}
		pieces = append(pieces, piece{text: buf.String(), fgColor: currentFGColor, bgColor: currentBGColor})
		buf.Reset()
	}

	for i := 0; i < len(input); {
		if input[i] == 0x1b && i+1 < len(input) && input[i+1] == '[' {
			j := i + 2
			for j < len(input) && input[j] != 'm' {
				j++
			}
			if j < len(input) && input[j] == 'm' {
				flush()
				params := input[i+2 : j]
				if params == "" {
					currentFGColor = ""
					currentBGColor = ""
				} else {
					for _, part := range strings.Split(params, ";") {
						if part == "" {
							part = "0"
						}
						code, err := strconv.Atoi(part)
						if err != nil {
							continue
						}
						if code == 0 {
							currentFGColor = ""
							currentBGColor = ""
							continue
						}
						if mapped, ok := ansiCodeToFGColorForApp(code); ok {
							currentFGColor = mapped
						}
						if mapped, ok := ansiCodeToBGColorForApp(code); ok {
							currentBGColor = mapped
						}
						if code == 39 {
							currentFGColor = ""
						}
						if code == 49 {
							currentBGColor = ""
						}
					}
				}
				i = j + 1
				continue
			}
		}
		buf.WriteByte(input[i])
		i++
	}
	flush()

	widthEngine := editor.NewWidthEngine(true)
	lineStartCol := a.cursor.Col

	for _, p := range pieces {
		for _, segment := range widthEngine.Segment(p.text) {
			if segment.Text == "\r" {
				continue
			}
			if segment.Text == "\n" {
				if a.cursor.Row < editor.Rows-1 {
					a.cursor.Row++
				}
				a.cursor.Col = lineStartCol
				continue
			}
			if a.cursor.Col >= editor.Columns {
				continue
			}
			nextCol, err := a.buffer.OverwriteAtWithColors(a.cursor.Row, a.cursor.Col, segment.Text, p.fgColor, p.bgColor)
			if err != nil {
				continue
			}
			a.cursor.Col = nextCol
			a.dirty = true
		}
	}
}

func ansiColorCodeForApp(name string) (string, bool) {
	switch name {
	case "black":
		return "30", true
	case "red":
		return "31", true
	case "green":
		return "32", true
	case "yellow":
		return "33", true
	case "blue":
		return "34", true
	case "magenta":
		return "35", true
	case "cyan":
		return "36", true
	case "white":
		return "37", true
	case "brightBlack":
		return "90", true
	case "brightRed":
		return "91", true
	case "brightGreen":
		return "92", true
	case "brightYellow":
		return "93", true
	case "brightBlue":
		return "94", true
	case "brightMagenta":
		return "95", true
	case "brightCyan":
		return "96", true
	case "brightWhite":
		return "97", true
	default:
		return "", false
	}
}

func ansiBGColorCodeForApp(name string) (string, bool) {
	switch name {
	case "black":
		return "40", true
	case "red":
		return "41", true
	case "green":
		return "42", true
	case "yellow":
		return "43", true
	case "blue":
		return "44", true
	case "magenta":
		return "45", true
	case "cyan":
		return "46", true
	case "white":
		return "47", true
	case "brightBlack":
		return "100", true
	case "brightRed":
		return "101", true
	case "brightGreen":
		return "102", true
	case "brightYellow":
		return "103", true
	case "brightBlue":
		return "104", true
	case "brightMagenta":
		return "105", true
	case "brightCyan":
		return "106", true
	case "brightWhite":
		return "107", true
	default:
		return "", false
	}
}

func ansiCodeToFGColorForApp(code int) (string, bool) {
	switch code {
	case 30:
		return "black", true
	case 31:
		return "red", true
	case 32:
		return "green", true
	case 33:
		return "yellow", true
	case 34:
		return "blue", true
	case 35:
		return "magenta", true
	case 36:
		return "cyan", true
	case 37:
		return "white", true
	case 90:
		return "brightBlack", true
	case 91:
		return "brightRed", true
	case 92:
		return "brightGreen", true
	case 93:
		return "brightYellow", true
	case 94:
		return "brightBlue", true
	case 95:
		return "brightMagenta", true
	case 96:
		return "brightCyan", true
	case 97:
		return "brightWhite", true
	default:
		return "", false
	}
}

func ansiCodeToBGColorForApp(code int) (string, bool) {
	switch code {
	case 40:
		return "black", true
	case 41:
		return "red", true
	case 42:
		return "green", true
	case 43:
		return "yellow", true
	case 44:
		return "blue", true
	case 45:
		return "magenta", true
	case 46:
		return "cyan", true
	case 47:
		return "white", true
	case 100:
		return "brightBlack", true
	case 101:
		return "brightRed", true
	case 102:
		return "brightGreen", true
	case 103:
		return "brightYellow", true
	case 104:
		return "brightBlue", true
	case 105:
		return "brightMagenta", true
	case 106:
		return "brightCyan", true
	case 107:
		return "brightWhite", true
	default:
		return "", false
	}
}

func styleByName(name string) lineStyle {
	style, ok := lineStyles[name]
	if ok {
		return style
	}
	return lineStyles["single"]
}

func connectionBit(dr, dc int) int {
	switch {
	case dr == -1:
		return connNorth
	case dr == 1:
		return connSouth
	case dc == -1:
		return connWest
	case dc == 1:
		return connEast
	default:
		return 0
	}
}

func oppositeBit(bit int) int {
	switch bit {
	case connNorth:
		return connSouth
	case connSouth:
		return connNorth
	case connWest:
		return connEast
	case connEast:
		return connWest
	default:
		return 0
	}
}

func (a *EditorApp) drawLineStepLocked(styleName string, dr, dc int) {
	nextRow := a.cursor.Row + dr
	nextCol := a.cursor.Col + dc
	if nextRow < 0 || nextRow >= editor.Rows || nextCol < 0 || nextCol >= editor.Columns {
		return
	}

	style := styleByName(styleName)
	dir := connectionBit(dr, dc)
	rev := oppositeBit(dir)
	if dir == 0 || rev == 0 {
		return
	}

	currentMask := a.lineMaskAt(a.cursor.Row, a.cursor.Col, style)
	nextMask := a.lineMaskAt(nextRow, nextCol, style)
	currentMask |= dir
	nextMask |= rev

	currentGlyph := style.maskToChar[currentMask]
	nextGlyph := style.maskToChar[nextMask]
	if currentGlyph == "" || nextGlyph == "" {
		return
	}
	if err := a.buffer.SetCharAtWithColors(a.cursor.Row, a.cursor.Col, currentGlyph, a.activeANSIFGColor, a.activeANSIBGColor); err != nil {
		return
	}
	if err := a.buffer.SetCharAtWithColors(nextRow, nextCol, nextGlyph, a.activeANSIFGColor, a.activeANSIBGColor); err != nil {
		return
	}
	a.cursor.Row = nextRow
	a.cursor.Col = nextCol
	a.dirty = true
}

func (a *EditorApp) lineMaskAt(row, col int, style lineStyle) int {
	ch, err := a.buffer.CharAt(row, col)
	if err != nil || ch == "" {
		return 0
	}
	mask, ok := style.charToMask[ch]
	if !ok {
		return 0
	}
	// Horizontal/vertical glyphs represent two directions, but at endpoints we
	// want the currently connected side(s) only. Infer active links from
	// neighboring cells and fall back to the glyph's canonical mask when isolated.
	if neighborMask := a.lineMaskFromNeighbors(row, col, style); neighborMask != 0 {
		return neighborMask & mask
	}
	return mask
}

func (a *EditorApp) lineMaskFromNeighbors(row, col int, style lineStyle) int {
	type direction struct {
		dr  int
		dc  int
		bit int
	}
	dirs := []direction{
		{dr: -1, dc: 0, bit: connNorth},
		{dr: 0, dc: 1, bit: connEast},
		{dr: 1, dc: 0, bit: connSouth},
		{dr: 0, dc: -1, bit: connWest},
	}
	mask := 0
	for _, dir := range dirs {
		nr := row + dir.dr
		nc := col + dir.dc
		if nr < 0 || nr >= editor.Rows || nc < 0 || nc >= editor.Columns {
			continue
		}
		neighbor, err := a.buffer.CharAt(nr, nc)
		if err != nil || neighbor == "" {
			continue
		}
		neighborMask, ok := style.charToMask[neighbor]
		if !ok {
			continue
		}
		if neighborMask&oppositeBit(dir.bit) != 0 {
			mask |= dir.bit
		}
	}
	return mask
}

func normalizeRect(startRow, startCol, endRow, endCol int) (int, int, int, int) {
	r1, r2 := startRow, endRow
	c1, c2 := startCol, endCol
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	if c1 > c2 {
		c1, c2 = c2, c1
	}
	r1 = max(0, min(r1, editor.Rows-1))
	r2 = max(0, min(r2, editor.Rows-1))
	c1 = max(0, min(c1, editor.Columns-1))
	c2 = max(0, min(c2, editor.Columns-1))
	return r1, c1, r2, c2
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
