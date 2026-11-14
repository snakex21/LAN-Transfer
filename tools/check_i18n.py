"""Checks the UI translations in locales/ui.<code>.json.

Run: python tools/check_i18n.py [code ...]   (also part of CI)

ui.pl.json is the source language, ui.en.json the fallback. Every file must:
- be valid UTF-8 JSON with exactly the keys of ui.pl.json;
- have no empty values and no leading/trailing whitespace;
- use, in every value, the same {placeholders} (with the same :format) as
  English; plural values are objects with exactly the CLDR categories the
  language needs for whole numbers (plus "other"), and only those categories
  may leave out {count};
- keep technical literals (SHA-256, UAC, LAN Transfer, CLI flags, variable
  names, command lines of the CLI help);
- contain no Polish text outside ui.pl.json (Polish-only letters, or a value
  copied from Polish), and, for languages with their own script, no value
  that is still plain English; Latin-script languages may share only a few
  values with English (product names, short technical words).
"""
from pathlib import Path
import json
import re
import sys

ROOT = Path(__file__).resolve().parents[1]
LOCALES = ROOT / "locales"
SOURCE, FALLBACK = "pl", "en"
CODES = [
    "en", "pl", "de", "fr", "es", "it", "pt", "nl", "sv", "no", "da", "fi", "cs", "sk", "hu", "ro", "tr", "id", "ms",
    "vi", "ru", "uk", "bg", "el", "ar", "he", "hi", "zh", "ja", "ko",
]

# CLDR cardinal categories needed for whole numbers, plus "other" (the
# fallback); must match PluralCategories in internal/i18n/plural.go.
PLURALS = {code: ["one", "other"] for code in CODES}
PLURALS.update({c: ["one", "few", "many", "other"] for c in ("pl", "ru", "uk")})
PLURALS.update({c: ["one", "few", "other"] for c in ("cs", "sk", "ro")})
PLURALS.update({c: ["one", "many", "other"] for c in ("fr", "es", "it", "pt")})
PLURALS.update({c: ["other"] for c in ("id", "ms", "vi", "zh", "ja", "ko")})
PLURALS["he"] = ["one", "two", "other"]
PLURALS["ar"] = ["zero", "one", "two", "few", "many", "other"]

PLACEHOLDER = re.compile(r"\{([A-Za-z0-9_]+)(?::([a-z]+))?\}")
FORMATS = {None, "size", "speed", "duration", "seconds", "percent", "number", "datetime"}
PROTECTED = ["LAN Transfer", "SHA-256", "UAC", "UTF-8", "HTTP", "Windows", "LANTransferCLI.exe", "'/'"]
TECH = re.compile(r"--[a-z]+|LANTRANSFER_[A-Z_]+|LANTransferCLI\.exe [a-z]+")
POLISH_LETTERS = re.compile(r"[ąćęłńśźżĄĆĘŁŃŚŹŻ]")
SCRIPTS = {
    "ru": r"Ѐ-ӿ", "uk": r"Ѐ-ӿ", "bg": r"Ѐ-ӿ", "el": r"Ͱ-Ͽ",
    "ar": r"؀-ۿ", "he": r"֐-׿", "hi": r"ऀ-ॿ",
    "zh": r"一-鿿", "ja": r"぀-ヿ一-鿿", "ko": r"가-힯",
}

# Values identical to English that a Latin-script language may keep.
MAX_SAME_AS_ENGLISH = 8

errors = []


def error(code, key, message):
    errors.append(f"ui.{code}.json: {key}: {message}" if key else f"ui.{code}.json: {message}")


def load(code):
    path = LOCALES / f"ui.{code}.json"
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        error(code, None, "file is missing")
        return None
    except Exception as exc:  # noqa: BLE001
        error(code, None, f"invalid JSON: {exc}")
        return None
    if not isinstance(data, dict):
        error(code, None, "top level must be an object")
        return None
    return data


def forms(value):
    return value.values() if isinstance(value, dict) else [value]


def placeholders(text):
    return {(m.group(1), m.group(2)) for m in PLACEHOLDER.finditer(text)}


def reference_placeholders(value):
    result = set()
    for text in forms(value):
        result |= placeholders(text)
    return result


def check_value(code, key, value, ref):
    if isinstance(ref, dict):
        if not isinstance(value, dict):
            error(code, key, "must be a plural object")
            return
        wanted = PLURALS[code]
        if sorted(value) != sorted(wanted):
            error(code, key, f"plural categories {sorted(value)}, expected {sorted(wanted)}")
    elif not isinstance(value, str):
        error(code, key, "must be a string")
        return
    expected = reference_placeholders(ref)
    for name, fmt in expected:
        if fmt not in FORMATS:
            error(code, key, f"unknown format :{fmt}")
    items = value.items() if isinstance(value, dict) else [("", value)]
    for category, text in items:
        where = f"{key}[{category}]" if category else key
        if not isinstance(text, str) or not text.strip():
            error(code, where, "empty value")
            continue
        if text != text.strip():
            error(code, where, "leading or trailing whitespace")
        got = placeholders(text)
        if extra := got - expected:
            error(code, where, f"unknown placeholders {sorted(extra)}")
        missing = expected - got
        if category and category != "other":
            missing = {p for p in missing if p[0] != "count"}
        if missing:
            error(code, where, f"missing placeholders {sorted(missing)}")
        if text.count("{") != text.count("}"):
            error(code, where, "unbalanced braces")
        if key != "cli.usage" and re.search(r"</?[a-z]+[ >/]", text):
            error(code, where, "HTML markup is not allowed (texts are inserted as plain text)")


