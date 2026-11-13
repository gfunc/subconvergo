package impl

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/generator/core"
	pc "github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func renderOmitTestConfig(t *testing.T, proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig) map[string]interface{} {
	t.Helper()
	output, err := (&ClashGenerator{}).Generate(proxies, groups, nil, nil, core.GeneratorOptions{
		Base:         "proxies: []\nproxy-groups: []\nrules: []",
		ProxySetting: config.ProxySetting{},
	})
	require.NoError(t, err)
	var result map[string]interface{}
	require.NoError(t, yaml.Unmarshal([]byte(output), &result))
	return result
}

func omitTestGroups(t *testing.T, result map[string]interface{}) map[string]map[string]interface{} {
	t.Helper()
	groups := make(map[string]map[string]interface{})
	raw, _ := result["proxy-groups"].([]interface{})
	for _, entry := range raw {
		group := entry.(map[string]interface{})
		groups[group["name"].(string)] = group
	}
	return groups
}

func omitTestMembers(t *testing.T, group map[string]interface{}) []string {
	t.Helper()
	raw, ok := group["proxies"].([]interface{})
	require.True(t, ok, "proxies missing from group %#v", group)
	members := make([]string, 0, len(raw))
	for _, member := range raw {
		members = append(members, member.(string))
	}
	return members
}

func TestClashOmitIfEmptyKeepsNonEmptyGroup(t *testing.T) {
	proxies := parseSingleLinkSub(t, exploitSSLink(t, "brand"))
	groups := omitTestGroups(t, renderOmitTestConfig(t, proxies, []config.ProxyGroupConfig{
		{Name: "Other Backup", Type: "load-balance", Rule: []string{"^brand$"}, OmitIfEmpty: true},
		{Name: "Other", Type: "fallback", Rule: []string{"[]Other Backup"}, OmitIfEmpty: true},
	}))

	require.Len(t, groups, 2)
	assert.Equal(t, []string{"brand"}, omitTestMembers(t, groups["Other Backup"]))
	assert.Equal(t, []string{"Other Backup"}, omitTestMembers(t, groups["Other"]))
}

func TestClashOmitIfEmptyPrunesEmptyBranches(t *testing.T) {
	proxies := parseSingleLinkSub(t, exploitSSLink(t, "JP-node"))
	groups := omitTestGroups(t, renderOmitTestConfig(t, proxies, []config.ProxyGroupConfig{
		{Name: "JP Backup", Type: "load-balance", Rule: []string{"^JP"}, OmitIfEmpty: true},
		{Name: "Other Backup", Type: "load-balance", Rule: []string{"^brand$"}, OmitIfEmpty: true},
		{Name: "Other", Type: "fallback", Rule: []string{"[]Other Backup"}, OmitIfEmpty: true},
		{Name: "Backup Locations", Type: "load-balance", Rule: []string{"[]JP Backup", "[]Other Backup"}, OmitIfEmpty: true},
		{Name: "Main", Type: "select", Rule: []string{"[]Other", "[]Backup Locations"}},
	}))

	require.Len(t, groups, 3)
	assert.NotContains(t, groups, "Other Backup")
	assert.NotContains(t, groups, "Other")
	assert.Equal(t, []string{"JP Backup"}, omitTestMembers(t, groups["Backup Locations"]))
	assert.Equal(t, []string{"Backup Locations"}, omitTestMembers(t, groups["Main"]))
}

func TestClashEmptyGroupDefaultRemainsDirect(t *testing.T) {
	proxies := parseSingleLinkSub(t, exploitSSLink(t, "node"))
	groups := omitTestGroups(t, renderOmitTestConfig(t, proxies, []config.ProxyGroupConfig{
		{Name: "Legacy Empty", Type: "select", Rule: []string{"^missing$"}},
		{Name: "OptIn Empty", Type: "select", Rule: []string{"^also-missing$"}, OmitIfEmpty: true},
		{Name: "Legacy Parent", Type: "select", Rule: []string{"[]OptIn Empty"}},
	}))

	require.Len(t, groups, 2)
	assert.Equal(t, []string{"DIRECT"}, omitTestMembers(t, groups["Legacy Empty"]))
	assert.Equal(t, []string{"DIRECT"}, omitTestMembers(t, groups["Legacy Parent"]))
}

func TestClashOmitIfEmptyRejectsRelay(t *testing.T) {
	proxies := parseSingleLinkSub(t, exploitSSLink(t, "node"))
	_, err := (&ClashGenerator{}).Generate(proxies, []config.ProxyGroupConfig{
		{Name: "Relay", Type: "relay", Rule: []string{"[]DIRECT", "^node$"}, OmitIfEmpty: true},
	}, nil, nil, core.GeneratorOptions{
		Base:         "proxies: []\nproxy-groups: []\nrules: []",
		ProxySetting: config.ProxySetting{},
	})
	require.ErrorContains(t, err, "omit_if_empty is not supported for relay group")
}
