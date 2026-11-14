package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"lantransfer/internal/app"
	"lantransfer/internal/discovery"
	"lantransfer/internal/service"
	"lantransfer/internal/transfer"
)

func Run(state *app.State, args []string) error {
	if len(args) == 0 {
		return errors.New("brak komendy")
	}

	switch args[0] {
	case "serve":
		return service.Run(state)
	case "devices":
		devices, err := discovery.Discover(1200*time.Millisecond, state.ID)
		if err != nil {
			return err
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(devices)
	case "send":
		if len(args) < 3 || len(args) > 4 {
			return errors.New("użycie: LANTransferCLI.exe send <plik|folder> <ip-lub-nazwa> [folder-docelowy]")
		}
		destination := ""
		if len(args) == 4 {
			destination = args[3]
		}
		return transfer.SendPathTo(args[1], args[2], state.ID, destination)
	default:
		return fmt.Errorf("nieznana komenda: %s", args[0])
	}
}
