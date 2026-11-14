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
	"lantransfer/internal/i18n"
	"lantransfer/internal/protocol"
	"lantransfer/internal/security"
	"lantransfer/internal/service"
	"lantransfer/internal/transfer"
	"lantransfer/internal/version"
)

// usage is the help text in the current language (cli.usage in locales/).
func usage() string { return t("cli.usage") }

// t renders a message in the current language (LANTRANSFER_LANG, the saved
// choice or the system language, see app.ResolveLanguage).
func t(key string, kv ...any) string { return i18n.T(i18n.Current(), key, kv...) }

// usageError is an error followed by the help text.
func usageError(err error) error {
	return i18n.Wrap(err, "cli.errorWithUsage", "error", err, "usage", i18n.M("cli.usage"))
}

// Run executes one CLI command.
func Run(state *app.State, args []string) error {
	return run(state, args, os.Stdout, os.Stderr)
}

func run(state *app.State, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New(usage())
	}
	switch args[0] {
	case "serve":
		fmt.Fprintln(stderr, t("cli.serving", "name", state.Name, "port", fmt.Sprint(config.HTTPPort)))
		fmt.Fprintln(stderr, t("cli.receiveDir", "path", state.ReceiveDir))
		fmt.Fprintln(stderr, t("cli.fingerprint", "fingerprint", security.ShortFingerprint(state.Identity.Fingerprint)))
		fmt.Fprintln(stderr, t("cli.pairingCode", "code", state.Pairing.Code()))
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
		fmt.Fprintln(stdout, t("cli.version", "version", version.String(), "protocol", fmt.Sprint(protocol.Version)))
		return nil
	case "help", "-h", "--help":
		fmt.Fprintln(stdout, usage())
		return nil
	default:
		return usageError(i18n.New("cli.unknownCommand", "command", args[0]))
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
			return opts, usageError(err)
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
			return opts, errors.New(usage())
		}
		opts.paths, opts.target = positional[:1], positional[1]
		if len(positional) == 3 {
			opts.dest = positional[2]
		}
	} else {
		opts.paths = positional
	}
	if len(opts.paths) == 0 {
		return opts, usageError(i18n.New("cli.noFiles"))
	}
	if strings.TrimSpace(opts.code) == "" {
		return opts, i18n.New("cli.noCode")
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
	fmt.Fprintln(stderr, t("cli.paired", "peer", session.PeerName, "address", address, "fingerprint", security.ShortFingerprint(session.Fingerprint)))

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
			fmt.Fprintln(stderr, "\r"+t("cli.sent", "name", plan.Name, "count", plan.TotalFiles, "size", plan.TotalBytes,
				"seconds", elapsed.Seconds(), "speed", result["mb_per_s"].(float64)*(1<<20)))
		}
		for _, skipped := range plan.Skipped {
			fmt.Fprintln(stderr, t("cli.skipped", "path", skipped))
		}
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return i18n.New("cli.interrupted")
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
			fmt.Fprint(w, "\r"+t("cli.progress", "name", plan.Name, "percent", pct, "done", bytes, "total", plan.TotalBytes,
				"speed", rate, "doneFiles", fmt.Sprint(progress.DoneFiles.Load()), "totalFiles", fmt.Sprint(plan.TotalFiles))+"   ")
		}
	}
}
