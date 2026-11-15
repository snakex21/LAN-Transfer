package app

import (
	"testing"

	"lantransfer/internal/config"
	"lantransfer/internal/i18n"
)

func TestRequireCodePersistsAcrossRestartAndLanguageChange(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LANTRANSFER_DATA_DIR", dir)
	t.Setenv("LANTRANSFER_RECEIVE_DIR", t.TempDir())
	t.Setenv("LANTRANSFER_REQUIRE_CODE", "")
	defer i18n.SetCurrent(i18n.Fallback)

	state, err := NewState()
	if err != nil {
		t.Fatal(err)
	}
	if !state.Pairing.CodeRequired() || state.Info().NoCode {
		t.Fatal("a new installation must require the code")
	}
	if err := state.SetCodeRequired(false); err != nil {
		t.Fatal(err)
	}
	if state.Pairing.CodeRequired() || !state.Info().NoCode {
		t.Fatal("setting not applied at once (pairing / announced info)")
	}
	// Changing the language must keep the other settings.
	if _, err := ChooseLanguage("de"); err != nil {
		t.Fatal(err)
	}
	if config.LoadSettings().CodeRequired() {
		t.Fatal("changing the language reset require_code")
	}
	restarted, err := NewState()
	if err != nil {
		t.Fatal(err)
	}
	if restarted.Pairing.CodeRequired() || !restarted.Info().NoCode {
		t.Fatal("require_code=false not restored after a restart")
	}
	if err := restarted.SetCodeRequired(true); err != nil {
		t.Fatal(err)
	}
	again, err := NewState()
	if err != nil {
		t.Fatal(err)
	}
	if !again.Pairing.CodeRequired() {
		t.Fatal("require_code=true not restored after a restart")
	}
}
