package impl

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	"github.com/gfunc/subconvergo/parser/sub"
	pc "github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Security tests for the parse-boundary sanitization of
// subscription-controlled fields in text output and the guard on
// source-provided group regex filters.
//
// Seam under test: crafted subscription content parsed through parser/sub and
// rendered through the public Generator interface — no internal hooks.

var universalLineSplit = regexp.MustCompile("\r\n|\r|\n")

// splitOutputLines splits generated text config on any line-break convention
// (\n, \r\n, or bare \r) so CR-based injection cannot hide from assertions.
func splitOutputLines(output string) []string {
	return universalLineSplit.Split(output, -1)
}

// assertNoInjectedLines fails if any output line is a bare injected INI section
// header or rewrite rule, i.e. if a subscription-controlled field managed to
// create a NEW line in the output.
func assertNoInjectedLines(t *testing.T, output string) {
	t.Helper()
	for _, line := range splitOutputLines(output) {
		trimmed := strings.TrimSpace(line)
		assert.NotEqual(t, "[rewrite_local]", trimmed, "injected INI section found in output:\n%s", output)
		assert.False(t, strings.HasPrefix(trimmed, "^https://attacker"), "injected rewrite rule found on its own line:\n%s", output)
		assert.False(t, strings.HasPrefix(trimmed, "^https://evil"), "injected rewrite rule found on its own line:\n%s", output)
	}
}

// exploitSSLink builds an exploit link:
// an ss:// node whose UrlDecoded fragment contains CR/LF config directives.
func exploitSSLink(t *testing.T, fragment string) string {
	t.Helper()
	payload := base64.URLEncoding.EncodeToString([]byte("aes-128-gcm:pass@1.2.3.4:8388"))
	return "ss://" + payload + "#" + fragment
}

func parseSingleLinkSub(t *testing.T, link string) []pc.ProxyInterface {
	t.Helper()
	sc, err := (&sub.SingleSubscriptionParser{}).Parse(link)
	require.NoError(t, err)
	require.NotEmpty(t, sc.Proxies)
	return sc.Proxies
}

func generateText(t *testing.T, gen core.Generator, proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig, surgeVer int) string {
	t.Helper()
	opts := core.GeneratorOptions{
		Base:         "[general]",
		ProxySetting: config.ProxySetting{SurgeVer: surgeVer},
	}
	output, err := gen.Generate(proxies, groups, nil, nil, opts)
	require.NoError(t, err)
	return output
}

// Red-first cases: field sanitization

func TestSub_QuanX_RejectsNewlineInRemark(t *testing.T) {
	// The verified exploit: %0A in the ss:// fragment is UrlDecoded at
	// parser/proxy/shadowsocks.go and, pre-fix, lands raw in `tag=%s`.
	link := exploitSSLink(t, "legit%0A[rewrite_local]%0A^https://attacker%20url%20reject")
	proxies := parseSingleLinkSub(t, link)

	output := generateText(t, &QuantumultXGenerator{}, proxies, nil, 0)

	assertNoInjectedLines(t, output)
	// The legitimate node must survive sanitization (strip, not reject).
	nodeLines := 0
	for _, line := range splitOutputLines(output) {
		if strings.Contains(line, "shadowsocks=1.2.3.4:8388") {
			nodeLines++
			// Whatever remains of the remark stays glued to the node line.
			assert.Contains(t, line, "tag=legit")
		}
	}
	assert.Equal(t, 1, nodeLines, "expected exactly one node line, output:\n%s", output)
}

func TestSub_Surge_RejectsNewlineInRemark(t *testing.T) {
	for _, fragment := range []string{
		"legit%0A[rewrite_local]%0A^https://attacker%20url%20reject",
		"legit%0D[rewrite_local]%0D^https://attacker%20url%20reject",
		"legit%0D%0A[rewrite_local]%0D%0A^https://attacker%20url%20reject",
	} {
		proxies := parseSingleLinkSub(t, exploitSSLink(t, fragment))
		output := generateText(t, &SurgeGenerator{}, proxies, nil, 5)

		assertNoInjectedLines(t, output)
		nodeLines := 0
		for _, line := range splitOutputLines(output) {
			if strings.Contains(line, "ss, 1.2.3.4, 8388") {
				nodeLines++
				assert.Contains(t, line, "legit")
			}
		}
		assert.Equal(t, 1, nodeLines, "fragment %q: expected exactly one node line, output:\n%s", fragment, output)
	}
}