def check_literals(code, key, value, en_value):
    en_text = " ".join(forms(en_value))
    text = " ".join(forms(value))
    for literal in PROTECTED:
        if literal in en_text and literal not in text:
            error(code, key, f"technical literal {literal!r} is missing")
    if sorted(TECH.findall(text)) != sorted(TECH.findall(en_text)):
        error(code, key, f"CLI literals differ from English: {sorted(set(TECH.findall(en_text)) ^ set(TECH.findall(text)))}")
    if key == "cli.usage":
        en_cmds = [line.split()[:2] for line in en_text.splitlines() if line.strip().startswith("LANTransferCLI.exe")]
        cmds = [line.split()[:2] for line in text.splitlines() if line.strip().startswith("LANTransferCLI.exe")]
        if cmds != en_cmds:
            error(code, key, "command lines differ from English")
        if text.count("\n") != en_text.count("\n"):
            error(code, key, "line count differs from English")


def strip_technical(text):
    text = PLACEHOLDER.sub(" ", text)
    text = TECH.sub(" ", text)
    for literal in PROTECTED + ["MIT", "IP", "PDF", "ip", "port", "json", "OK"]:
        text = text.replace(literal, " ")
    text = re.sub(r"\b[A-Z]\b|\b\d+\b", " ", text)
    return text


def check_language(code, data, source, en):
    if set(data) != set(source):
        if missing := sorted(set(source) - set(data)):
            error(code, None, f"missing keys: {', '.join(missing)}")
        if extra := sorted(set(data) - set(source)):
            error(code, None, f"keys not in ui.{SOURCE}.json: {', '.join(extra)}")
    script = re.compile(f"[{SCRIPTS[code]}]") if code in SCRIPTS else None
    same_as_english = []
    for key, value in data.items():
        if key not in source:
            continue
        check_value(code, key, value, en.get(key, source[key]))
        check_literals(code, key, value, en.get(key, source[key]))
        if code == SOURCE:
            continue
        for text in forms(value):
            if isinstance(text, str) and POLISH_LETTERS.search(text):
                error(code, key, f"contains Polish text: {text!r}")
        if value == source[key] and source[key] != en.get(key):
            # one-word values such as "Typ" may legitimately match Polish
            if any(len(re.findall(r"[^\W\d_]+", strip_technical(t))) >= 2 for t in forms(value)):
                error(code, key, "copied from Polish (not translated)")
        if code != FALLBACK and value == en.get(key) and not script:
            if any(len(re.findall(r"[A-Za-z]", strip_technical(t))) >= 4 for t in forms(value)):
                same_as_english.append(key)
        if script and code != FALLBACK:
            for text in forms(value):
                rest = strip_technical(text) if isinstance(text, str) else ""
                if len(re.findall(r"[A-Za-z]", rest)) >= 4 and not script.search(text):
                    error(code, key, f"looks untranslated (no {code} script): {text!r}")
    if len(same_as_english) > MAX_SAME_AS_ENGLISH:
        error(code, None, f"{len(same_as_english)} values are identical to English (max {MAX_SAME_AS_ENGLISH}): "
              f"{', '.join(same_as_english[:20])}")


def main():
    for stream in (sys.stdout, sys.stderr):
        stream.reconfigure(encoding="utf-8")
    selected = sys.argv[1:] or CODES
    source = load(SOURCE)
    en = load(FALLBACK)
    if source is None or en is None:
        print("\n".join(errors), file=sys.stderr)
        return 1
    if not sys.argv[1:]:
        files = {p.name for p in LOCALES.glob("ui.*.json")}
        expected = {f"ui.{c}.json" for c in CODES}
        if extra := sorted(files - expected):
            errors.append(f"unexpected locale files: {', '.join(extra)}")
    for code in selected:
        if code not in CODES:
            errors.append(f"unknown language code {code}")
            continue
        data = source if code == SOURCE else en if code == FALLBACK else load(code)
        if data is not None:
            check_language(code, data, source, en)
    if errors:
        print("\n".join(f"I18N ERROR: {e}" for e in errors), file=sys.stderr)
        print(f"{len(errors)} problem(s)", file=sys.stderr)
        return 1
    plural_keys = sum(isinstance(v, dict) for v in source.values())
    print(f"I18N OK: {len(selected)}/{len(CODES)} languages, {len(source)} keys ({plural_keys} with plural forms)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
