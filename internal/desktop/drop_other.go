//go:build !windows

package desktop

import webview "github.com/abemedia/go-webview"

func installFileDrop(_ webview.WebView, _ func([]string)) func() {
	return func() {}
}
