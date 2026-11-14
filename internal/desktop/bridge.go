package desktop

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	webview "github.com/abemedia/go-webview"

	"lantransfer/internal/app"
	"lantransfer/internal/config"
	"lantransfer/internal/discovery"
	"lantransfer/internal/filebrowser"
	"lantransfer/internal/i18n"
	"lantransfer/internal/model"
	"lantransfer/internal/networksetup"
	"lantransfer/internal/preview"
	"lantransfer/internal/protocol"
	"lantransfer/internal/security"
	"lantransfer/internal/transfer"
	"lantransfer/internal/ui"
	"lantransfer/internal/version"
)

// result is the uniform shape every binding returns to JavaScript.
// A failure carries "error" (text in the current language, for logs) and
// "error_msg" (translation key + parameters, rendered by the page).
type result map[string]any

func ok(fields result) result {
	if fields == nil {
		fields = result{}
	}
	fields["ok"] = true
	fields["error"] = ""
	return fields
}

func fail(err error) result {
	r := result{"ok": false, "error": err.Error(), "error_msg": i18n.MsgOf(err)}
	if errors.Is(err, transfer.ErrNotPaired) {
		r["need_pairing"] = true
	}
	return r
}

// bridge holds the functions exposed to the UI through webview.Bind.
type bridge struct {
	state  *app.State
	window webview.WebView
}

func (b *bridge) bindAll() error {
	bindings := map[string]any{
		"appInfo":          b.appInfo,
		"discoverDevices":  b.discoverDevices,
		"pairDevice":       b.pairDevice,
		"newPairingCode":   b.newPairingCode,
		"pickLocalFiles":   b.pickLocalFiles,
		"pickLocalFolder":  b.pickLocalFolder,
		"previewLocalPath": b.previewLocalPath,
		"browseRemote":     b.browseRemote,
		"startTransfer":    b.startTransfer,
		"cancelTransfer":   b.cancelTransfer,
		"retryTransfer":    b.retryTransfer,
		"clearFinished":    b.clearFinished,
		"transferState":    b.transferState,
		"clearHistory":     b.clearHistory,
		"openFolder":       b.openFolder,
		"firewallStatus":   b.firewallStatus,
		"updateFirewall":   b.updateFirewall,
		"setLanguage":      b.setLanguage,
	}
	for name, fn := range bindings {
		if err := b.window.Bind(name, fn); err != nil {
			return fmt.Errorf("bind %s: %w", name, err)
		}
	}
	return nil
}

func (b *bridge) appInfo() result {
	s := b.state
	return ok(result{
		"id": s.ID, "name": s.Name, "receive_dir": s.ReceiveDir, "port": config.HTTPPort,
		"protocol": protocol.Version, "fingerprint": security.ShortFingerprint(s.Identity.Fingerprint),
		"pairing_code": s.Pairing.Code(), "version": version.String(),
	})
}

// setLanguage saves the choice from the settings dialog ("auto" or a code),
// switches the backend language, retitles the native window and returns the
// new bootstrap (catalogs) for the page.
func (b *bridge) setLanguage(choice string) result {
	info, err := app.ChooseLanguage(choice)
	if err != nil {
		return fail(err)
	}
	boot, err := ui.NewBootstrap(info.Language, info.Source, info.Detected)
	if err != nil {
		return fail(err)
	}
	title := WindowTitle(info.Language)
	b.window.Dispatch(func() { b.window.SetTitle(title) })
	return ok(result{"i18n": boot, "title": title})
}

// deviceView adds pairing state to a discovered device.
type deviceView struct {
	model.Device
	Address     string `json:"address"`
	Paired      bool   `json:"paired"`
	Compatible  bool   `json:"compatible"`
	ShortFinger string `json:"short_fingerprint"`
}

