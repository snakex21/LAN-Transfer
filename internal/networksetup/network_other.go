//go:build !windows

package networksetup

func HandleSpecialMode(_ []string) (bool, error) { return false, nil }
func EnsureConfigured()                          {}
func Status() FirewallStatus                     { return FirewallStatus{Skipped: true} }
func UpdateLegacyRules() error                   { return nil }
