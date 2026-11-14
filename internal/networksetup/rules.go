package networksetup

import "fmt"

// Firewall rules created by LAN Transfer 2.x. They allow inbound traffic only
// from the local subnet and only on the Private and Domain profiles, so that
// a laptop on a network marked Public (café, hotel, airport) does not accept
// connections at all.
const (
	RuleDiscovery = "LAN Transfer 2 Discovery (UDP)"
	RuleTransfer  = "LAN Transfer 2 Transfer (TCP)"

	// FirewallProfiles is the netsh "profile=" value of every rule we create.
	FirewallProfiles = "private,domain"
)

// LegacyRuleNames are the rules created by LAN Transfer 1.x with
// profile=any (Public included). They are only replaced after the user
// confirms it in the UI.
var LegacyRuleNames = []string{"LAN Transfer Discovery", "LAN Transfer Transfer"}

// Rule describes one inbound allow rule.
type Rule struct {
	Name     string
	Protocol string // "UDP" or "TCP"
	Port     int
}

// Rules returns the rules for the given discovery (UDP) and transfer (TCP) ports.
func Rules(udpPort, tcpPort int) []Rule {
	return []Rule{
		{Name: RuleDiscovery, Protocol: "UDP", Port: udpPort},
		{Name: RuleTransfer, Protocol: "TCP", Port: tcpPort},
	}
}

// AddRuleArgs returns the netsh arguments that create the rule.
func AddRuleArgs(r Rule) []string {
	return []string{
		"advfirewall", "firewall", "add", "rule",
		"name=" + r.Name,
		"dir=in",
		"action=allow",
		"protocol=" + r.Protocol,
		fmt.Sprintf("localport=%d", r.Port),
		"profile=" + FirewallProfiles,
		"remoteip=LocalSubnet",
		"enable=yes",
	}
}

// DeleteRuleArgs returns the netsh arguments that delete every rule with this name.
func DeleteRuleArgs(name string) []string {
	return []string{"advfirewall", "firewall", "delete", "rule", "name=" + name}
}

// ShowRuleArgs returns the netsh arguments that print the rule (exit code 1 if missing).
func ShowRuleArgs(name string) []string {
	return []string{"advfirewall", "firewall", "show", "rule", "name=" + name}
}

// FirewallStatus is reported to the UI.
type FirewallStatus struct {
	// Configured: the 2.x rules exist.
	Configured bool `json:"configured"`
	// Legacy: 1.x rules (all profiles, Public included) exist; the UI offers
	// to replace them. Nothing is changed until the user clicks the button.
	Legacy bool `json:"legacy"`
	// Skipped: firewall handling is disabled (loopback / LANTRANSFER_NO_FIREWALL)
	// or the platform has no Windows Firewall.
	Skipped bool `json:"skipped"`
}
