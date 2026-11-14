package ui

import (
	"embed"
	"encoding/json"
	"strings"

	"lantransfer/internal/i18n"
)

//go:embed assets/index.html assets/style.css assets/i18n.js assets/app.js
var assets embed.FS

// Bootstrap is the language state handed to the page before it renders.
type Bootstrap struct {
	Language  string                     `json:"language"`
	Source    string                     `json:"source"`   // auto | manual | env
	Detected  string                     `json:"detected"` // what "auto" resolves to
	Languages []i18n.Language            `json:"languages"`
	Catalogs  map[string]json.RawMessage `json:"catalogs"`
}

// NewBootstrap prepares the catalogs of lang and of the English fallback.
func NewBootstrap(lang, source, detected string) (Bootstrap, error) {
	boot := Bootstrap{Language: lang, Source: source, Detected: detected, Languages: i18n.Languages,
		Catalogs: map[string]json.RawMessage{}}
	for _, code := range []string{i18n.Fallback, lang} {
		raw, err := i18n.RawCatalog(code)
		if err != nil {
			return boot, err
		}
		boot.Catalogs[code] = raw
	}
	return boot, nil
}

// HTML returns the page with styles, scripts and the language bootstrap inlined.
func HTML(boot Bootstrap) (string, error) {
	html, err := assets.ReadFile("assets/index.html")
	if err != nil {
		return "", err
	}
	style, err := assets.ReadFile("assets/style.css")
	if err != nil {
		return "", err
	}
	i18nScript, err := assets.ReadFile("assets/i18n.js")
	if err != nil {
		return "", err
	}
	script, err := assets.ReadFile("assets/app.js")
	if err != nil {
		return "", err
	}
	// json.Marshal escapes <, > and & so the data cannot close the <script>.
	data, err := json.Marshal(boot)
	if err != nil {
		return "", err
	}
	lang, _ := i18n.Lookup(boot.Language)
	dir := "ltr"
	if lang.RTL {
		dir = "rtl"
	}

	result := strings.Replace(string(html), `<html lang="pl">`, `<html lang="`+boot.Language+`" dir="`+dir+`">`, 1)
	result = strings.Replace(result, "/*__STYLE__*/", string(style), 1)
	result = strings.Replace(result, "//__I18N__", "window.__LT_I18N = "+string(data)+";\n"+string(i18nScript), 1)
	result = strings.Replace(result, "//__SCRIPT__", string(script), 1)
	return result, nil
}

// HTMLAssets returns the raw page sources (key coverage tests).
func HTMLAssets() (html, script, i18nScript string) {
	h, _ := assets.ReadFile("assets/index.html")
	s, _ := assets.ReadFile("assets/app.js")
	i, _ := assets.ReadFile("assets/i18n.js")
	return string(h), string(s), string(i)
}
