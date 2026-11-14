package cli

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseSendForms(t *testing.T) {
	t.Setenv("LANTRANSFER_CODE", "")
	opts, err := parseSend([]string{"a.txt", "b", "--to", "10.0.0.5", "--code", "123456", "--parallel", "8"})
	if err != nil || !reflect.DeepEqual(opts.paths, []string{"a.txt", "b"}) || opts.target != "10.0.0.5" || opts.parallel != 8 {
		t.Fatalf("%+v %v", opts, err)
	}
	opts, err = parseSend([]string{"plik.zip", "LAPTOP", `D:\Odbior`, "--code", "1"})
	if err != nil || opts.target != "LAPTOP" || opts.dest != `D:\Odbior` || len(opts.paths) != 1 {
		t.Fatalf("legacy form: %+v %v", opts, err)
	}
	if _, err := parseSend([]string{"a", "b"}); err == nil || !strings.Contains(err.Error(), "kodu") {
		t.Fatalf("missing code not reported: %v", err)
	}
	t.Setenv("LANTRANSFER_CODE", "999999")
	if opts, err := parseSend([]string{"a", "b"}); err != nil || opts.code != "999999" {
		t.Fatalf("env code: %+v %v", opts, err)
	}
	if _, err := parseSend([]string{"--to", "x", "--code", "1"}); err == nil {
		t.Fatal("no paths accepted")
	}
}

func TestVersionCommand(t *testing.T) {
	var out strings.Builder
	if err := run(nil, []string{"version"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "LAN Transfer ") || !strings.Contains(out.String(), "protokół 2") {
		t.Fatalf("unexpected version output: %q", out.String())
	}
}
