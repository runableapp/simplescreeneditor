package editor

import (
	"errors"
	"strings"
)

const (
	Columns = 80
	Rows    = 25
)

var (
	ErrOutOfBounds   = errors.New("out of bounds")
	ErrInvalidColumn = errors.New("invalid grapheme boundary")
	ErrLineFull      = errors.New("line does not have enough room")
)

type Buffer struct {
	lines [][]Grapheme
	eng   *WidthEngine
}

func NewBuffer(eng *WidthEngine) *Buffer {
	lines := make([][]Grapheme, Rows)
	for i := 0; i < Rows; i++ {
		lines[i] = []Grapheme{}
	}
	return &Buffer{lines: lines, eng: eng}
}

func (b *Buffer) lineWidth(row int) (int, error) {
	if row < 0 || row >= Rows {
		return 0, ErrOutOfBounds
	}
	width := 0
	for _, g := range b.lines[row] {
		width += g.Width
	}
	return width, nil
}

func (b *Buffer) lineBoundaries(row int) ([]int, error) {
	if row < 0 || row >= Rows {
		return nil, ErrOutOfBounds
	}
	boundaries := []int{0}
	cell := 0
	for _, g := range b.lines[row] {
		cell += g.Width
		boundaries = append(boundaries, cell)
	}
	return boundaries, nil
}

func boundaryIndex(boundaries []int, col int) (int, bool) {
	for i, v := range boundaries {
		if v == col {
			return i, true
		}
	}
	return 0, false
}

func (b *Buffer) InsertText(row, col int, text string) (int, error) {
	if row < 0 || row >= Rows || col < 0 || col > Columns {
		return col, ErrOutOfBounds
	}
	boundaries, err := b.lineBoundaries(row)
	if err != nil {
		return col, err
	}
	insertAt, ok := boundaryIndex(boundaries, col)
	if !ok {
		return col, ErrInvalidColumn
	}
	segments := b.eng.Segment(text)
	insertWidth := 0
	for _, seg := range segments {
		insertWidth += seg.Width
	}
	currentWidth, _ := b.lineWidth(row)
	if currentWidth+insertWidth > Columns {
		return col, ErrLineFull
	}

	left := append([]Grapheme{}, b.lines[row][:insertAt]...)
	right := append([]Grapheme{}, b.lines[row][insertAt:]...)
	next := make([]Grapheme, 0, len(left)+len(segments)+len(right))
	next = append(next, left...)
	next = append(next, segments...)
	next = append(next, right...)
	b.lines[row] = next

	return col + insertWidth, nil
}

func (b *Buffer) Backspace(row, col int) (int, error) {
	if row < 0 || row >= Rows || col < 0 || col > Columns {
		return col, ErrOutOfBounds
	}
	boundaries, err := b.lineBoundaries(row)
	if err != nil {
		return col, err
	}
	idx, ok := boundaryIndex(boundaries, col)
	if !ok {
		return col, ErrInvalidColumn
	}
	if idx == 0 {
		return col, nil
	}
	newCol := boundaries[idx-1]
	b.lines[row] = append(b.lines[row][:idx-1], b.lines[row][idx:]...)
	return newCol, nil
}

func (b *Buffer) Delete(row, col int) error {
	if row < 0 || row >= Rows || col < 0 || col > Columns {
		return ErrOutOfBounds
	}
	boundaries, err := b.lineBoundaries(row)
	if err != nil {
		return err
	}
	idx, ok := boundaryIndex(boundaries, col)
	if !ok {
		return ErrInvalidColumn
	}
	if idx >= len(b.lines[row]) {
		return nil
	}
	b.lines[row] = append(b.lines[row][:idx], b.lines[row][idx+1:]...)
	return nil
}

