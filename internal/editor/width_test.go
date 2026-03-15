package editor

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestMixedWidthFixtures(t *testing.T) {
	eng := NewWidthEngine(true)
	path := filepath.Join("..", "..", "testdata", "alignment", "mixed_cases.tsv")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) < 2 {
		t.Fatalf("fixture has no cases")
	}
	for _, line := range lines[1:] {
		fields := strings.Split(line, "\t")
		if len(fields) != 3 {
			t.Fatalf("invalid fixture line: %q", line)
		}
		label := fields[0]
		text := fields[1]
		expected, err := strconv.Atoi(fields[2])
		if err != nil {
			t.Fatalf("fixture expected width parse failed for %s: %v", label, err)
		}
		if got := eng.StringWidth(text); got != expected {
			t.Fatalf("%s width mismatch: got=%d expected=%d", label, got, expected)
		}
	}
}
