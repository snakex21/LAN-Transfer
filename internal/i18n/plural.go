package i18n

import (
	"errors"
	"math"
)

// AsLocalized is errors.As for the Localized interface.
func AsLocalized(err error, target *Localized) bool { return errors.As(err, target) }

// PluralCategory returns the CLDR cardinal plural category of n in lang, the
// same value as Intl.PluralRules(locale).select(n) in the UI. Counts are
// integers; a fractional n falls into "other" (or "many" where CLDR says so).
// testdata/plurals.json (generated from Intl) checks every language.
func PluralCategory(lang string, n float64) string {
	if n < 0 {
		n = -n
	}
	integer := n == math.Trunc(n)
	i := int64(n)
	mod10, mod100 := i%10, i%100
	switch lang {
	case "id", "ms", "vi", "zh", "ja", "ko":
		return "other"
	case "en", "de", "nl", "sv", "no", "fi", "el", "hu", "tr", "bg":
		if integer && i == 1 {
			return "one"
		}
	case "da":
		if i == 1 || (!integer && i == 0) {
			return "one"
		}
	case "hi":
		if i == 0 || n == 1 {
			return "one"
		}
	case "he":
		switch {
		case integer && i == 1:
			return "one"
		case integer && i == 2:
			return "two"
		}
	case "fr":
		switch {
		case i == 0 || i == 1:
			return "one"
		case integer && i != 0 && i%1000000 == 0:
			return "many"
		}
	case "es", "it", "pt":
		switch {
		case integer && i == 1:
			return "one"
		case integer && i != 0 && i%1000000 == 0:
			return "many"
		}
	case "pl":
		switch {
		case !integer:
			return "other"
		case i == 1:
			return "one"
		case mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14):
			return "few"
		default:
			return "many"
		}
	case "ru", "uk":
		switch {
		case !integer:
			return "other"
		case mod10 == 1 && mod100 != 11:
			return "one"
		case mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14):
			return "few"
		default:
			return "many"
		}
	case "cs", "sk":
		switch {
		case !integer:
			return "many"
		case i == 1:
			return "one"
		case i >= 2 && i <= 4:
			return "few"
		}
	case "ro":
		switch {
		case integer && i == 1:
			return "one"
		case !integer || i == 0 || (mod100 >= 1 && mod100 <= 19):
			return "few"
		}
	case "ar":
		if !integer {
			return "other"
		}
		switch {
		case i == 0:
			return "zero"
		case i == 1:
			return "one"
		case i == 2:
			return "two"
		case mod100 >= 3 && mod100 <= 10:
			return "few"
		case mod100 >= 11 && mod100 <= 99:
			return "many"
		}
	}
	return "other"
}

// PluralCategories lists the categories a catalog must define for lang:
// every category integers can produce plus "other" (the fallback).
func PluralCategories(lang string) []string {
	switch lang {
	case "pl", "ru", "uk":
		return []string{"one", "few", "many", "other"}
	case "cs", "sk", "ro":
		return []string{"one", "few", "other"}
	case "fr", "es", "it", "pt":
		return []string{"one", "many", "other"}
	case "he":
		return []string{"one", "two", "other"}
	case "ar":
		return []string{"zero", "one", "two", "few", "many", "other"}
	case "id", "ms", "vi", "zh", "ja", "ko":
		return []string{"other"}
	}
	return []string{"one", "other"}
}
