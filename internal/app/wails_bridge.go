package app

import (
	"context"
	"path/filepath"

	"github.com/runableapp/simplescreeneditor/internal/platform"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Bridge struct {
	editor *EditorApp
	ctx    context.Context
}

type FileActionResult struct {
	State     State  `json:"state"`
	Path      string `json:"path"`
	Cancelled bool   `json:"cancelled"`
}

func NewBridge() *Bridge {
	return &Bridge{
		editor: New(platform.OSClipboard{}),
	}
}

func (b *Bridge) Startup(ctx context.Context) {
	b.ctx = ctx
}

func (b *Bridge) Snapshot() State {
	return b.editor.Snapshot()
}

func (b *Bridge) InsertText(input string) State {
	return b.editor.InsertText(input)
}

func (b *Bridge) Backspace() State {
	return b.editor.Backspace()
}

func (b *Bridge) Delete() State {
	return b.editor.Delete()
}

func (b *Bridge) ClearScreen() State {
	return b.editor.ClearScreen()
}

func (b *Bridge) Enter() State {
	return b.editor.Enter()
}

func (b *Bridge) MoveLeft() State {
	return b.editor.MoveLeft()
}

func (b *Bridge) MoveRight() State {
	return b.editor.MoveRight()
}

func (b *Bridge) MoveUp() State {
	return b.editor.MoveUp()
}

func (b *Bridge) MoveDown() State {
	return b.editor.MoveDown()
}

func (b *Bridge) MoveHome() State {
	return b.editor.MoveHome()
}

func (b *Bridge) MoveEnd() State {
	return b.editor.MoveEnd()
}

func (b *Bridge) MoveTop() State {
	return b.editor.MoveTop()
}

func (b *Bridge) MoveBottom() State {
	return b.editor.MoveBottom()
}

func (b *Bridge) SetCursor(row, col int) State {
	return b.editor.SetCursor(row, col)
}

func (b *Bridge) DrawLineLeft(style string) State {
	return b.editor.DrawLineLeft(style)
}

func (b *Bridge) DrawLineRight(style string) State {
	return b.editor.DrawLineRight(style)
}

func (b *Bridge) DrawLineUp(style string) State {
	return b.editor.DrawLineUp(style)
}

func (b *Bridge) DrawLineDown(style string) State {
	return b.editor.DrawLineDown(style)
}

func (b *Bridge) SetActiveANSIColor(color string) {
	b.editor.SetActiveANSIColor(color)
}

func (b *Bridge) SetActiveANSIFGColor(color string) {
	b.editor.SetActiveANSIFGColor(color)
}

func (b *Bridge) SetActiveANSIBGColor(color string) {
	b.editor.SetActiveANSIBGColor(color)
}

func (b *Bridge) ClearANSIColors() State {
	return b.editor.ClearANSIColors()
}

func (b *Bridge) SetRegionColor(startRow, startCol, endRow, endCol int, color string) State {
	return b.editor.SetRegionColor(startRow, startCol, endRow, endCol, color)
}

func (b *Bridge) SetRegionBGColor(startRow, startCol, endRow, endCol int, color string) State {
	return b.editor.SetRegionBGColor(startRow, startCol, endRow, endCol, color)
}

func (b *Bridge) SetRegionTextStyle(startRow, startCol, endRow, endCol int, style string) State {
	return b.editor.SetRegionTextStyle(startRow, startCol, endRow, endCol, style)
}

func (b *Bridge) FillRegion(startRow, startCol, endRow, endCol int, text string) State {
	return b.editor.FillRegion(startRow, startCol, endRow, endCol, text)
}

func (b *Bridge) ClearRegion(startRow, startCol, endRow, endCol int) State {
	return b.editor.ClearRegion(startRow, startCol, endRow, endCol)
}

func (b *Bridge) CopyBuffer() error {
	return b.editor.CopyBuffer()
}

func (b *Bridge) CopyRegion(startRow, startCol, endRow, endCol int) error {
	return b.editor.CopyRegion(startRow, startCol, endRow, endCol)
}

func (b *Bridge) PasteFromClipboard() State {
	return b.editor.PasteFromClipboard()
}

func (b *Bridge) SaveFile(path string) error {
	return b.editor.SaveFile(path)
}

func (b *Bridge) OpenFile(path string) (State, error) {
	return b.editor.OpenFile(path)
}

func (b *Bridge) OpenExternalURL(url string) {
	if url == "" {
		return
	}
	runtime.BrowserOpenURL(b.ctx, url)
}

func (b *Bridge) OpenFileDialog() (FileActionResult, error) {
	path, err := runtime.OpenFileDialog(b.ctx, runtime.OpenDialogOptions{
		Title: "Open text file",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text Files", Pattern: "*.txt;*.ans;*.asc"},
			{DisplayName: "All Files", Pattern: "*"},
		},
	})
	if err != nil {
		return FileActionResult{}, err
	}
	if path == "" {
		return FileActionResult{
			State:     b.editor.Snapshot(),
			Cancelled: true,
		}, nil
	}

	state, err := b.editor.OpenFile(path)
	if err != nil {
		return FileActionResult{}, err
	}
	return FileActionResult{
		State: state,
		Path:  path,
	}, nil
}

func (b *Bridge) Save() (FileActionResult, error) {
	path := b.editor.CurrentFilename()
	if path == "" {
		return b.SaveAsDialog()
	}
	if err := b.editor.SaveFile(path); err != nil {
		return FileActionResult{}, err
	}
	return FileActionResult{
		State: b.editor.Snapshot(),
		Path:  path,
	}, nil
}

func (b *Bridge) SaveAsDialog() (FileActionResult, error) {
	current := b.editor.CurrentFilename()
	defaultName := "untitled.txt"
	if current != "" {
		defaultName = filepath.Base(current)
	}

	path, err := runtime.SaveFileDialog(b.ctx, runtime.SaveDialogOptions{
		Title:           "Save text file as",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "Text Files", Pattern: "*.txt"},
			{DisplayName: "All Files", Pattern: "*"},
		},
	})
	if err != nil {
		return FileActionResult{}, err
	}
	if path == "" {
		return FileActionResult{
			State:     b.editor.Snapshot(),
			Cancelled: true,
		}, nil
	}
	if err := b.editor.SaveFileANSI(path); err != nil {
		return FileActionResult{}, err
	}
	return FileActionResult{
		State: b.editor.Snapshot(),
		Path:  path,
	}, nil
}
