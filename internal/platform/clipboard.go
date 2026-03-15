// Package platform provides OS integration helpers.
package platform

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
)

type Clipboard interface {
	ReadText() (string, error)
	WriteText(text string) error
}

type OSClipboard struct{}

func (OSClipboard) ReadText() (string, error) {
	if clipboardMode() == "wayland" {
		out, err := exec.Command("wl-paste", "--no-newline").Output()
		if err == nil {
			return string(out), nil
		}
	}
	return clipboard.ReadAll()
}

func (OSClipboard) WriteText(text string) error {
	if clipboardMode() == "wayland" {
		cmd := exec.Command("wl-copy")
		cmd.Stdin = bytes.NewBufferString(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	return clipboard.WriteAll(text)
}

func clipboardMode() string {
	// Only Linux has the Wayland/X11 split we care about.
	if runtime.GOOS != "linux" {
		return "default"
	}

	sessionType := strings.ToLower(strings.TrimSpace(os.Getenv("XDG_SESSION_TYPE")))
	isWaylandSession := os.Getenv("WAYLAND_DISPLAY") != "" || sessionType == "wayland"
	if !isWaylandSession {
		// Includes X11 and unknown Linux sessions; use normal path.
		return "default"
	}
	if _, err := exec.LookPath("wl-copy"); err != nil {
		return "default"
	}
	if _, err := exec.LookPath("wl-paste"); err != nil {
		return "default"
	}
	return "wayland"
}

type MemoryClipboard struct {
	data string
}

func (m *MemoryClipboard) ReadText() (string, error) {
	return m.data, nil
}

func (m *MemoryClipboard) WriteText(text string) error {
	m.data = text
	return nil
}
