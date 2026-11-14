package networksetup

import (
	"strings"
	"testing"
)

func TestAddRuleArgsArePrivateAndDomainOnly(t *testing.T) {
	rules := Rules(45454, 45455)
	if len(rules) != 2 {
		t.Fatalf("want 2 rules, got %d", len(rules))
	}
	want := map[string][]string{
		RuleDiscovery: {"advfirewall", "firewall", "add", "rule", "name=" + RuleDiscovery, "dir=in", "action=allow",
			"protocol=UDP", "localport=45454", "profile=private,domain", "remoteip=LocalSubnet", "enable=yes"},
		RuleTransfer: {"advfirewall", "firewall", "add", "rule", "name=" + RuleTransfer, "dir=in", "action=allow",
			"protocol=TCP", "localport=45455", "profile=private,domain", "remoteip=LocalSubnet", "enable=yes"},
	}
	for _, r := range rules {
		got := AddRuleArgs(r)
		if strings.Join(got, "\x00") != strings.Join(want[r.Name], "\x00") {
			t.Errorf("%s:\n got  %q\n want %q", r.Name, got, want[r.Name])
		}
		for _, arg := range got {
			lower := strings.ToLower(arg)
			if strings.HasPrefix(lower, "profile=") && (strings.Contains(lower, "public") || strings.Contains(lower, "any")) {
				t.Errorf("%s: rule must not apply to the Public profile: %s", r.Name, arg)
			}
		}
	}
}

func TestRuleNamesDoNotCollideWithLegacy(t *testing.T) {
	for _, legacy := range LegacyRuleNames {
		for _, r := range Rules(1, 2) {
			// netsh matches name= exactly; a shared name would make the 2.x
			// helper delete or report the 1.x rule as its own.
			if strings.EqualFold(legacy, r.Name) {
				t.Fatalf("2.x rule reuses the legacy name %q", legacy)
			}
		}
	}
	if got := DeleteRuleArgs("X"); strings.Join(got, " ") != "advfirewall firewall delete rule name=X" {
		t.Fatalf("delete args: %q", got)
	}
	if got := ShowRuleArgs("X"); strings.Join(got, " ") != "advfirewall firewall show rule name=X" {
		t.Fatalf("show args: %q", got)
	}
}
