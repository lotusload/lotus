// Copyright (c) 2018 Lotus Load
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lotusload/lotus/pkg/log"
	"github.com/lotusload/lotus/pkg/version"
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
