// Package i18n translates user-facing text of the window, the CLI and the
// messages the Go backend sends to the UI.
//
// Messages are identified by stable keys ("err.wrongCode") with named
// parameters ("{name}", "{need:size}"). The backend passes key + parameters to
// the UI (Msg, *Error), which renders them in the language chosen there; the
// CLI and log output render them with Render in the current language.
// Polish (ui.pl.json) is the source language, English the fallback.
package i18n

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"lantransfer/locales"
)

// Source and Fallback language codes.
const (
	Source   = "pl"
	Fallback = "en"
)

// Language describes one supported UI language.
type Language struct {
	Code   string `json:"code"`   // file suffix: ui.<code>.json
	Name   string `json:"name"`   // native name for the language picker
	Locale string `json:"locale"` // BCP 47 tag for Intl (number, date, plural rules)
	RTL    bool   `json:"rtl"`
}

// Languages are the 30 UI languages, the same set as the README translations.
var Languages = []Language{
	{"en", "English", "en-US", false},
	{"pl", "Polski", "pl-PL", false},
	{"de", "Deutsch", "de-DE", false},
	{"fr", "Français", "fr-FR", false},
	{"es", "Español", "es-ES", false},
	{"it", "Italiano", "it-IT", false},
	{"pt", "Português", "pt-PT", false},
	{"nl", "Nederlands", "nl-NL", false},
	{"sv", "Svenska", "sv-SE", false},
	{"no", "Norsk", "nb-NO", false},
	{"da", "Dansk", "da-DK", false},
	{"fi", "Suomi", "fi-FI", false},
	{"cs", "Čeština", "cs-CZ", false},
	{"sk", "Slovenčina", "sk-SK", false},
	{"hu", "Magyar", "hu-HU", false},
	{"ro", "Română", "ro-RO", false},
	{"tr", "Türkçe", "tr-TR", false},
	{"id", "Bahasa Indonesia", "id-ID", false},
	{"ms", "Bahasa Melayu", "ms-MY", false},
	{"vi", "Tiếng Việt", "vi-VN", false},
	{"ru", "Русский", "ru-RU", false},
	{"uk", "Українська", "uk-UA", false},
	{"bg", "Български", "bg-BG", false},
	{"el", "Ελληνικά", "el-GR", false},
	{"ar", "العربية", "ar-u-nu-latn", true},
	{"he", "עברית", "he-IL", true},
	{"hi", "हिन्दी", "hi-IN", false},
	{"zh", "中文", "zh-CN", false},
	{"ja", "日本語", "ja-JP", false},
	{"ko", "한국어", "ko-KR", false},
}

// Lookup returns the language with this code.
func Lookup(code string) (Language, bool) {
	for _, l := range Languages {
		if l.Code == code {
			return l, true
		}
	}
	return Language{}, false
}

// aliases maps language subtags that differ from our file codes.
var aliases = map[string]string{"nb": "no", "nn": "no", "iw": "he", "in": "id", "zsm": "ms"}

// Normalize turns a locale tag ("de-AT", "pt_BR", "nb-NO", "zh-Hans-CN",
// "en_US.UTF-8") into a supported language code, or "" when unsupported.
func Normalize(tag string) string {
	tag = strings.ToLower(strings.TrimSpace(tag))
	if i := strings.IndexAny(tag, ".@"); i >= 0 {
		tag = tag[:i]
	}
	if i := strings.IndexAny(tag, "-_"); i >= 0 {
		tag = tag[:i]
	}
	if alias, ok := aliases[tag]; ok {
		tag = alias
	}
	if _, ok := Lookup(tag); !ok {
		return ""
	}
	return tag
}

// FromPreferred returns the first supported language of an ordered list of
// locale tags (most preferred first), or Fallback.
func FromPreferred(tags []string) string {
	for _, tag := range tags {
		if code := Normalize(tag); code != "" {
			return code
		}
	}
	return Fallback
}

var current atomic.Value // string

// SetCurrent selects the language used by Error() and the CLI.
func SetCurrent(code string) {
	if _, ok := Lookup(code); !ok {
		code = Fallback
	}
	current.Store(code)
}

// Current returns the active language (Fallback until SetCurrent is called).
func Current() string {
	if code, ok := current.Load().(string); ok {
		return code
	}
	return Fallback
}

// ---- catalogs --------------------------------------------------------------

type entry struct {
	text   string
	plural map[string]string
}

var (
	loadOnce sync.Once
	catalogs map[string]map[string]entry
	loadErr  error
)

func load() {
	catalogs = map[string]map[string]entry{}
	for _, l := range Languages {
		raw, err := RawCatalog(l.Code)
		if err != nil {
			loadErr = err
			continue
		}
		var values map[string]json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			loadErr = fmt.Errorf("ui.%s.json: %w", l.Code, err)
			continue
		}
		cat := make(map[string]entry, len(values))
		for key, value := range values {
			var e entry
			if err := json.Unmarshal(value, &e.text); err != nil {
				if err := json.Unmarshal(value, &e.plural); err != nil {
					loadErr = fmt.Errorf("ui.%s.json: %s: %w", l.Code, key, err)
					continue
				}
			}
			cat[key] = e
		}
		catalogs[l.Code] = cat
	}
}

// RawCatalog returns the JSON file of a language as embedded in the binary.
func RawCatalog(code string) ([]byte, error) {
	if _, ok := Lookup(code); !ok {
		return nil, fmt.Errorf("unsupported language %q", code)
	}
	return locales.FS.ReadFile("ui." + code + ".json")
}