func (b *Buffer) InsertNewLine(row, col int) (int, int, error) {
	if row < 0 || row >= Rows || col < 0 || col > Columns {
		return row, col, ErrOutOfBounds
	}
	boundaries, err := b.lineBoundaries(row)
	if err != nil {
		return row, col, err
	}
	idx, ok := boundaryIndex(boundaries, col)
	if !ok {
		return row, col, ErrInvalidColumn
	}
	left := append([]Grapheme{}, b.lines[row][:idx]...)
	right := append([]Grapheme{}, b.lines[row][idx:]...)
	b.lines[row] = left

	for r := Rows - 1; r > row+1; r-- {
		b.lines[r] = b.lines[r-1]
	}
	if row+1 < Rows {
		b.lines[row+1] = right
	}
	return min(row+1, Rows-1), 0, nil
}

func (b *Buffer) MoveLeft(row, col int) (int, int, error) {
	if row < 0 || row >= Rows {
		return row, col, ErrOutOfBounds
	}
	boundaries, err := b.lineBoundaries(row)
	if err != nil {
		return row, col, err
	}
	idx, ok := boundaryIndex(boundaries, col)
	if !ok {
		return row, col, ErrInvalidColumn
	}
	if idx > 0 {
		return row, boundaries[idx-1], nil
	}
	return row, col, nil
}

func (b *Buffer) MoveRight(row, col int) (int, int, error) {
	if row < 0 || row >= Rows {
		return row, col, ErrOutOfBounds
	}
	boundaries, err := b.lineBoundaries(row)
	if err != nil {
		return row, col, err
	}
	idx, ok := boundaryIndex(boundaries, col)
	if !ok {
		return row, col, ErrInvalidColumn
	}
	if idx < len(boundaries)-1 {
		return row, boundaries[idx+1], nil
	}
	return row, col, nil
}

func (b *Buffer) LineStart(row int) (int, error) {
	if row < 0 || row >= Rows {
		return 0, ErrOutOfBounds
	}
	return 0, nil
}

func (b *Buffer) LineEnd(row int) (int, error) {
	if row < 0 || row >= Rows {
		return 0, ErrOutOfBounds
	}
	boundaries, err := b.lineBoundaries(row)
	if err != nil {
		return 0, err
	}
	return boundaries[len(boundaries)-1], nil
}

func (b *Buffer) CharAt(row, col int) (string, error) {
	if row < 0 || row >= Rows || col < 0 || col >= Columns {
		return "", ErrOutOfBounds
	}
	boundaries, err := b.lineBoundaries(row)
	if err != nil {
		return "", err
	}
	lineEnd := boundaries[len(boundaries)-1]
	if col > lineEnd {
		return "", nil
	}
	idx, ok := boundaryIndex(boundaries, col)
	if !ok {
		return "", ErrInvalidColumn
	}
	if idx >= len(b.lines[row]) {
		return "", nil
	}
	return b.lines[row][idx].Text, nil
}

func (b *Buffer) SetCharAt(row, col int, text string) error {
	return b.SetCharAtWithColors(row, col, text, "", "")
}

func (b *Buffer) SetCharAtWithColor(row, col int, text, color string) error {
	return b.SetCharAtWithColors(row, col, text, color, "")
}

func (b *Buffer) SetCharAtWithColors(row, col int, text, color, bgColor string) error {
	if row < 0 || row >= Rows || col < 0 || col >= Columns {
		return ErrOutOfBounds
	}
	segments := b.eng.Segment(text)
	if len(segments) != 1 || segments[0].Width != 1 {
		return ErrInvalidColumn
	}
	target := segments[0]
	target.Color = color
	target.BgColor = bgColor
	currentWidth, err := b.lineWidth(row)
	if err != nil {
		return err
	}
	if col > currentWidth {
		space := b.eng.Segment(" ")[0]
		for i := currentWidth; i < col; i++ {
			b.lines[row] = append(b.lines[row], space)
		}
		b.lines[row] = append(b.lines[row], target)
		return nil
	}
	if col == currentWidth {
		b.lines[row] = append(b.lines[row], target)
		return nil
	}

	boundaries, err := b.lineBoundaries(row)
	if err != nil {
		return err
	}
	idx, ok := boundaryIndex(boundaries, col)
	if !ok {
		return ErrInvalidColumn
	}
	if idx >= len(b.lines[row]) {
		b.lines[row] = append(b.lines[row], target)
		return nil
	}
	if b.lines[row][idx].Width != 1 {
		return ErrInvalidColumn
	}
	b.lines[row][idx] = target
	return nil
}

