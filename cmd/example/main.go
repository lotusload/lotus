package main

import (
	"log"

	"github.com/nghialv/lotus/pkg/cli"
	"github.com/nghialv/lotus/pkg/app/example/cmd/helloworld"
	"github.com/nghialv/lotus/pkg/app/example/cmd/simplegrpc"
	"github.com/nghialv/lotus/pkg/app/example/cmd/simplehttp"
	"github.com/nghialv/lotus/pkg/app/example/cmd/threesteps"
	"github.com/nghialv/lotus/pkg/app/example/cmd/virtualuser"
)

func main() {
	app := cli.NewApp(
		"lotus-example",
		"Example of using lotus.",
	)
	app.AddCommands(
		simplehttp.NewCommand(),
		simplegrpc.NewCommand(),
		threesteps.NewCommand(),
		virtualuser.NewCommand(),
		helloworld.NewCommand(),
	)
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
