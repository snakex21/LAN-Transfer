package desktop

import (
	"encoding/json"
	"fmt"
	"time"

	webview "github.com/abemedia/go-webview"
	_ "github.com/abemedia/go-webview/embedded"

	"lantransfer/internal/app"
	"lantransfer/internal/config"
	"lantransfer/internal/discovery"
	"lantransfer/internal/filebrowser"
	"lantransfer/internal/preview"
	"lantransfer/internal/service"
	"lantransfer/internal/transfer"
	"lantransfer/internal/ui"
)

func Run(state *app.State) error {
	server, err := service.Start(state)
	if err != nil {
		return err
	}
	defer server.Close()

	html, err := ui.HTML()
	if err != nil {
		return err
	}

	window := webview.New(false)
	defer window.Destroy()
	window.SetTitle("LAN Transfer")

	if err := bind(window, state); err != nil {
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

func bind(window webview.WebView, state *app.State) error {
	if err := window.Bind("appInfo", func() map[string]any {
		return map[string]any{
			"id":          state.ID,
			"name":        state.Name,
			"receive_dir": state.ReceiveDir,
			"port":        config.HTTPPort,
		}
	}); err != nil {
		return fmt.Errorf("bind appInfo: %w", err)
	}

	if err := window.Bind("discoverDevices", func() map[string]any {
		devices, err := discovery.Discover(700*time.Millisecond, state.ID)
		if err != nil {
			return map[string]any{"ok": false, "error": err.Error(), "devices": []any{}}
		}
		return map[string]any{"ok": true, "error": "", "devices": devices}
	}); err != nil {
		return fmt.Errorf("bind discoverDevices: %w", err)
	}

	if err := window.Bind("pickLocalFiles", func() map[string]any {
		paths, err := filebrowser.PickFiles()
		if err != nil {
			return map[string]any{"ok": false, "error": err.Error(), "paths": []string{}}
		}
		return map[string]any{"ok": true, "error": "", "paths": paths}
	}); err != nil {
		return fmt.Errorf("bind pickLocalFiles: %w", err)
	}

	if err := window.Bind("pickLocalFolder", func() map[string]any {
		path, err := filebrowser.PickFolder()
		if err != nil {
			return map[string]any{"ok": false, "error": err.Error(), "path": ""}
		}
		return map[string]any{"ok": true, "error": "", "path": path}
	}); err != nil {
		return fmt.Errorf("bind pickLocalFolder: %w", err)
	}

	if err := window.Bind("previewLocalPath", func(path string) map[string]any {
		info, err := preview.Load(path)
		if err != nil {
			return map[string]any{"ok": false, "error": err.Error()}
		}
		return map[string]any{"ok": true, "error": "", "preview": info}
	}); err != nil {
		return fmt.Errorf("bind previewLocalPath: %w", err)
	}

	if err := window.Bind("browseRemote", func(target, path string) map[string]any {
		listing, err := filebrowser.BrowseRemote(target, path)
		if err != nil {
			return map[string]any{"ok": false, "error": err.Error()}
		}
		return map[string]any{"ok": true, "error": "", "listing": listing}
	}); err != nil {
		return fmt.Errorf("bind browseRemote: %w", err)
	}

	if err := window.Bind("sendLocalPaths", func(paths []string, target, destination string) map[string]any {
		for _, path := range paths {
			if err := transfer.SendPathTo(path, target, state.ID, destination); err != nil {
				return map[string]any{"ok": false, "error": err.Error()}
			}
		}
		return map[string]any{"ok": true, "error": ""}
	}); err != nil {
		return fmt.Errorf("bind sendLocalPaths: %w", err)
	}
	return nil
}
