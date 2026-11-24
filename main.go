package main

import (
	"embed"

	"MastermindHost/backend"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	host, err := backend.NewHost()
	if err != nil {
		println("Error:", err.Error())
		return
	}

	err = wails.Run(&options.App{
		Title:     "MastermindHost",
		Width:     1400,
		Height:    900,
		MinWidth:  1000,
		MinHeight: 650,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 5, G: 8, B: 6, A: 1},
		OnStartup:        host.Startup,
		OnShutdown:       host.OnShutdown,
		OnBeforeClose:    host.BeforeClose,
		Windows: &windows.Options{
			Theme: windows.Dark,
		},
		Bind: []interface{}{
			host,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
