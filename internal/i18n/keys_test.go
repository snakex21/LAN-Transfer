package i18n

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const repoRoot = "../.."

var (
	keyShape = regexp.MustCompile(`^[a-z]+(\.[A-Za-z0-9]+)+$`)
	fileName = regexp.MustCompile(`\.(json|js|go|html|css|exe|md|txt)$`)
)

// goKeys collects every string literal in the Go sources (tests excluded)
// that names a catalog key: a literal shaped like "<namespace>.<name>" whose
// namespace is used in the catalog. This covers i18n.New/Wrap/M/T and the
// small helpers that take a key (badRequest, cli t, preview message, …).
func goKeys(t *testing.T, namespaces map[string]bool) map[string][]string {
	t.Helper()
	keys := map[string][]string{}
	fset := token.NewFileSet()
	err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(lit.Value)
			if err != nil || !keyShape.MatchString(value) || fileName.MatchString(value) || !namespaces[strings.SplitN(value, ".", 2)[0]] {
				return true
			}
			keys[value] = append(keys[value], fset.Position(lit.Pos()).String())
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return keys
}

var (
	jsKey       = regexp.MustCompile(`\bt[nm]?\(\s*'([^']+)'`)
	htmlKey     = regexp.MustCompile(`data-i18n(?:-[a-z-]+)?="([^"]+)"`)
	htmlComment = regexp.MustCompile(`(?s)<!--.*?-->`)
)

func webKeys(t *testing.T) map[string][]string {
	t.Helper()
	keys := map[string][]string{}
	for _, name := range []string{"app.js", "i18n.js", "index.html"} {
		data, err := os.ReadFile(filepath.Join(repoRoot, "internal", "ui", "assets", name))
		if err != nil {
			t.Fatal(err)
		}
		source := htmlComment.ReplaceAllString(string(data), "")
		for _, re := range []*regexp.Regexp{jsKey, htmlKey} {
			for _, m := range re.FindAllStringSubmatch(source, -1) {
				keys[m[1]] = append(keys[m[1]], name)
			}
		}
	}
	return keys
}

// Every key used by the Go backend, the CLI and the page exists in all 30
// catalogs, and every catalog key is used somewhere.
func TestEveryUsedKeyExistsInAllLocales(t *testing.T) {
	if err := LoadError(); err != nil {
		t.Fatal(err)
	}
	reference := Keys(Source)
	namespaces := map[string]bool{}
	for key := range reference {
		namespaces[strings.SplitN(key, ".", 2)[0]] = true
	}
	used := goKeys(t, namespaces)
	if len(used) < 50 {
		t.Fatalf("found only %d keys in Go sources; is the scanner broken?", len(used))
	}
	for key, where := range webKeys(t) {
		used[key] = append(used[key], where...)
	}
	for _, lang := range Languages {
		keys := Keys(lang.Code)
		var missing []string
		for key, where := range used {
			if !keys[key] {
				missing = append(missing, key+" ("+where[0]+")")
			}
		}
		sort.Strings(missing)
		if len(missing) > 0 {
			t.Errorf("ui.%s.json lacks %d used keys: %s", lang.Code, len(missing), strings.Join(missing, ", "))
		}
	}
	var unused []string
	for key := range reference {
		if _, ok := used[key]; !ok {
			unused = append(unused, key)
		}
	}
	sort.Strings(unused)
	if len(unused) > 0 {
		t.Errorf("keys in ui.pl.json that nothing uses: %s", strings.Join(unused, ", "))
	}
}

// The page must not contain hard-coded Polish text any more.
func TestNoHardcodedPolishInPage(t *testing.T) {
	polish := regexp.MustCompile(`[ąćęłńśźżĄĆĘŁŃŚŹŻ]`)
	comment := regexp.MustCompile(`(?s)/\*.*?\*/|//[^\n]*|<!--.*?-->`)
	for _, name := range []string{"app.js", "i18n.js", "index.html"} {
		data, err := os.ReadFile(filepath.Join(repoRoot, "internal", "ui", "assets", name))
		if err != nil {
			t.Fatal(err)
		}
		code := comment.ReplaceAllString(string(data), "")
		if loc := polish.FindStringIndex(code); loc != nil {
			start := max(loc[0]-40, 0)
			t.Errorf("%s contains Polish text: …%s…", name, code[start:min(loc[1]+40, len(code))])
		}
	}
}
