// Package desktop runs the WebView2 window and exposes Go functions to the UI.
package desktop

import (
	"encoding/json"
	"fmt"

	webview "github.com/abemedia/go-webview"
	_ "github.com/abemedia/go-webview/embedded"

	"lantransfer/internal/app"
	"lantransfer/internal/config"
	"lantransfer/internal/i18n"
	"lantransfer/internal/service"
	"lantransfer/internal/ui"
	"lantransfer/internal/version"
)

// Run starts the receiving service and blocks until the window closes.
func Run(state *app.State) error {
	server, err := service.Start(state)
	if err != nil {
		return i18n.Wrap(err, "err.serviceStart", "port", fmt.Sprint(config.HTTPPort), "error", err)
	}
	defer server.Close()

	language := state.Language
	if language.Language == "" {
		language = app.ResolveLanguage()
	}
	boot, err := ui.NewBootstrap(language.Language, language.Source, language.Detected)
	if err != nil {
		return err
	}
	html, err := ui.HTML(boot)
	if err != nil {
		return err
	}

	window := webview.New(false)
	defer window.Destroy()
	window.SetTitle(WindowTitle(language.Language))

	api := &bridge{state: state, window: window}
	if err := api.bindAll(); err != nil {
		return err
	}
	window.SetHtml(html)
	cleanupRestore := restoreWindowState(window)
	defer cleanupRestore()
	cleanupDrop := installFileDrop(window, func(paths []string) {
		payload, err := json.Marshal(paths)
		if err != nil {
			return
		}
		window.Dispatch(func() {
			window.Eval("window.__addDroppedPaths(" + string(payload) + ")")
		})
	})
	defer cleanupDrop()
	cleanupWindowState := installWindowStateTracking(window)
	defer cleanupWindowState()
	window.Run()
	return nil
}

// WindowTitle is the native window title in lang ("LAN Transfer v2.1.0 –
// file transfer on your local network"). The page sets the same text as
// document.title; the setLanguage binding updates both live.
func WindowTitle(lang string) string {
	return i18n.T(lang, "app.windowTitle", "version", version.String())
}
