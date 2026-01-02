package main

import (
	"embed"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logFile, logPath := setupLogging()
	if logFile != nil {
		defer logFile.Close()
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("panic: %v\n%s", recovered, debug.Stack())
		}
	}()
	log.Printf("PoolCensus starting (log: %s)", logPath)

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
		log.Printf("wails.Run error: %v", err)
	}
}

func setupLogging() (*os.File, string) {
	cacheDir, err := os.UserCacheDir()
	if err != nil || cacheDir == "" {
		cacheDir = "."
	}
	dir := filepath.Join(cacheDir, "PoolCensus")
	_ = os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, "poolcensus.log")

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.SetOutput(os.Stderr)
		return nil, path
	}

	log.SetOutput(io.MultiWriter(os.Stderr, file))
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	return file, path
}
