package i18n

import (
	"math"
	"strconv"
	"strings"
)

type numberFormat struct {
	decimal  string
	group    string
	minGroup int // 2: no separator in 4-digit numbers (pl, es, …)
	percent  string
	units    map[string]string
}

func formatOf(lang string) numberFormat {
	if f, ok := numberFormats[lang]; ok {
		return f
	}
	return numberFormats[Fallback]
}

// FormatNumber formats n with at most maxFrac fraction digits (trailing
// zeros dropped) using the separators of lang and Latin digits.
func FormatNumber(lang string, n float64, maxFrac int) string {
	f := formatOf(lang)
	negative := n < 0
	text := strconv.FormatFloat(math.Abs(n), 'f', maxFrac, 64)
	intPart, frac, _ := strings.Cut(text, ".")
	frac = strings.TrimRight(frac, "0")
	if f.group != "" && len(intPart) >= 4+f.minGroup-1 {
		var b strings.Builder
		for i, digit := range intPart {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				b.WriteString(f.group)
			}
			b.WriteRune(digit)
		}
		intPart = b.String()
	}
	if frac != "" {
		intPart += f.decimal + frac
	}
	if negative && strings.Trim(intPart, "0"+f.decimal) != "" {
		intPart = "-" + intPart
	}
	return intPart
}

// FormatUnit renders n with a CLDR short unit ("second", "megabyte", …).
func FormatUnit(lang, unit string, n float64, maxFrac int) string {
	pattern, ok := formatOf(lang).units[unit]
	if !ok {
		pattern = "{n} " + unit
	}
	return strings.Replace(pattern, "{n}", FormatNumber(lang, n, maxFrac), 1)
}

var sizeUnits = []string{"byte", "kilobyte", "megabyte", "gigabyte", "terabyte"}

func scale(n float64) (float64, int, int) {
	unit := 0
	for n >= 1024 && unit < len(sizeUnits)-1 {
		n /= 1024
		unit++
	}
	frac := 1
	if n >= 10 || unit == 0 {
		frac = 0
	}
	return n, unit, frac
}

// FormatSize renders a byte count ("1,5 MB", "1.5 Mo"), like formatBytes in the UI.
func FormatSize(lang string, bytes int64) string {
	if bytes <= 0 {
		return FormatUnit(lang, "byte", 0, 0)
	}
	value, unit, frac := scale(float64(bytes))
	return FormatUnit(lang, sizeUnits[unit], value, frac)
}

// FormatSpeed renders bytes per second ("12 MB/s").
func FormatSpeed(lang string, bps float64) string {
	if bps <= 0 || math.IsNaN(bps) || math.IsInf(bps, 0) {
		return "—"
	}
	value, unit, frac := scale(bps)
	return FormatUnit(lang, sizeUnits[unit]+"-per-second", value, frac)
}

// FormatDuration renders seconds as "42 s", "3 min 5 s" or "2 h 10 min".
func FormatDuration(lang string, seconds float64) string {
	if seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return "—"
	}
	if seconds < 60 {
		return FormatUnit(lang, "second", math.Ceil(seconds), 0)
	}
	minutes := math.Floor(seconds / 60)
	if minutes < 60 {
		return FormatUnit(lang, "minute", minutes, 0) + " " + FormatUnit(lang, "second", math.Round(math.Mod(seconds, 60)), 0)
	}
	return FormatUnit(lang, "hour", math.Floor(minutes/60), 0) + " " + FormatUnit(lang, "minute", math.Mod(minutes, 60), 0)
}

// FormatPercent renders a percentage given as 0–100 with one fraction digit.
func FormatPercent(lang string, value float64) string {
	return strings.Replace(formatOf(lang).percent, "{n}", FormatNumber(lang, value, 1), 1)
}
