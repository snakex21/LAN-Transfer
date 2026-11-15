package cli

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lantransfer/internal/app"
	"lantransfer/internal/config"
	"lantransfer/internal/i18n"
	"lantransfer/internal/service"
)

// newState creates an instance with its own data and receive folders.
func newState(t *testing.T) *app.State {
	t.Helper()
	t.Setenv("LANTRANSFER_DATA_DIR", t.TempDir())
	t.Setenv("LANTRANSFER_RECEIVE_DIR", t.TempDir())
	state, err := app.NewState()
	if err != nil {
		t.Fatal(err)
	}
	return state
}

// serve runs the receiver's HTTPS service on a free loopback port.
func serve(t *testing.T, state *app.State) string {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := service.NewServer(state)
	go server.ServeTLS(listener, "", "")
	t.Cleanup(func() { server.Close() })
	return listener.Addr().String()
}

// End to end through the CLI "send" command on 127.0.0.1: without --code the
// transfer works only while the receiver has the code switched off.
func TestCLISendWithoutCode(t *testing.T) {
	defer func(old bool) { config.Loopback = old }(config.Loopback)
	config.Loopback = true
	defer i18n.SetCurrent(i18n.Fallback)
	i18n.SetCurrent("en")
	t.Setenv("LANTRANSFER_CODE", "")
	t.Setenv("LANTRANSFER_REQUIRE_CODE", "")

	receiver := newState(t)
	address := serve(t, receiver)
	sender := newState(t)
	src := filepath.Join(t.TempDir(), "smoke.txt")
	if err := os.WriteFile(src, []byte("bez kodu / no code"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Code required (default): refused with a clear message, nothing stored.
	var out, errs strings.Builder
	err := run(sender, []string{"send", src, "--to", address}, &out, &errs)
	var e *i18n.Error
	if !errors.As(err, &e) || e.Key != "cli.noCode" {
		t.Fatalf("code required: got %v", err)
	}
	if _, err := os.Stat(filepath.Join(receiver.ReceiveDir, "smoke.txt")); err == nil {
		t.Fatal("file received although the code is required")
	}

	// The receiver switches the code off (saved like the window does).
	if err := receiver.SetCodeRequired(false); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errs.Reset()
	if err := run(sender, []string{"send", src, "--to", address, "--json"}, &out, &errs); err != nil {
		t.Fatalf("send without code: %v\n%s", err, errs.String())
	}
	if !strings.Contains(out.String(), `"ok":true`) {
		t.Fatalf("unexpected output: %s", out.String())
	}
	got, err := os.ReadFile(filepath.Join(receiver.ReceiveDir, "smoke.txt"))
	if err != nil || string(got) != "bez kodu / no code" {
		t.Fatalf("received %q, %v", got, err)
	}
}

func TestParseServeNoCode(t *testing.T) {
	if noCode, err := parseServe(nil); err != nil || noCode {
		t.Fatalf("serve: %v %v", noCode, err)
	}
	if noCode, err := parseServe([]string{"--no-code"}); err != nil || !noCode {
		t.Fatalf("serve --no-code: %v %v", noCode, err)
	}
	if _, err := parseServe([]string{"--bogus"}); err == nil {
		t.Fatal("unknown flag accepted")
	}
}