// LoadError reports a problem with the embedded catalogs (tests).
func LoadError() error {
	loadOnce.Do(load)
	return loadErr
}

// Keys returns every key of a language catalog (tests).
func Keys(code string) map[string]bool {
	loadOnce.Do(load)
	keys := map[string]bool{}
	for key := range catalogs[code] {
		keys[key] = true
	}
	return keys
}

func lookup(lang, key string) (entry, string, bool) {
	loadOnce.Do(load)
	if e, ok := catalogs[lang][key]; ok {
		return e, lang, true
	}
	if e, ok := catalogs[Fallback][key]; ok {
		return e, Fallback, true
	}
	return entry{}, "", false
}

// ---- messages --------------------------------------------------------------

// Params are the named values of a message. A value is a string, a number,
// a nested Msg or, while building, an error (converted by P).
type Params map[string]any

// Msg is a translatable message: a key plus parameters. An empty Key means
// untranslatable text (an OS error, a peer's free text) in Params["text"].
type Msg struct {
	Key    string `json:"key"`
	Params Params `json:"params,omitempty"`
}

// M builds a message from a key and name/value pairs.
func M(key string, kv ...any) Msg { return Msg{Key: key, Params: P(kv...)} }

// Text wraps untranslatable text as a message.
func Text(s string) Msg { return Msg{Params: Params{"text": s}} }

// P builds Params from name/value pairs. Errors become nested messages
// (translatable ones) or their text, so that Params always marshal to JSON.
func P(kv ...any) Params {
	if len(kv) == 0 {
		return nil
	}
	params := make(Params, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		name := fmt.Sprint(kv[i])
		switch v := kv[i+1].(type) {
		case error:
			params[name] = MsgOf(v)
		case *Msg:
			params[name] = *v
		case fmt.Stringer:
			params[name] = v.String()
		default:
			params[name] = v
		}
	}
	return params
}

// Localized is implemented by errors that carry a translatable message.
type Localized interface {
	error
	Message() Msg
}

// MsgOf returns the translatable message of err (the outermost Localized in
// its chain) or err's text.
func MsgOf(err error) Msg {
	var loc Localized
	if AsLocalized(err, &loc) {
		return loc.Message()
	}
	return Text(err.Error())
}

// Payload is MsgOf for JSON results: nil when err is nil.
func Payload(err error) *Msg {
	if err == nil {
		return nil
	}
	m := MsgOf(err)
	return &m
}

// Error is an error with a translation key. Error() renders it in the
// current language; the UI renders Message() in its own language.
type Error struct {
	Key    string
	Params Params
	Err    error // wrapped error (errors.Is/As), not shown
}

// New creates a translatable error.
func New(key string, kv ...any) *Error { return &Error{Key: key, Params: P(kv...)} }

// Wrap creates a translatable error that wraps err (errors.Is(result, err)).
// Pass err again as a parameter when its text should be shown.
func Wrap(err error, key string, kv ...any) *Error {
	return &Error{Key: key, Params: P(kv...), Err: err}
}

func (e *Error) Error() string  { return Render(Current(), e.Message()) }
func (e *Error) Unwrap() error  { return e.Err }
func (e *Error) Message() Msg   { return Msg{Key: e.Key, Params: e.Params} }
func (e *Error) String() string { return e.Error() }

// ---- rendering ---------------------------------------------------------------

var placeholder = regexp.MustCompile(`\{([A-Za-z0-9_]+)(?::([a-z]+))?\}`)

// T renders key with name/value pairs in lang.
func T(lang, key string, kv ...any) string { return Render(lang, M(key, kv...)) }

// Render renders a message in lang (falling back to English, then to the key).
func Render(lang string, m Msg) string {
	if m.Key == "" {
		return fmt.Sprint(m.Params["text"])
	}
	e, from, ok := lookup(lang, m.Key)
	if !ok {
		return m.Key
	}
	template := e.text
	if e.plural != nil {
		n, _ := number(m.Params["count"])
		form, ok := e.plural[PluralCategory(from, n)]
		if !ok {
			form = e.plural["other"]
		}
		template = form
	}
	return placeholder.ReplaceAllStringFunc(template, func(token string) string {
		parts := placeholder.FindStringSubmatch(token)
		value, ok := m.Params[parts[1]]
		if !ok {
			return token
		}
		return formatValue(lang, value, parts[2])
	})
}

func formatValue(lang string, value any, kind string) string {
	switch v := value.(type) {
	case Msg:
		return Render(lang, v)
	case map[string]any: // a nested Msg decoded from JSON (remote errors)
		return Render(lang, msgFromMap(v))
	case string:
		return v
	}
	n, ok := number(value)
	if !ok {
		return fmt.Sprint(value)
	}
	switch kind {
	case "size":
		return FormatSize(lang, int64(n))
	case "speed":
		return FormatSpeed(lang, n)
	case "duration":
		return FormatDuration(lang, n)
	case "seconds":
		return FormatUnit(lang, "second", n, 1)
	case "percent":
		return FormatPercent(lang, n)
	}
	if n == math.Trunc(n) {
		return FormatNumber(lang, n, 0)
	}
	return FormatNumber(lang, n, 2)
}

func msgFromMap(v map[string]any) Msg {
	m := Msg{}
	m.Key, _ = v["key"].(string)
	if params, ok := v["params"].(map[string]any); ok {
		m.Params = Params(params)
	}
	return m
}

func number(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	}
	return 0, false
}
