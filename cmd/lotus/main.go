package main

import (
	"log"

	"github.com/nghialv/lotus/pkg/cli"
	"github.com/nghialv/lotus/pkg/app/lotus/cmd/controller"
	"github.com/nghialv/lotus/pkg/app/lotus/cmd/monitor"
)

func main() {
	app := cli.NewApp(
		"lotus",
		"Load testing tool.",
	)
	app.AddCommands(
		controller.NewCommand(),
		monitor.NewCommand(),
	)
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
