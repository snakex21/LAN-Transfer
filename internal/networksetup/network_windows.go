//go:build windows

package networksetup

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	"lantransfer/internal/config"
)

const (
	// configureArg: create the 2.x rules (first start, no rules present).
	configureArg = "--lantransfer-configure-network"
	// updateArg: replace the 1.x rules with the 2.x ones (user clicked "Update").
	updateArg = "--lantransfer-update-network"
)

var (
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteW = shell32.NewProc("ShellExecuteW")
)

// HandleSpecialMode handles the short-lived elevated helper invocation.
func HandleSpecialMode(args []string) (bool, error) {
	if len(args) != 1 {
		return false, nil
	}
	switch args[0] {
	case configureArg:
		return true, configureFirewall(false)
	case updateArg:
		return true, configureFirewall(true)
	}
	return false, nil
}

// EnsureConfigured makes the application independent from Windows' generic
// Network Discovery switch. We only open the two ports LAN Transfer actually
// needs, for the local subnet, on the Private and Domain profiles. If no rules
// exist yet, a one-time elevated helper is started. Rules left by 1.x are not
// touched here: the UI offers to update them (see Status / UpdateLegacyRules).
func EnsureConfigured() {
	status := Status()
	if status.Skipped || status.Configured || status.Legacy {
		return
	}
	_ = launchElevatedHelper(configureArg)
}

// Status reports which LAN Transfer firewall rules exist.
func Status() FirewallStatus {
	if config.SkipFirewall {
		return FirewallStatus{Skipped: true}
	}
	configured := true
	for _, r := range Rules(config.UDPPort, config.HTTPPort) {
		configured = configured && firewallRuleExists(r.Name)
	}
	legacy := false
	for _, name := range LegacyRuleNames {
		legacy = legacy || firewallRuleExists(name)
	}
	return FirewallStatus{Configured: configured, Legacy: legacy}
}

// UpdateLegacyRules starts the elevated helper (UAC prompt) that removes the
// 1.x rules and creates the 2.x ones. Only called after a user action.
func UpdateLegacyRules() error {
	if config.SkipFirewall {
		return nil
	}
	return launchElevatedHelper(updateArg)
}

func netsh(args []string) *exec.Cmd {
	cmd := exec.Command("netsh", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}

func firewallRuleExists(name string) bool {
	out, err := netsh(ShowRuleArgs(name)).CombinedOutput()
	if err != nil {
		return false
	}
	return bytes.Contains(bytes.ToLower(out), bytes.ToLower([]byte(name)))
}

func configureFirewall(removeLegacy bool) error {
	if removeLegacy {
		for _, name := range LegacyRuleNames {
			_ = netsh(DeleteRuleArgs(name)).Run()
		}
	}
	for _, r := range Rules(config.UDPPort, config.HTTPPort) {
		_ = netsh(DeleteRuleArgs(r.Name)).Run()
		if out, err := netsh(AddRuleArgs(r)).CombinedOutput(); err != nil {
			return fmt.Errorf("firewall rule %q: %w (%s)", r.Name, err, bytes.TrimSpace(out))
		}
	}
	return nil
}

func launchElevatedHelper(arg string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exe)
	params, _ := syscall.UTF16PtrFromString(arg)
	dir, _ := syscall.UTF16PtrFromString(filepath.Dir(exe))

	result, _, callErr := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(params)),
		uintptr(unsafe.Pointer(dir)),
		0,
	)
	if result <= 32 {
		if callErr != syscall.Errno(0) {
			return callErr
		}
		return fmt.Errorf("ShellExecuteW failed: %d", result)
	}
	return nil
}
