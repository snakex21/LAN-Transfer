// Package desktop runs the WebView2 window and exposes Go functions to the UI.
package desktop

import (
	"encoding/json"
	"fmt"

	webview "github.com/abemedia/go-webview"
	_ "github.com/abemedia/go-webview/embedded"

	"lantransfer/internal/app"
	"lantransfer/internal/service"
	"lantransfer/internal/ui"
)

// Run starts the receiving service and blocks until the window closes.
func Run(state *app.State) error {
	server, err := service.Start(state)
	if err != nil {
		return fmt.Errorf("nie można uruchomić odbioru na porcie (czy LAN Transfer już działa?): %w", err)
	}
	defer server.Close()

	html, err := ui.HTML()
	if err != nil {
		return err
	}

	window := webview.New(false)
	defer window.Destroy()
	window.SetTitle("LAN Transfer")

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
