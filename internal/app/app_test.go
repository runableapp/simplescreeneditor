package app

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/runableapp/simplescreeneditor/internal/platform"
)

func TestCopyPasteRoundTrip(t *testing.T) {
	clip := &platform.MemoryClipboard{}
	a := New(clip)
	a.InsertText("abc한글")

	if err := a.CopyBuffer(); err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	b := New(clip)
	state := b.PasteFromClipboard()
	if len(state.Lines[0]) == 0 {
		t.Fatalf("paste produced empty line")
	}
}

func TestOpenSaveUTF8(t *testing.T) {
	a := New(&platform.MemoryClipboard{})
	a.InsertText("오늘 news")
	path := filepath.Join(t.TempDir(), "sample.txt")

	if err := a.SaveFile(path); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if _, err := a.OpenFile(path); err != nil {
		t.Fatalf("open failed: %v", err)
	}
}

func TestMoveDownToEmptyLineSnapsColumnForTyping(t *testing.T) {
	a := New(&platform.MemoryClipboard{})
	a.InsertText("12345")
	a.MoveDown() // empty line
	state := a.Snapshot()
	if state.Cursor.Col != 0 {
		t.Fatalf("cursor should snap to col 0 on empty line, got %d", state.Cursor.Col)
	}
	state = a.InsertText(" ")
	if len(state.Lines[1]) == 0 {
		t.Fatalf("space insert on empty line failed after move down")
	}
}

func TestHomeEndMoveWithinCurrentLine(t *testing.T) {
	a := New(&platform.MemoryClipboard{})
	a.InsertText("ab한")

	state := a.MoveHome()
	if state.Cursor.Row != 0 || state.Cursor.Col != 0 {
		t.Fatalf("home should move to line start, got row=%d col=%d", state.Cursor.Row, state.Cursor.Col)
	}

	state = a.MoveEnd()
	if state.Cursor.Row != 0 || state.Cursor.Col != 4 {
		t.Fatalf("end should move to line end width, got row=%d col=%d", state.Cursor.Row, state.Cursor.Col)
	}
}

func TestCtrlStyleTopBottomDocumentMove(t *testing.T) {
	a := New(&platform.MemoryClipboard{})
	a.InsertText("top")
	a.MoveDown()
	a.InsertText("middle")
	a.MoveDown()
	a.InsertText("bottom")

	state := a.MoveTop()
	if state.Cursor.Row != 0 || state.Cursor.Col != 0 {
		t.Fatalf("top should move to beginning of document, got row=%d col=%d", state.Cursor.Row, state.Cursor.Col)
	}

	state = a.MoveBottom()
	if state.Cursor.Row != 24 || state.Cursor.Col != 0 {
		t.Fatalf("bottom should move to end of last row content, got row=%d col=%d", state.Cursor.Row, state.Cursor.Col)
	}
}

func TestLineDrawingCreatesCornersAndIntersections(t *testing.T) {
	a := New(&platform.MemoryClipboard{})

	a.DrawLineRight("single")
	a.DrawLineDown("single")
	a.DrawLineLeft("single")
	a.DrawLineUp("single")

	lines := a.buffer.LinesAsText()
	if lines[0] != "┌┐" {
		t.Fatalf("unexpected top line after loop corners: %q", lines[0])
	}
	if lines[1] != "└┘" {
		t.Fatalf("unexpected second line after loop corners: %q", lines[1])
	}
}

func TestLineDrawingTurnUsesCornerThenUpgradesWhenConnected(t *testing.T) {
	a := New(&platform.MemoryClipboard{})

	a.MoveDown()
	a.DrawLineRight("double")
	a.DrawLineUp("double")

	lines := a.buffer.LinesAsText()
	if lines[1] != "═╝" {
		t.Fatalf("turn should produce corner without phantom branch, got %q", lines[1])
	}

	// Add a real east connection later; the same cell should upgrade to a tee.
	a.MoveDown()
	a.DrawLineRight("double")
	lines = a.buffer.LinesAsText()
	if lines[1] != "═╩═" {
		t.Fatalf("corner should upgrade when east connection is added, got %q", lines[1])
	}
}