func (b *bridge) discoverDevices() result {
	devices, err := discovery.Discover(700*time.Millisecond, b.state.ID)
	paired := b.state.Sender.Paired()
	views := make([]deviceView, 0, len(devices))
	for _, d := range devices {
		address := d.Address()
		fingerprint, isPaired := paired[address]
		// A paired device that now announces a different certificate restarted
		// (new pairing needed) or is being impersonated.
		if isPaired && d.Fingerprint != "" && d.Fingerprint != fingerprint {
			b.state.Sender.Forget(address)
			isPaired = false
		}
		views = append(views, deviceView{
			Device: d, Address: address, Paired: isPaired, Compatible: d.Protocol >= protocol.Version,
			ShortFinger: security.ShortFingerprint(d.Fingerprint),
		})
	}
	if err != nil && len(views) == 0 {
		return fail(i18n.Wrap(err, "err.discovery", "error", err))
	}
	return ok(result{"devices": views})
}

func (b *bridge) pairDevice(address, code string) result {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	session, err := b.state.Sender.Pair(ctx, address, code)
	if err != nil {
		return fail(err)
	}
	return ok(result{"peer": session.PeerName, "fingerprint": security.ShortFingerprint(session.Fingerprint)})
}

func (b *bridge) newPairingCode() result {
	b.state.Pairing.Rotate()
	return ok(result{"pairing_code": b.state.Pairing.Code()})
}

func (b *bridge) pickLocalFiles() result {
	paths, err := filebrowser.PickFiles()
	if err != nil {
		return fail(err)
	}
	return ok(result{"paths": paths})
}

func (b *bridge) pickLocalFolder() result {
	path, err := filebrowser.PickFolder()
	if err != nil {
		return fail(err)
	}
	return ok(result{"path": path})
}

func (b *bridge) previewLocalPath(path string) result {
	info, err := preview.Load(path)
	if err != nil {
		return fail(err)
	}
	return ok(result{"preview": info})
}

func (b *bridge) browseRemote(address, path string) result {
	session := b.state.Sender.Session(address)
	if session == nil {
		return fail(transfer.ErrNotPaired)
	}
	listing, err := session.Browse(context.Background(), path)
	if err != nil {
		if errors.Is(err, transfer.ErrNotPaired) {
			b.state.Sender.Forget(address)
		}
		return fail(err)
	}
	return ok(result{"listing": listing})
}

func (b *bridge) startTransfer(paths []string, address, destination string) result {
	ids, err := b.state.Sender.Enqueue(paths, address, destination)
	if err != nil {
		return fail(err)
	}
	return ok(result{"ids": ids})
}

func (b *bridge) cancelTransfer(id string) result {
	if !b.state.Sender.Cancel(id) {
		return fail(i18n.New("err.cannotCancel"))
	}
	return ok(nil)
}

func (b *bridge) retryTransfer(id string) result {
	if err := b.state.Sender.Retry(id); err != nil {
		return fail(err)
	}
	return ok(nil)
}

func (b *bridge) clearFinished() result {
	b.state.Sender.ClearFinished()
	return ok(nil)
}

func (b *bridge) transferState() result {
	return ok(result{
		"outgoing":     b.state.Sender.Snapshot(),
		"incoming":     b.state.Incoming.Snapshot(),
		"history":      b.state.History.Entries(),
		"pairing_code": b.state.Pairing.Code(),
		"peers":        len(b.state.Pairing.Peers()),
	})
}

func (b *bridge) clearHistory() result {
	b.state.History.Clear()
	return ok(nil)
}

// openFolder shows a folder in Explorer. Only the receive folder or a folder
// recorded in history can be opened, never an arbitrary path from the page.
func (b *bridge) openFolder(path string) result {
	allowed := path == b.state.ReceiveDir
	for _, entry := range b.state.History.Entries() {
		if entry.Location == path {
			allowed = true
			break
		}
	}
	if !allowed {
		return fail(i18n.New("err.folderNotAllowed"))
	}
	if err := exec.Command("explorer.exe", "/select,", path).Start(); err != nil {
		return fail(err)
	}
	return ok(nil)
}

func (b *bridge) firewallStatus() result {
	return ok(result{"firewall": networksetup.Status()})
}

// updateFirewall replaces the 1.x rules (all profiles) with the 2.x ones
// (Private + Domain). Runs only when the user clicks the button in the banner;
// Windows shows its own UAC prompt.
func (b *bridge) updateFirewall() result {
	if err := networksetup.UpdateLegacyRules(); err != nil {
		return fail(i18n.Wrap(err, "err.firewallUpdate", "error", err))
	}
	return ok(nil)
}
