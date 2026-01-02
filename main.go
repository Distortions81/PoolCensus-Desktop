package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewPoolCensusApp()

	err := wails.Run(&options.App{
		Title:            "PoolCensus",
		Width:            1200,
		Height:           780,
		MinWidth:         800,
		MinHeight:        600,
		BackgroundColour: &options.RGBA{R: 12, G: 17, B: 29, A: 1},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind:      []interface{}{app},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