func TestRegionFillAndColor(t *testing.T) {
	a := New(&platform.MemoryClipboard{})
	a.FillRegion(0, 0, 0, 2, "X")
	lines := a.buffer.LinesAsText()
	if !strings.HasPrefix(lines[0], "XXX") {
		t.Fatalf("fill region should place text, got %q", lines[0])
	}

	a.SetRegionColor(0, 0, 0, 1, "red")
	tokens, err := a.buffer.RenderTokens(0)
	if err != nil {
		t.Fatalf("render tokens failed: %v", err)
	}
	if len(tokens) < 2 || tokens[0].Color != "red" || tokens[1].Color != "red" {
		t.Fatalf("set region color failed: %+v", tokens)
	}

	a.ClearRegion(0, 0, 0, 1)
	lines = a.buffer.LinesAsText()
	if !strings.HasPrefix(lines[0], "  X") {
		t.Fatalf("clear region should blank selected cells, got %q", lines[0])
	}
}

func TestCopyBufferIncludesANSICodes(t *testing.T) {
	clip := &platform.MemoryClipboard{}
	a := New(clip)
	a.SetActiveANSIColor("red")
	a.InsertText("A")
	a.SetActiveANSIColor("")
	a.InsertText("B")

	if err := a.CopyBuffer(); err != nil {
		t.Fatalf("copy failed: %v", err)
	}
	got, err := clip.ReadText()
	if err != nil {
		t.Fatalf("clipboard read failed: %v", err)
	}
	wantPrefix := "\x1b[31mA\x1b[39mB"
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("copy should include ANSI codes, got %q", got)
	}
}

func TestPasteFromClipboardParsesANSIColors(t *testing.T) {
	clip := &platform.MemoryClipboard{}
	a := New(clip)
	_ = clip.WriteText("\x1b[31mA\x1b[0mB")
	state := a.PasteFromClipboard()
	if len(state.Lines[0]) < 2 {
		t.Fatalf("expected two tokens after paste, got %+v", state.Lines[0])
	}
	if state.Lines[0][0].Text != "A" || state.Lines[0][0].Color != "red" {
		t.Fatalf("first token color parse failed: %+v", state.Lines[0][0])
	}
	if state.Lines[0][1].Text != "B" || state.Lines[0][1].Color != "" {
		t.Fatalf("reset color parse failed: %+v", state.Lines[0][1])
	}
}

func TestPasteFromClipboardOverwritesWithoutPushingLines(t *testing.T) {
	clip := &platform.MemoryClipboard{}
	a := New(clip)

	a.InsertText("AAAA")
	a.SetCursor(1, 0)
	a.InsertText("BBBB")

	a.SetCursor(0, 1)
	_ = clip.WriteText("Z\nY")
	a.PasteFromClipboard()

	lines := a.buffer.LinesAsText()
	if !strings.HasPrefix(lines[0], "AZAA") {
		t.Fatalf("row 0 should be overwritten in place, got %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "BYBB") {
		t.Fatalf("row 1 should be overwritten in place, got %q", lines[1])
	}
}

func TestCopyRegionUsesANSICodes(t *testing.T) {
	clip := &platform.MemoryClipboard{}
	a := New(clip)
	a.SetActiveANSIColor("green")
	a.FillRegion(0, 0, 0, 1, "X")
	if err := a.CopyRegion(0, 0, 0, 1); err != nil {
		t.Fatalf("copy region failed: %v", err)
	}
	got, err := clip.ReadText()
	if err != nil {
		t.Fatalf("clipboard read failed: %v", err)
	}
	if !strings.Contains(got, "\x1b[32m") {
		t.Fatalf("region copy should contain ANSI green code, got %q", got)
	}
}
