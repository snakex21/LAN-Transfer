//go:build !windows

package i18n

import (
	"os"
	"strings"
)

// SystemLanguages returns the POSIX locale (LANGUAGE, LC_ALL, LC_MESSAGES,
// LANG), most preferred first.
func SystemLanguages() []string {
	var result []string
	for _, name := range []string{"LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
		for _, tag := range strings.Split(os.Getenv(name), ":") {
			if tag = strings.TrimSpace(tag); tag != "" && tag != "C" && tag != "POSIX" {
				result = append(result, tag)
			}
		}
	}
	return result
}