func (b *Buffer) SnapColumn(row, col int) (int, error) {
	if row < 0 || row >= Rows {
		return col, ErrOutOfBounds
	}
	boundaries, err := b.lineBoundaries(row)
	if err != nil {
		return col, err
	}
	if len(boundaries) == 0 {
		return 0, nil
	}
	if col <= 0 {
		return 0, nil
	}
	last := boundaries[len(boundaries)-1]
	if col >= last {
		return last, nil
	}
	for i := len(boundaries) - 1; i >= 0; i-- {
		if boundaries[i] <= col {
			return boundaries[i], nil
		}
	}
	return 0, nil
}

type RowToken struct {
	Col   int    `json:"col"`
	Width int    `json:"width"`
	Text  string `json:"text"`
	Color string `json:"color,omitempty"`
	BgColor string `json:"bgColor,omitempty"`
}

func (b *Buffer) RenderTokens(row int) ([]RowToken, error) {
	if row < 0 || row >= Rows {
		return nil, ErrOutOfBounds
	}
	tokens := make([]RowToken, 0, len(b.lines[row]))
	col := 0
	for _, g := range b.lines[row] {
		tokens = append(tokens, RowToken{
			Col:   col,
			Width: g.Width,
			Text:  g.Text,
			Color: g.Color,
			BgColor: g.BgColor,
		})
		col += g.Width
	}
	return tokens, nil
}

func (b *Buffer) SetColorRange(row, startCol, endCol int, color string) error {
	if row < 0 || row >= Rows || startCol < 0 || endCol < startCol {
		return ErrOutOfBounds
	}
	boundaries, err := b.lineBoundaries(row)
	if err != nil {
		return err
	}
	for i := 0; i < len(b.lines[row]); i++ {
		segStart := boundaries[i]
		segEnd := boundaries[i+1]
		if segStart >= endCol || segEnd <= startCol {
			continue
		}
		b.lines[row][i].Color = color
	}
	return nil
}

func (b *Buffer) SetBGColorRange(row, startCol, endCol int, bgColor string) error {
	if row < 0 || row >= Rows || startCol < 0 || endCol < startCol {
		return ErrOutOfBounds
	}
	boundaries, err := b.lineBoundaries(row)
	if err != nil {
		return err
	}
	for i := 0; i < len(b.lines[row]); i++ {
		segStart := boundaries[i]
		segEnd := boundaries[i+1]
		if segStart >= endCol || segEnd <= startCol {
			continue
		}
		b.lines[row][i].BgColor = bgColor
	}
	return nil
}

func (b *Buffer) ClearColors() {
	for row := 0; row < Rows; row++ {
		for i := range b.lines[row] {
			b.lines[row][i].Color = ""
			b.lines[row][i].BgColor = ""
		}
	}
}

