//go:build !windows

package desktop

import (
	webview "github.com/abemedia/go-webview"

	"lantransfer/internal/config"
)

func restoreWindowState(window webview.WebView) func() {
	state, _ := config.LoadWindowState()
	window.SetSize(state.Width, state.Height, webview.HintNone)
	return func() {}
}

func installWindowStateTracking(window webview.WebView) func() { return func() {} }
