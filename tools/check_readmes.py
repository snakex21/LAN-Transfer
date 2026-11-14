"""Checks that README.md and the localized README/README.<code>.md files stay in sync.

Run: python tools/check_readmes.py   (also part of CI)

Every localized file must have the same language bar, the same Markdown
structure (heading levels, list items, table rows, code fences, images), every
technical literal written in `inline code` in the English README, working
#anchors (GitHub heading slugs) and existing relative links.
"""
from pathlib import Path
import re
import sys
import unicodedata

ROOT = Path(__file__).resolve().parents[1]
MAIN = ROOT / "README.md"
README_DIR = ROOT / "README"
CODES = [
    "pl", "de", "fr", "es", "it", "pt", "nl", "sv", "no", "da", "fi", "cs", "sk", "hu", "ro", "tr", "id", "ms", "vi",
    "ru", "uk", "bg", "el", "ar", "he", "hi", "zh", "ja", "ko",
]
LANGUAGES = 1 + len(CODES)

LINK = re.compile(r"!?\[[^\]]*\]\(([^)\s]+)\)")


def fail(message: str) -> None:
    print(f"README ERROR: {message}", file=sys.stderr)
    raise SystemExit(1)


def read(path: Path) -> str:
    try:
        return path.read_text(encoding="utf-8")
    except Exception as exc:  # noqa: BLE001
        fail(f"cannot read {path.relative_to(ROOT)} as UTF-8: {exc}")


def strip_code(lines):
    """Lines outside fenced code blocks."""
    inside = False
    for line in lines:
        if line.startswith("```"):
            inside = not inside
            continue
        if not inside:
            yield line


def structure(text: str):
    lines = text.splitlines()
    body = list(strip_code(lines))
    headings = tuple(len(m.group(1)) for line in body if (m := re.match(r"^(#{1,6})\s+", line)))
    numbered = sum(bool(re.match(r"^\d+\.\s+", line)) for line in body)
    bullets = sum(bool(re.match(r"^-\s+", line)) for line in body)
    table_rows = sum(line.startswith("|") for line in body)
    fences = sum(line.startswith("```") for line in lines)
    images = sum(line.count("![") for line in body)
    return headings, numbered, bullets, table_rows, fences, images


def code_blocks(text: str):
    blocks, current, inside = [], [], False
    for line in text.splitlines():
        if line.startswith("```"):
            if inside:
                blocks.append("\n".join(current))
                current = []
            inside = not inside
            continue
        if inside:
            current.append(line)
    return blocks


def slug(heading: str) -> str:
    """GitHub-style anchor for a heading text."""
    text = re.sub(r"`|\*\*|__", "", heading.strip()).lower()
    kept = []
    for ch in text:
        cat = unicodedata.category(ch)
        if ch in " -_" or cat[0] in ("L", "M", "N"):
            kept.append("-" if ch == " " else ch)
    return "".join(kept)


def anchors(text: str):
    seen = {}
    result = set()
    for line in strip_code(text.splitlines()):
        m = re.match(r"^#{1,6}\s+(.*)$", line)
        if not m:
            continue
        base = slug(m.group(1))
        n = seen.get(base, 0)
        seen[base] = n + 1
        result.add(base if n == 0 else f"{base}-{n}")
    return result


def check_links(path: Path, text: str) -> None:
    own = anchors(text)
    for line in strip_code(text.splitlines()):
        for target in LINK.findall(line):
            if target.startswith(("http://", "https://", "mailto:")):
                continue
            if target.startswith("#"):
                if target[1:] not in own:
                    fail(f"{path.relative_to(ROOT)}: anchor {target} matches no heading")
                continue
            file_part = target.split("#", 1)[0]
            if not (path.parent / file_part).resolve().exists():
                fail(f"{path.relative_to(ROOT)}: broken link {target}")


def language_bar(path: Path, text: str) -> None:
    lines = text.splitlines()
    if not lines or not lines[0].startswith("**Languages:**"):
        fail(f"{path.relative_to(ROOT)}: does not start with the language navigation bar")
    links = re.findall(r"\[[^\]]+\]\(([^)]+)\)", lines[0])
    if len(links) != LANGUAGES:
        fail(f"{path.relative_to(ROOT)}: language bar has {len(links)} links, expected {LANGUAGES}")
    for target in links:
        if not (path.parent / target).resolve().is_file():
            fail(f"{path.relative_to(ROOT)}: broken language link: {target}")


main = read(MAIN)
expected_structure = structure(main)
expected_blocks = code_blocks(main)
technical_tokens = set(re.findall(r"`([^`\n]+)`", main))
language_bar(MAIN, main)
check_links(MAIN, main)

# Optional arguments limit the check to some languages (while translating).
selected = sys.argv[1:] or CODES
if not sys.argv[1:]:
    expected_files = {f"README.{code}.md" for code in CODES}
    actual_files = {p.name for p in README_DIR.glob("README.*.md")}
    if missing := sorted(expected_files - actual_files):
        fail(f"missing localized README files: {', '.join(missing)}")
    if extra := sorted(actual_files - expected_files):
        fail(f"unexpected localized README files: {', '.join(extra)}")

for code in selected:
    path = README_DIR / f"README.{code}.md"
    text = read(path)
    language_bar(path, text)
    if (got := structure(text)) != expected_structure:
        fail(f"{path.name}: Markdown structure {got} differs from README.md {expected_structure}")
    if code_blocks(text) != expected_blocks:
        fail(f"{path.name}: code blocks differ from README.md (commands must not be translated)")
    present = set(re.findall(r"`([^`\n]+)`", text))
    if missing := sorted(technical_tokens - present):
        fail(f"{path.name}: missing technical literals: {', '.join(missing)}")
    check_links(path, text)

print(
    f"README OK: {len(selected)}/{len(CODES)} localized files checked, structure={expected_structure}, "
    f"code blocks={len(expected_blocks)}, technical literals={len(technical_tokens)}"
)
