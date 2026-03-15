package main

import (
	"embed"
	"fmt"
	"log"
	"strings"

	"github.com/runableapp/simplescreeneditor/internal/app"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

//go:embed VERSION
var appVersionText string

func main() {
	bridge := app.NewBridge()
	baseTitle := "Simple Screen Editor"
	version := strings.TrimSpace(appVersionText)
	windowTitle := baseTitle
	if version != "" {
		windowTitle = fmt.Sprintf("%s v%s", baseTitle, version)
	}
	err := wails.Run(&options.App{
		Title:     windowTitle,
		Width:     1320,
		Height:    800,
		MinWidth:  1320,
		MinHeight: 800,
		MaxWidth:  1320,
		MaxHeight: 800,
		DisableResize: true,
		OnStartup: bridge.Startup,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Linux: &linux.Options{
			Icon: appIcon,
		},
		Bind: []any{bridge},
	})
	if err != nil {
		log.Fatal(err)
	}
}