func TestSub_Loon_RejectsNewlineInRemark(t *testing.T) {
	for _, fragment := range []string{
		"legit%0A[rewrite_local]%0A^https://attacker%20url%20reject",
		"legit%0D[rewrite_local]%0D^https://attacker%20url%20reject",
	} {
		proxies := parseSingleLinkSub(t, exploitSSLink(t, fragment))
		output := generateText(t, &LoonGenerator{}, proxies, nil, 0)

		assertNoInjectedLines(t, output)
		nodeLines := 0
		for _, line := range splitOutputLines(output) {
			if strings.Contains(line, "Shadowsocks,1.2.3.4,8388") {
				nodeLines++
				assert.Contains(t, line, "legit")
			}
		}
		assert.Equal(t, 1, nodeLines, "fragment %q: expected exactly one node line, output:\n%s", fragment, output)
	}
}

// TestSub_TextTargets_SanitizeCredentialAndGroupFields places CR/LF payloads in
// password, host (sni) and source proxy-group name fields via a Clash-format
// subscription, then checks every line-oriented target.
func TestSub_TextTargets_SanitizeCredentialAndGroupFields(t *testing.T) {
	source := `proxies:
  - name: "cred-node"
    type: ss
    server: 1.2.3.4
    port: 8388
    cipher: aes-128-gcm
    password: "pw\n[rewrite_local]\n^https://evil url reject"
  - name: "host-node"
    type: trojan
    server: 2.3.4.5
    port: 443
    password: pass
    sni: "evil.com\n[rewrite_local]\n^https://evil url reject"
proxy-groups:
  - name: "grp\n[rewrite_local]\n^https://evil url reject"
    type: select
    proxies:
      - "cred-node"
`
	sc, err := sub.ParseMihomoConfig(source)
	require.NoError(t, err)
	require.Len(t, sc.Proxies, 2)
	require.Len(t, sc.Groups, 1)

	targets := []struct {
		name string
		gen  core.Generator
	}{
		{"quanx", &QuantumultXGenerator{}},
		{"surge", &SurgeGenerator{}},
		{"loon", &LoonGenerator{}},
	}
	for _, target := range targets {
		t.Run(target.name, func(t *testing.T) {
			output := generateText(t, target.gen, sc.Proxies, sc.Groups, 5)
			assertNoInjectedLines(t, output)
			// Nodes and the group survive in sanitized form.
			assert.Contains(t, output, "cred-node")
			assert.Contains(t, output, "host-node")
			foundGroup := false
			for _, line := range splitOutputLines(output) {
				if strings.Contains(line, "grp") && (strings.Contains(line, "static=") || strings.Contains(line, "select")) {
					foundGroup = true
				}
			}
			assert.True(t, foundGroup, "group must be preserved (stripped, not dropped); output:\n%s", output)
		})
	}
}

// Green path: legitimate remarks/group names containing commas, `=`, quotes
// and unicode must render unmangled on a single line per node.
func TestSub_TextTargets_LegitNamesRenderCorrectly(t *testing.T) {
	remark := `HK 节点 01, "vip" = premium 🚀`
	link := exploitSSLink(t, url.QueryEscape(remark))
	proxies := parseSingleLinkSub(t, link)

	groups := []config.ProxyGroupConfig{
		{Name: `选择 Group, "A" = 1`, Type: "select", Rule: []string{".*"}},
	}

	targets := []struct {
		name     string
		gen      core.Generator
		nodeMark string
	}{
		{"quanx", &QuantumultXGenerator{}, "shadowsocks=1.2.3.4:8388"},
		{"surge", &SurgeGenerator{}, "ss, 1.2.3.4, 8388"},
		{"loon", &LoonGenerator{}, "Shadowsocks,1.2.3.4,8388"},
	}
	for _, target := range targets {
		t.Run(target.name, func(t *testing.T) {
			output := generateText(t, target.gen, proxies, groups, 5)
			nodeLines := 0
			for _, line := range splitOutputLines(output) {
				if strings.Contains(line, target.nodeMark) {
					nodeLines++
					assert.Contains(t, line, remark, "remark must render unmangled")
				}
			}
			assert.Equal(t, 1, nodeLines, "output:\n%s", output)
			assert.Contains(t, output, `选择 Group, "A" = 1`, "group name must render unmangled; output:\n%s", output)
		})
	}
}

