package editor

import "testing"

func TestBufferInsertBackspaceWithKorean(t *testing.T) {
	buf := NewBuffer(NewWidthEngine(true))
	col, err := buf.InsertText(0, 0, "ab")
	if err != nil {
		t.Fatalf("insert ascii failed: %v", err)
	}
	if col != 2 {
		t.Fatalf("unexpected col after ascii insert: %d", col)
	}

	col, err = buf.InsertText(0, col, "한")
	if err != nil {
		t.Fatalf("insert korean failed: %v", err)
	}
	if col != 4 {
		t.Fatalf("unexpected col after korean insert: %d", col)
	}

	col, err = buf.Backspace(0, col)
	if err != nil {
		t.Fatalf("backspace failed: %v", err)
	}
	if col != 2 {
		t.Fatalf("unexpected col after backspace: %d", col)
	}

	lines := buf.LinesAsText()
	if lines[0] != "ab" {
		t.Fatalf("line mismatch: %q", lines[0])
	}
}

func TestNewLineSplit(t *testing.T) {
	buf := NewBuffer(NewWidthEngine(true))
	col, err := buf.InsertText(0, 0, "abc한글")
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	if col != 7 {
		t.Fatalf("unexpected width col: %d", col)
	}
	nextRow, nextCol, err := buf.InsertNewLine(0, 3)
	if err != nil {
		t.Fatalf("newline failed: %v", err)
	}
	if nextRow != 1 || nextCol != 0 {
		t.Fatalf("unexpected cursor: (%d,%d)", nextRow, nextCol)
	}
	lines := buf.LinesAsText()
	if lines[0] != "abc" {
		t.Fatalf("line0 mismatch: %q", lines[0])
	}
	if lines[1] != "한글" {
		t.Fatalf("line1 mismatch: %q", lines[1])
	}
}
