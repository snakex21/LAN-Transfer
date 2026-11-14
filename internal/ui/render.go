package ui

import (
	"embed"
	"strings"
)

//go:embed assets/index.html assets/style.css assets/app.js
var assets embed.FS

func HTML() (string, error) {
	html, err := assets.ReadFile("assets/index.html")
	if err != nil {
		return "", err
	}
	style, err := assets.ReadFile("assets/style.css")
	if err != nil {
		return "", err
	}
	script, err := assets.ReadFile("assets/app.js")
	if err != nil {
		return "", err
	}

	result := strings.Replace(string(html), "/*__STYLE__*/", string(style), 1)
	result = strings.Replace(result, "//__SCRIPT__", string(script), 1)
	return result, nil
}
