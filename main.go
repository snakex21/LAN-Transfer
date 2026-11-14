// Windows file properties (version 2.0.0) and icon: rsrc_windows_amd64.syso is
// generated from winres/versioninfo.json; regenerate after a version bump.
//
//go:generate go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo@v1.4.1 -64 -o rsrc_windows_amd64.syso winres/versioninfo.json
package main

import (
	"fmt"
	"os"

	"lantransfer/internal/app"
	"lantransfer/internal/cli"
	"lantransfer/internal/desktop"
	"lantransfer/internal/networksetup"
)

func main() {
	if handled, err := networksetup.HandleSpecialMode(os.Args[1:]); handled {
		if err != nil {
			fail(err)
		}
		return
	}
	if len(os.Args) == 1 {
		networksetup.EnsureConfigured()
	}

	state, err := app.NewState()
	if err != nil {
		fail(err)
	}

	if len(os.Args) > 1 {
		if err := cli.Run(state, os.Args[1:]); err != nil {
			fail(err)
		}
		return
	}

	if err := desktop.Run(state); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