func (b *Buffer) FillCells(row, startCol, endCol int, text, color, bgColor string) error {
	if row < 0 || row >= Rows || startCol < 0 || endCol < startCol || endCol > Columns {
		return ErrOutOfBounds
	}
	if startCol == endCol {
		return nil
	}
	segments := b.eng.Segment(text)
	if len(segments) != 1 || segments[0].Width != 1 {
		return ErrInvalidColumn
	}
	fill := segments[0]
	fill.Color = color
	fill.BgColor = bgColor

	type cell struct {
		g        Grapheme
		start    bool
		cont     bool
		startCol int
	}
	cells := make([]cell, Columns)

	col := 0
	for _, g := range b.lines[row] {
		if col >= Columns {
			break
		}
		if g.Width <= 1 {
			cells[col] = cell{g: g, start: true}
			col++
			continue
		}
		if col+1 >= Columns {
			break
		}
		cells[col] = cell{g: g, start: true}
		cells[col+1] = cell{cont: true, startCol: col}
		col += g.Width
	}

	clearWidePair := func(c int) {
		if c < 0 || c >= Columns {
			return
		}
		if cells[c].cont {
			s := cells[c].startCol
			cells[s] = cell{}
			if s+1 < Columns {
				cells[s+1] = cell{}
			}
			return
		}
		if cells[c].start && cells[c].g.Width > 1 {
			cells[c] = cell{}
			if c+1 < Columns && cells[c+1].cont {
				cells[c+1] = cell{}
			}
		}
	}

	for c := startCol; c < endCol; c++ {
		clearWidePair(c)
		cells[c] = cell{g: fill, start: true}
	}

	next := make([]Grapheme, 0, len(b.lines[row]))
	for c := 0; c < Columns; c++ {
		if cells[c].cont || !cells[c].start || cells[c].g.Text == "" {
			continue
		}
		next = append(next, cells[c].g)
		if cells[c].g.Width > 1 {
			c += cells[c].g.Width - 1
		}
	}
	b.lines[row] = next
	return nil
}

func (b *Buffer) LinesAsText() []string {
	out := make([]string, 0, Rows)
	for row := 0; row < Rows; row++ {
		var builder strings.Builder
		for _, g := range b.lines[row] {
			builder.WriteString(g.Text)
		}
		out = append(out, builder.String())
	}
	return trimTrailingEmptyLines(out)
}

func (b *Buffer) LinesAsANSIText() []string {
	out := make([]string, 0, Rows)
	for row := 0; row < Rows; row++ {
		var builder strings.Builder
		activeColor := ""
		activeBGColor := ""
		for _, g := range b.lines[row] {
			if g.Color != activeColor {
				if g.Color == "" {
					builder.WriteString("\x1b[39m")
					activeColor = ""
				} else if code, ok := ansiColorCode(g.Color); ok {
					builder.WriteString("\x1b[")
					builder.WriteString(code)
					builder.WriteString("m")
					activeColor = g.Color
				}
			}
			if g.BgColor != activeBGColor {
				if g.BgColor == "" {
					builder.WriteString("\x1b[49m")
					activeBGColor = ""
				} else if code, ok := ansiBGColorCode(g.BgColor); ok {
					builder.WriteString("\x1b[")
					builder.WriteString(code)
					builder.WriteString("m")
					activeBGColor = g.BgColor
				}
			}
			builder.WriteString(g.Text)
		}
		if activeColor != "" || activeBGColor != "" {
			builder.WriteString("\x1b[0m")
		}
		out = append(out, builder.String())
	}
	return trimTrailingEmptyLines(out)
}

func trimTrailingEmptyLines(lines []string) []string {
	lastNonEmpty := -1
	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] != "" {
			lastNonEmpty = i
			break
		}
	}
	if lastNonEmpty < 0 {
		return []string{}
	}
	return lines[:lastNonEmpty+1]
}

func ansiColorCode(name string) (string, bool) {
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

func ansiBGColorCode(name string) (string, bool) {
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

func (b *Buffer) SetLines(lines []string) {
	for row := 0; row < Rows; row++ {
		if row >= len(lines) {
			b.lines[row] = []Grapheme{}
			continue
		}
		rowSegments := b.eng.Segment(lines[row])
		fit := make([]Grapheme, 0, len(rowSegments))
		width := 0
		for _, seg := range rowSegments {
			if width+seg.Width > Columns {
				break
			}
			fit = append(fit, seg)
			width += seg.Width
		}
		b.lines[row] = fit
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
