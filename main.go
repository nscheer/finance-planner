package main

import (
	"embed"
	"log"
	"sync"
	"time"

	"finance-planner/planner"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// The built frontend (frontend/dist) is embedded into the binary.
//
//go:embed all:frontend/dist
var assets embed.FS

// Default window size: made for wide screens; the page itself caps its
// content width. The minimum keeps the table and the statistics side by side.
const (
	defaultWidth  = 1440
	defaultHeight = 900
	minWidth      = 1200
	minHeight     = 700
)

func main() {
	dataPath, err := planner.DefaultDataPath()
	if err != nil {
		log.Fatal(err)
	}
	service, err := planner.NewService(dataPath)
	if err != nil {
		log.Fatal(err)
	}

	app := application.New(application.Options{
		Name:        "Finance Planner",
		Description: "Plan monthly and yearly income and spendings",
		Services: []application.Service{
			application.NewServiceWithOptions(service, application.ServiceOptions{
				// Send coded errors as JSON so the frontend can translate them.
				MarshalError: planner.MarshalError,
			}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// Restore the last window geometry, if one was saved.
	opts := application.WebviewWindowOptions{
		Title:            "Finance Planner",
		Width:            defaultWidth,
		Height:           defaultHeight,
		MinWidth:         minWidth,
		MinHeight:        minHeight,
		BackgroundColour: application.NewRGB(245, 246, 250),
		URL:              "/",
	}
	if saved := service.GetState().Settings.Window; saved.Width >= minWidth && saved.Height >= minHeight {
		opts.Width, opts.Height = saved.Width, saved.Height
		opts.InitialPosition = application.WindowXY
		opts.X, opts.Y = saved.X, saved.Y
	}
	window := app.Window.NewWithOptions(opts)

	// Remember size and position. Resize/move events fire continuously, so
	// the write is debounced.
	var geometryTimer *time.Timer
	var geometryMu sync.Mutex
	remember := func(*application.WindowEvent) {
		geometryMu.Lock()
		defer geometryMu.Unlock()
		if geometryTimer != nil {
			geometryTimer.Stop()
		}
		geometryTimer = time.AfterFunc(500*time.Millisecond, func() {
			w, h := window.Size()
			x, y := window.Position()
			if w < minWidth || h < minHeight {
				return
			}
			if err := service.SetWindow(planner.WindowGeometry{Width: w, Height: h, X: x, Y: y}); err != nil {
				log.Println("saving window geometry:", err)
			}
		})
	}
	window.OnWindowEvent(events.Common.WindowDidResize, remember)
	window.OnWindowEvent(events.Common.WindowDidMove, remember)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