// Regression: structured targets (Clash YAML, sing-box JSON) serialize safely
// already; sanitization must not drop or corrupt nodes there.
func TestSub_StructuredTargets_SanitizeWithoutCorruption(t *testing.T) {
	link := exploitSSLink(t, "legit%0A[rewrite_local]%0A^https://attacker%20url%20reject")
	proxies := parseSingleLinkSub(t, link)

	clashOut, err := (&ClashGenerator{}).Generate(proxies, nil, nil, nil, core.GeneratorOptions{
		Base:         "proxies: []\nproxy-groups: []\nrules: []",
		ProxySetting: config.ProxySetting{},
	})
	require.NoError(t, err)
	var clashResult map[string]interface{}
	require.NoError(t, yaml.Unmarshal([]byte(clashOut), &clashResult))
	clashProxies := clashResult["proxies"].([]interface{})
	require.Len(t, clashProxies, 1)
	entry := clashProxies[0].(map[string]interface{})
	name, ok := entry["name"].(string)
	require.True(t, ok)
	assert.NotContains(t, name, "\n", "clash proxy name must not contain control chars")
	assert.True(t, strings.HasPrefix(name, "legit"), "legitimate part of the remark must survive")
	assert.Equal(t, "1.2.3.4", entry["server"])
	assert.Equal(t, "pass", entry["password"])

	singboxOut, err := (&SingBoxGenerator{}).Generate(proxies, nil, nil, nil, core.GeneratorOptions{
		Base:         "{}",
		ProxySetting: config.ProxySetting{},
	})
	require.NoError(t, err)
	assert.NotContains(t, singboxOut, "[rewrite_local]\n", "sing-box output must not contain raw injected lines")
	assert.Contains(t, singboxOut, "legit")
}

// Red-first cases: group filter guard

// parseClashGroups is a small helper: parse a Clash subscription and return the
// generated Clash YAML decoded as a map.
func parseAndGenerateClash(t *testing.T, source string) map[string]interface{} {
	t.Helper()
	sc, err := sub.ParseMihomoConfig(source)
	require.NoError(t, err)
	out, err := (&ClashGenerator{}).Generate(sc.Proxies, sc.Groups, nil, nil, core.GeneratorOptions{
		Base:         "proxies: []\nproxy-groups: []\nrules: []",
		ProxySetting: config.ProxySetting{},
	})
	require.NoError(t, err)
	var result map[string]interface{}
	require.NoError(t, yaml.Unmarshal([]byte(out), &result))
	return result
}

func groupByName(t *testing.T, result map[string]interface{}, name string) map[string]interface{} {
	t.Helper()
	groups, ok := result["proxy-groups"].([]interface{})
	require.True(t, ok, "proxy-groups missing from output")
	for _, g := range groups {
		gm := g.(map[string]interface{})
		if gm["name"] == name {
			return gm
		}
	}
	return nil
}

