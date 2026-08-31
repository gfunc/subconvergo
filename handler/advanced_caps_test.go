package handler

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
)

// TestEnforceAdvancedCaps_Rulesets pins max_allowed_rulesets truncation and
// the 0 = unlimited escape hatch.
func TestEnforceAdvancedCaps_Rulesets(t *testing.T) {
	defer func() { config.Global.Advanced.MaxAllowedRulesets = 0 }()

	rulesets := []config.RulesetConfig{{Group: "g1"}, {Group: "g2"}, {Group: "g3"}}

	config.Global.Advanced.MaxAllowedRulesets = 2
	got, _ := enforceAdvancedCaps(rulesets, nil)
	if len(got) != 2 || got[0].Group != "g1" || got[1].Group != "g2" {
		t.Fatalf("expected first 2 rulesets kept, got %#v", got)
	}

	config.Global.Advanced.MaxAllowedRulesets = 0
	got, _ = enforceAdvancedCaps(rulesets, nil)
	if len(got) != 3 {
		t.Fatalf("0 must mean unlimited, got %d rulesets", len(got))
	}
}

// TestEnforceAdvancedCaps_RawRules pins max_allowed_rules truncation and the
// 0 = unlimited escape hatch.
func TestEnforceAdvancedCaps_RawRules(t *testing.T) {
	defer func() { config.Global.Advanced.MaxAllowedRules = 0 }()

	rules := []string{"r1", "r2", "r3"}

	config.Global.Advanced.MaxAllowedRules = 1
	_, got := enforceAdvancedCaps(nil, rules)
	if len(got) != 1 || got[0] != "r1" {
		t.Fatalf("expected first rule kept, got %#v", got)
	}

	config.Global.Advanced.MaxAllowedRules = 0
	_, got = enforceAdvancedCaps(nil, rules)
	if len(got) != 3 {
		t.Fatalf("0 must mean unlimited, got %d rules", len(got))
	}
}
