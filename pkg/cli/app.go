package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nghialv/lotus/pkg/log"
	"github.com/nghialv/lotus/pkg/version"
)

type App struct {
	rootCmd     *cobra.Command
	logLevel    string
	logEncoding string
}

func NewApp(name, desc string) *App {
	a := &App{
		rootCmd: &cobra.Command{
			Use:   name,
			Short: desc,
		},
		logLevel:    log.DefaultLevel,
		logEncoding: log.DefaultEncoding,
	}
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the information of current binary",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Get())
		},
	}
	a.rootCmd.AddCommand(versionCmd)
	a.setGlobalFlags()
	return a
}

func (a *App) AddCommands(cmds ...*cobra.Command) {
	for _, cmd := range cmds {
		a.rootCmd.AddCommand(cmd)
	}
}

func (a *App) Run() error {
	return a.rootCmd.Execute()
}

func (a *App) setGlobalFlags() {
	a.rootCmd.PersistentFlags().StringVar(&a.logLevel, "log-level", a.logLevel, "The minimum enabled logging level")
	a.rootCmd.PersistentFlags().StringVar(&a.logEncoding, "log-encoding", a.logEncoding, "The encoding type for logger [json|console]")
}
