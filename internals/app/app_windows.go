package app

import (
	"log"
	"os"

	"github.com/gur22-09/net-spec/internals/network"
)

type Application struct {
	Logger         *log.Logger
	NetworkHandler *network.NetworkHandler
}

func NewApplication() (*Application, error) {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	networkHandler := network.NewNetworkHandler(logger)

	return &Application{
		NetworkHandler: networkHandler,
		Logger:         logger,
	}, nil
}
