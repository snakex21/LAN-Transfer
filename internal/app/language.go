package app

import (
	"os"

	"lantransfer/internal/config"
	"lantransfer/internal/i18n"
)

// Language sources.
const (
	LanguageAuto   = "auto"   // Windows display language
	LanguageManual = "manual" // chosen in the settings dialog
	LanguageEnv    = "env"    // LANTRANSFER_LANG (tests, screenshots, scripts)
)

// LanguageInfo is the resolved UI language.
type LanguageInfo struct {
	Language string `json:"language"`
	Source   string `json:"source"`
	Detected string `json:"detected"` // what "auto" resolves to
}

// ResolveLanguage picks the UI language: LANTRANSFER_LANG, then the choice
// saved in data/settings.json, then the Windows display language (the first
// supported one in the user's preferred list), then English.
func ResolveLanguage() LanguageInfo {
	return resolveLanguage(os.Getenv("LANTRANSFER_LANG"), config.LoadSettings(), i18n.SystemLanguages())
}

func resolveLanguage(env string, settings config.Settings, system []string) LanguageInfo {
	info := LanguageInfo{Detected: i18n.FromPreferred(system)}
	switch {
	case i18n.Normalize(env) != "":
		info.Language, info.Source = i18n.Normalize(env), LanguageEnv
	case settings.LanguageSource == LanguageManual && i18n.Normalize(settings.Language) != "":
		info.Language, info.Source = i18n.Normalize(settings.Language), LanguageManual
	default:
		info.Language, info.Source = info.Detected, LanguageAuto
	}
	return info
}

// ChooseLanguage saves the choice from the settings dialog ("auto" or a
// language code), makes it current and returns the result.
func ChooseLanguage(choice string) (LanguageInfo, error) {
	detected := i18n.FromPreferred(i18n.SystemLanguages())
	info := LanguageInfo{Language: detected, Source: LanguageAuto, Detected: detected}
	settings := config.Settings{LanguageSource: LanguageAuto}
	if choice != LanguageAuto {
		code := i18n.Normalize(choice)
		if code == "" {
			return LanguageInfo{}, i18n.New("err.unsupportedLanguage", "language", choice)
		}
		info.Language, info.Source = code, LanguageManual
		settings = config.Settings{Language: code, LanguageSource: LanguageManual}
	}
	if err := config.SaveSettings(settings); err != nil {
		return LanguageInfo{}, i18n.Wrap(err, "err.saveSettings", "error", err)
	}
	i18n.SetCurrent(info.Language)
	return info, nil
}
