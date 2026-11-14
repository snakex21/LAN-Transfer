// Package cli implements LANTransferCLI.exe commands for scripts and agents.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"lantransfer/internal/app"
	"lantransfer/internal/config"
	"lantransfer/internal/discovery"
	"lantransfer/internal/protocol"
	"lantransfer/internal/security"
	"lantransfer/internal/service"
	"lantransfer/internal/transfer"
	"lantransfer/internal/version"
)

const usage = `użycie:
  LANTransferCLI.exe version
  LANTransferCLI.exe serve
  LANTransferCLI.exe devices
  LANTransferCLI.exe send <plik|folder>... --to <ip[:port]|nazwa> --code <kod> [--dest <folder>] [--parallel N] [--json]
  LANTransferCLI.exe send <plik|folder> <ip-lub-nazwa> [folder-docelowy] --code <kod>

Kod parowania wyświetla odbiorca (okno programu albo "serve").
Zamiast --code można ustawić zmienną LANTRANSFER_CODE.`

// Run executes one CLI command.
func Run(state *app.State, args []string) error {
	return run(state, args, os.Stdout, os.Stderr)
}

func run(state *app.State, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New(usage)
	}
	switch args[0] {
	case "serve":
		fmt.Fprintf(stderr, "LAN Transfer odbiera pliki jako %q na porcie %d\n", state.Name, config.HTTPPort)
		fmt.Fprintf(stderr, "Folder odbioru: %s\n", state.ReceiveDir)
		fmt.Fprintf(stderr, "Odcisk certyfikatu: %s\n", security.ShortFingerprint(state.Identity.Fingerprint))
		fmt.Fprintf(stderr, "Kod parowania: %s\n", state.Pairing.Code())
		return service.Run(state)
	case "devices":
		devices, err := discovery.Discover(1200*time.Millisecond, state.ID)
		if err != nil {
			return err
		}
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(devices)
	case "send":
		return send(state, args[1:], stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "LAN Transfer %s (protokół %d)\n", version.String(), protocol.Version)
		return nil
	case "help", "-h", "--help":
		fmt.Fprintln(stdout, usage)
		return nil
	default:
		return fmt.Errorf("nieznana komenda: %s\n%s", args[0], usage)
	}
}

type sendOptions struct {
	paths    []string
	target   string
	dest     string
	code     string
	parallel int
	json     bool
}

func parseSend(args []string) (sendOptions, error) {
	opts := sendOptions{code: os.Getenv("LANTRANSFER_CODE")}
	fs := flag.NewFlagSet("send", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&opts.target, "to", "", "")
	fs.StringVar(&opts.dest, "dest", "", "")
	fs.StringVar(&opts.code, "code", opts.code, "")
	fs.IntVar(&opts.parallel, "parallel", transfer.DefaultParallel, "")
	fs.BoolVar(&opts.json, "json", false, "")

	// Allow flags anywhere: collect positionals between flag groups.
	var positional []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return opts, fmt.Errorf("%v\n%s", err, usage)
		}
		args = fs.Args()
		if len(args) > 0 {
			positional = append(positional, args[0])
			args = args[1:]
		}
	}
	if opts.target == "" {
		// Legacy form: send <path> <target> [destination]
		if len(positional) < 2 || len(positional) > 3 {
			return opts, errors.New(usage)
		}
		opts.paths, opts.target = positional[:1], positional[1]
		if len(positional) == 3 {
			opts.dest = positional[2]
		}
	} else {
		opts.paths = positional
	}
	if len(opts.paths) == 0 {
		return opts, errors.New("nie podano plików do wysłania\n" + usage)
	}
	if strings.TrimSpace(opts.code) == "" {
		return opts, errors.New("brak kodu parowania: podaj --code albo ustaw LANTRANSFER_CODE")
	}
	return opts, nil
}

func send(state *app.State, args []string, stdout, stderr io.Writer) error {
	opts, err := parseSend(args)
	if err != nil {
		return err
	}
	address, err := transfer.ResolveTarget(opts.target, state.ID)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	session, err := transfer.Pair(ctx, address, opts.code, state.Info())
	if err != nil {
		return err
	}
	defer session.Close()
	fmt.Fprintf(stderr, "Sparowano z %s (%s), certyfikat %s\n", session.PeerName, address, security.ShortFingerprint(session.Fingerprint))

	for _, path := range opts.paths {
		plan, err := transfer.PlanPath(path, address, opts.dest)
		if err != nil {
			return err
		}
		progress := &transfer.Progress{}
		started := time.Now()
		done := make(chan struct{})
		go reportProgress(stderr, plan, progress, started, done)
		err = session.Send(ctx, plan, opts.dest, opts.parallel, progress)
		close(done)
		elapsed := time.Since(started)
		result := map[string]any{
			"name": plan.Name, "files": plan.TotalFiles, "bytes": plan.TotalBytes,
			"resumed_bytes": progress.Resumed.Load(), "seconds": elapsed.Seconds(),
			"mb_per_s": float64(plan.TotalBytes-progress.Resumed.Load()) / (1 << 20) / max(elapsed.Seconds(), 1e-9),
			"ok":       err == nil,
		}
		if err != nil {
			result["error"] = err.Error()
		}
		if opts.json {
			_ = json.NewEncoder(stdout).Encode(result)
		} else if err == nil {
			fmt.Fprintf(stderr, "\rWysłano %s: %d plików, %s w %.1f s (%.0f MB/s), SHA-256 zgodne\n",
				plan.Name, plan.TotalFiles, transfer.FormatBytes(plan.TotalBytes), elapsed.Seconds(), result["mb_per_s"])
		}
		for _, skipped := range plan.Skipped {
			fmt.Fprintf(stderr, "Pominięto (dowiązanie lub plik specjalny): %s\n", skipped)
		}
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return errors.New("przerwano – uruchom tę samą komendę ponownie, aby wznowić")
			}
			return err
		}
	}
	return nil
}

func reportProgress(w io.Writer, plan *transfer.Plan, progress *transfer.Progress, started time.Time, done <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			bytes := progress.DoneBytes.Load()
			pct := 100.0
			if plan.TotalBytes > 0 {
				pct = float64(bytes) * 100 / float64(plan.TotalBytes)
			}
			rate := float64(bytes-progress.Resumed.Load()) / time.Since(started).Seconds()
			fmt.Fprintf(w, "\r%s: %5.1f%%  %s / %s  %.0f MB/s  pliki %d/%d   ", plan.Name, pct,
				transfer.FormatBytes(bytes), transfer.FormatBytes(plan.TotalBytes), rate/(1<<20),
				progress.DoneFiles.Load(), plan.TotalFiles)
		}
	}
}
