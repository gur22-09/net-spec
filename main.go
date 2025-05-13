//go:build windows

package main

import "github.com/gur22-09/net-spec/internals/app"

func main() {
	app, err := app.NewApplication()

	if err != nil {
		panic(err)
	}

	app.Logger.Println("app running")

	err = app.NetworkHandler.SetupNetworkListener()

	if err != nil {
		app.Logger.Fatal(err)
	}
}