func TestSub_Clash_StripsPathologicalGroupFilter(t *testing.T) {
	source := `proxies:
  - {name: "HK-01", type: ss, server: 1.1.1.1, port: 8388, cipher: aes-128-gcm, password: x}
  - {name: "SG-02", type: ss, server: 2.2.2.2, port: 8388, cipher: aes-128-gcm, password: x}
  - {name: "US-03", type: ss, server: 3.3.3.3, port: 8388, cipher: aes-128-gcm, password: x}
proxy-groups:
  - {name: "evil-nested", type: select, filter: "^(a+)+$"}
  - {name: "evil-alt", type: select, filter: "(a|a)*"}
  - {name: "evil-star", type: select, filter: "(a*)+"}
`
	result := parseAndGenerateClash(t, source)

	raw, err := yaml.Marshal(result)
	require.NoError(t, err)
	for _, pathological := range []string{"^(a+)+$", "(a|a)*", "(a*)+"} {
		assert.NotContains(t, string(raw), pathological, "pathological filter must not reach generated output")
	}

	// Groups are kept in a safe form rather than dropped: with the filter
	// stripped, each falls back to the unfiltered node list.
	for _, name := range []string{"evil-nested", "evil-alt", "evil-star"} {
		g := groupByName(t, result, name)
		require.NotNil(t, g, "group %s must be preserved", name)
		members := g["proxies"].([]interface{})
		assert.Contains(t, members, "HK-01")
		assert.Contains(t, members, "US-03")
	}
}

// Green path: ordinary RE2-safe filters keep working — the group
// membership reflects the filter, so the filter's effect survives unchanged.
func TestSub_Clash_SafeGroupFilterSurvives(t *testing.T) {
	source := `proxies:
  - {name: "HK-01", type: ss, server: 1.1.1.1, port: 8388, cipher: aes-128-gcm, password: x}
  - {name: "SG-02", type: ss, server: 2.2.2.2, port: 8388, cipher: aes-128-gcm, password: x}
  - {name: "US-03", type: ss, server: 3.3.3.3, port: 8388, cipher: aes-128-gcm, password: x}
proxy-groups:
  - {name: "asia", type: select, filter: "^(HK|SG)"}
  - {name: "keyword", type: select, filter: "HK"}
`
	result := parseAndGenerateClash(t, source)

	asia := groupByName(t, result, "asia")
	require.NotNil(t, asia)
	asiaMembers := asia["proxies"].([]interface{})
	assert.Contains(t, asiaMembers, "HK-01")
	assert.Contains(t, asiaMembers, "SG-02")
	assert.NotContains(t, asiaMembers, "US-03", "safe RE2 filter must still filter group members")

	keyword := groupByName(t, result, "keyword")
	require.NotNil(t, keyword)
	assert.Equal(t, []interface{}{"HK-01"}, keyword["proxies"].([]interface{}))
}

func TestSub_Clash_FilterLengthAndCountCapped(t *testing.T) {
	// One group carrying an over-long (but RE2-valid) filter, plus more groups
	// than the per-subscription cap allows.
	var sb strings.Builder
	sb.WriteString("proxies:\n")
	sb.WriteString("  - {name: \"HK-01\", type: ss, server: 1.1.1.1, port: 8388, cipher: aes-128-gcm, password: x}\n")
	sb.WriteString("  - {name: \"US-03\", type: ss, server: 3.3.3.3, port: 8388, cipher: aes-128-gcm, password: x}\n")
	sb.WriteString("proxy-groups:\n")
	longFilter := "^HK" + strings.Repeat("x{0,1}", 100) // 704 chars, RE2-valid
	fmt.Fprintf(&sb, "  - {name: \"long-filter\", type: select, filter: %q}\n", longFilter)
	const totalGroups = 150
	for i := 0; i < totalGroups; i++ {
		fmt.Fprintf(&sb, "  - {name: \"grp-%03d\", type: select, proxies: [\"HK-01\"]}\n", i)
	}

	result := parseAndGenerateClash(t, sb.String())

	groups := result["proxy-groups"].([]interface{})
	assert.LessOrEqual(t, len(groups), 100, "source-provided group count must be capped, got %d", len(groups))

	// Over-long filter is stripped: group preserved, falls back to unfiltered.
	lf := groupByName(t, result, "long-filter")
	require.NotNil(t, lf, "group with over-long filter must be preserved")
	members := lf["proxies"].([]interface{})
	assert.Contains(t, members, "HK-01")
	assert.Contains(t, members, "US-03", "over-long filter must be stripped (no filtering effect)")

	raw, err := yaml.Marshal(result)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), longFilter)
}
