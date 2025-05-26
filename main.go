//go:build windows

package main

import (
	fyneApp "fyne.io/fyne/v2/app"
	"github.com/gur22-09/net-spec/internals/app"
	"github.com/gur22-09/net-spec/internals/ui"
)

func main() {
	app, err := app.NewApplication()

	if err != nil {
		panic(err)
	}

	app.Logger.Println("app running")

	data, err := app.NetworkHandler.GetAllConnections()

	if err != nil {
		app.Logger.Fatal(err)
	}

	mainWIndow := ui.NewMainWindow(fyneApp.New(), data)

	mainWIndow.Window.ShowAndRun()
}
