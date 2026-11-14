// Package locales embeds the UI translations (ui.<code>.json).
//
// ui.pl.json is the source language, ui.en.json is the fallback for missing
// keys. Every file has exactly the same keys (tools/check_i18n.py, CI).
// A value is either a string or, for counts, an object keyed by the CLDR
// plural category of the language (zero/one/two/few/many/other).
package locales

import "embed"

// FS holds every ui.<code>.json file.
//
//go:embed ui.*.json
var FS embed.FS
