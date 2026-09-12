package main

import (
	"embed"
	"log"

	"finance-planner/planner"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// The built frontend (frontend/dist) is embedded into the binary.
//
//go:embed all:frontend/dist
var assets embed.FS

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

	// Sized for wide screens; the page itself caps its content width so it
	// stays readable on very wide monitors. The minimum keeps the table and
	// the statistics box side by side.
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Finance Planner",
		Width:            1440,
		Height:           900,
		MinWidth:         1100,
		MinHeight:        700,
		BackgroundColour: application.NewRGB(245, 246, 250),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
