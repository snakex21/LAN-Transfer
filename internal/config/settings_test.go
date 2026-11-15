package config

import "testing"

func TestRequireCodeSettingPersists(t *testing.T) {
	t.Setenv("LANTRANSFER_DATA_DIR", t.TempDir())
	t.Setenv("LANTRANSFER_REQUIRE_CODE", "")
	if !LoadSettings().CodeRequired() {
		t.Fatal("the code must be required by default (no settings file)")
	}
	off := false
	if err := SaveSettings(Settings{Language: "de", LanguageSource: "manual", RequireCode: &off}); err != nil {
		t.Fatal(err)
	}
	loaded := LoadSettings()
	if loaded.CodeRequired() || loaded.Language != "de" {
		t.Fatalf("saved settings not read back: %+v", loaded)
	}
	t.Setenv("LANTRANSFER_REQUIRE_CODE", "1")
	if !LoadSettings().CodeRequired() {
		t.Fatal("LANTRANSFER_REQUIRE_CODE=1 must override the file")
	}
	t.Setenv("LANTRANSFER_REQUIRE_CODE", "0")
	if (Settings{}).CodeRequired() {
		t.Fatal("LANTRANSFER_REQUIRE_CODE=0 must override the default")
	}
}
