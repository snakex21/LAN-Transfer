package cli

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"lantransfer/internal/i18n"
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
	var e *i18n.Error
	if _, err := parseSend([]string{"a", "b"}); !errors.As(err, &e) || e.Key != "cli.noCode" {
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
	defer i18n.SetCurrent(i18n.Fallback)
	for lang, want := range map[string]string{"en": "(protocol 2)", "pl": "(protokół 2)", "de": "(Protokoll 2)"} {
		i18n.SetCurrent(lang)
		var out strings.Builder
		if err := run(nil, []string{"version"}, &out, &out); err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(out.String(), "LAN Transfer ") || !strings.Contains(out.String(), want) {
			t.Errorf("%s: unexpected version output: %q", lang, out.String())
		}
	}
}

func TestHelpIsTranslatedButKeepsCommands(t *testing.T) {
	defer i18n.SetCurrent(i18n.Fallback)
	for _, lang := range i18n.Languages {
		i18n.SetCurrent(lang.Code)
		var out strings.Builder
		if err := run(nil, []string{"help"}, &out, &out); err != nil {
			t.Fatal(err)
		}
		for _, literal := range []string{"LANTransferCLI.exe send", "--to", "--code", "--dest", "--parallel", "--json", "LANTRANSFER_CODE", "LANTRANSFER_LANG"} {
			if !strings.Contains(out.String(), literal) {
				t.Errorf("%s: help lacks %q", lang.Code, literal)
			}
		}
	}
}
