package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/ini.v1"
)

func TestLoadYAMLConfig(t *testing.T) {
	// Create a temporary YAML config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yml")

	yamlContent := `
common:
  api_mode: false
  api_access_token: "test-password"
  base_path: "testbase"
  exclude_remarks: ["过期", "流量"]
  include_remarks: ["香港", "台湾"]
  clash_rule_base: "base/clash.tpl"
  proxy_config: "SYSTEM"
  
node_pref:
  clash_use_new_field_name: true
  clash_proxies_style: "flow"
  clash_proxy_groups_style: "block"
  sort_flag: false
  append_sub_userinfo: true
  
managed_config:
  write_managed_config: true
  managed_config_prefix: "http://example.com:8080"
  config_update_interval: 3600
  config_update_strict: true
  
server:
  listen: "127.0.0.1"
  port: 8888
  
advanced:
  log_level: "debug"
  skip_failed_links: true
  max_allowed_rulesets: 100
  cache_subscription: 120
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load the config
	if err := loadYAMLConfig(configPath); err != nil {
		t.Fatalf("Failed to load YAML config: %v", err)
	}

	// Verify common section
	if Global.Common.APIMode != false {
		t.Errorf("Expected APIMode=false, got %v", Global.Common.APIMode)
	}
	if Global.Common.APIAccessToken != "test-password" {
		t.Errorf("Expected APIAccessToken=test-password, got %s", Global.Common.APIAccessToken)
	}
	if Global.Common.BasePath != "testbase" {
		t.Errorf("Expected BasePath=testbase, got %s", Global.Common.BasePath)
	}
	if len(Global.Common.ExcludeRemarks) != 2 {
		t.Errorf("Expected 2 exclude remarks, got %d", len(Global.Common.ExcludeRemarks))
	}
	if len(Global.Common.IncludeRemarks) != 2 {
		t.Errorf("Expected 2 include remarks, got %d", len(Global.Common.IncludeRemarks))
	}

	if Global.NodePref.ClashProxiesStyle != "flow" {
		t.Errorf("Expected ClashProxiesStyle=flow, got %s", Global.NodePref.ClashProxiesStyle)
	}
	if !Global.NodePref.AppendSubUserinfo {
		t.Error("Expected AppendSubUserinfo=true")
	}

	// Verify managed_config section
	if !Global.ManagedConfig.WriteManagedConfig {
		t.Error("Expected WriteManagedConfig=true")
	}
	if Global.ManagedConfig.ManagedConfigPrefix != "http://example.com:8080" {
		t.Errorf("Expected ManagedConfigPrefix=http://example.com:8080, got %s", Global.ManagedConfig.ManagedConfigPrefix)
	}
	if Global.ManagedConfig.ConfigUpdateInterval != 3600 {
		t.Errorf("Expected ConfigUpdateInterval=3600, got %d", Global.ManagedConfig.ConfigUpdateInterval)
	}
	if !Global.ManagedConfig.ConfigUpdateStrict {
		t.Error("Expected ConfigUpdateStrict=true")
	}

	// Verify server section
	if Global.Server.Listen != "127.0.0.1" {
		t.Errorf("Expected Listen=127.0.0.1, got %s", Global.Server.Listen)
	}
	if Global.Server.Port != 8888 {
		t.Errorf("Expected Port=8888, got %d", Global.Server.Port)
	}

	// Verify advanced section
	if Global.Advanced.LogLevel != "debug" {
		t.Errorf("Expected LogLevel=debug, got %s", Global.Advanced.LogLevel)
	}
	if !Global.Advanced.SkipFailedLinks {
		t.Error("Expected SkipFailedLinks=true")
	}
	if Global.Advanced.MaxAllowedRulesets != 100 {
		t.Errorf("Expected MaxAllowedRulesets=100, got %d", Global.Advanced.MaxAllowedRulesets)
	}
}

func TestLoadTOMLConfig(t *testing.T) {
	// Reset global config
	Global = &Settings{}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	tomlContent := `
[common]
api_mode = true
api_access_token = "toml-token"
base_path = "toml-base"
exclude_remarks = ["测试", "过期"]
proxy_config = "NONE"
clash_rule_base = "toml/clash.tpl"

[node_pref]
clash_use_new_field_name = false
sort_flag = true
filter_deprecated_nodes = true
singbox_add_clash_modes = false

[[node_pref.rename_node]]
match = "HK"
replace = "Hong Kong"

[managed_config]
write_managed_config = false
managed_config_prefix = "https://toml.example.com"
config_update_interval = 7200

[server]
listen = "0.0.0.0"
port = 9000

[advanced]
log_level = "info"
max_concurrent_threads = 4
enable_cache = true
cache_ruleset = 43200
`

	if err := os.WriteFile(configPath, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	if err := loadTOMLConfig(configPath); err != nil {
		t.Fatalf("Failed to load TOML config: %v", err)
	}

	// Verify common section
	if !Global.Common.APIMode {
		t.Error("Expected APIMode=true")
	}
	if Global.Common.APIAccessToken != "toml-token" {
		t.Errorf("Expected APIAccessToken=toml-token, got %s", Global.Common.APIAccessToken)
	}
	if Global.Common.BasePath != "toml-base" {
		t.Errorf("Expected BasePath=toml-base, got %s", Global.Common.BasePath)
	}
	if Global.Common.ProxyConfig != "NONE" {
		t.Errorf("Expected ProxyConfig=NONE, got %s", Global.Common.ProxyConfig)
	}

	if !Global.NodePref.SortFlag {
		t.Error("Expected SortFlag=true")
	}
	if !Global.NodePref.FilterDeprecatedNodes {
		t.Error("Expected FilterDeprecatedNodes=true")
	}
	if Global.NodePref.SingBoxAddClashModes {
		t.Error("Expected SingBoxAddClashModes=false")
	}

	// Verify rename rules
	if len(Global.NodePref.RenameNodes) != 1 {
		t.Errorf("Expected 1 rename node, got %d", len(Global.NodePref.RenameNodes))
	} else {
		if Global.NodePref.RenameNodes[0].Match != "HK" {
			t.Errorf("Expected Match=HK, got %s", Global.NodePref.RenameNodes[0].Match)
		}
		if Global.NodePref.RenameNodes[0].Replace != "Hong Kong" {
			t.Errorf("Expected Replace='Hong Kong', got %s", Global.NodePref.RenameNodes[0].Replace)
		}
	}

	// Verify managed config
	if Global.ManagedConfig.WriteManagedConfig {
		t.Error("Expected WriteManagedConfig=false")
	}
	if Global.ManagedConfig.ConfigUpdateInterval != 7200 {
		t.Errorf("Expected ConfigUpdateInterval=7200, got %d", Global.ManagedConfig.ConfigUpdateInterval)
	}

	// Verify server
	if Global.Server.Port != 9000 {
		t.Errorf("Expected Port=9000, got %d", Global.Server.Port)
	}

	// Verify advanced
	if Global.Advanced.MaxConcurrentThreads != 4 {
		t.Errorf("Expected MaxConcurrentThreads=4, got %d", Global.Advanced.MaxConcurrentThreads)
	}
	if !Global.Advanced.EnableCache {
		t.Error("Expected EnableCache=true")
	}
}

func TestLoadINIConfig(t *testing.T) {
	// Reset global config
	Global = &Settings{}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.ini")

	iniContent := `
[common]
api_mode=true
api_access_token=ini-token
base_path=ini-base
exclude_remarks=(过期|流量|时间)
include_remarks=HK|TW|US
proxy_subscription=SYSTEM
clash_rule_base=ini/clash.tpl

[userinfo]
stream_rule=^剩余流量：(.*?)\\|总流量：(.*)$|total=$2&left=$1
stream_rule=^Bandwidth: (.*?)/(.*)$|used=$1&total=$2
time_rule=^过期时间：(\\d+)-(\\d+)-(\\d+) (\\d+):(\\d+):(\\d+)$|$1:$2:$3:$4:$5:$6

[node_pref]
clash_use_new_field_name=true
clash_proxies_style=compact
append_sub_userinfo=false
rename_node=HK@香港
rename_node=!!import:../base/snippets/rename_node.txt

[managed_config]
write_managed_config=true
managed_config_prefix=http://ini.example.com
config_update_interval=1800
config_update_strict=false

[surge_external_proxy]
resolve_hostname=false

[emojis]
add_emoji=true
remove_old_emoji=false
rule=(香港|HK),🇭🇰
rule=!!import:../base/snippets/emoji.txt

[rulesets]
enabled=true
overwrite_original_rules=true
update_ruleset_on_request=false
ruleset=DIRECT,rules/local.list
ruleset=Proxy,surge:https://example.com/rules/proxy.list,3600
ruleset=!!import:../base/snippets/rulesets.txt

[proxy_groups]
custom_proxy_group=Auto` + "`" + `url-test` + "`" + `.*` + "`" + `http://www.gstatic.com/generate_204` + "`" + `300
custom_proxy_group=Proxy` + "`" + `select` + "`" + `.*` + "`" + `[]DIRECT
custom_proxy_group=!!import:../base/snippets/groups.txt

[template]
template_path=/path/to/templates
clash.http_port=7890
clash.socks_port=7891

[aliases]
/v=/version
/clash=/sub?target=clash

[tasks]
task=refresh` + "`" + `0 */6 * * *` + "`" + `refresh.js` + "`" + `10

[server]
listen=0.0.0.0
port=7777

[advanced]
log_level=warn
print_debug_info=true
max_allowed_rules=50000
skip_failed_links=false
`

	if err := os.WriteFile(configPath, []byte(iniContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	if err := loadINIConfig(configPath); err != nil {
		t.Fatalf("Failed to load INI config: %v", err)
	}

	// Verify common section
	if !Global.Common.APIMode {
		t.Error("Expected APIMode=true")
	}
	if Global.Common.APIAccessToken != "ini-token" {
		t.Errorf("Expected APIAccessToken=ini-token, got %s", Global.Common.APIAccessToken)
	}
	if Global.Common.BasePath != "ini-base" {
		t.Errorf("Expected BasePath=ini-base, got %s", Global.Common.BasePath)
	}

	// Verify userinfo section
	if len(Global.UserInfo.StreamRules) < 2 {
		t.Errorf("Expected at least 2 stream rules, got %d", len(Global.UserInfo.StreamRules))
	}
	if len(Global.UserInfo.TimeRules) < 1 {
		t.Errorf("Expected at least 1 time rule, got %d", len(Global.UserInfo.TimeRules))
	}

	// Verify node_pref
	if Global.NodePref.ClashProxiesStyle != "compact" {
		t.Errorf("Expected ClashProxiesStyle=compact, got %s", Global.NodePref.ClashProxiesStyle)
	}
	if Global.NodePref.AppendSubUserinfo {
		t.Error("Expected AppendSubUserinfo=false")
	}

	// Verify rename rules parsing
	if len(Global.NodePref.RenameNodes) < 2 {
		t.Errorf("Expected at least 2 rename rules, got %d", len(Global.NodePref.RenameNodes))
	} else {
		// Check first rename rule
		if Global.NodePref.RenameNodes[0].Match != "HK" {
			t.Errorf("Expected first rename match=HK, got %s", Global.NodePref.RenameNodes[0].Match)
		}
		if Global.NodePref.RenameNodes[0].Replace != "香港" {
			t.Errorf("Expected first rename replace=香港, got %s", Global.NodePref.RenameNodes[0].Replace)
		}
		// Check import rule
		if Global.NodePref.RenameNodes[1].Match != `\(?((x|X)?(\d+)(\.?\d+)?)((\s?倍率?)|(x|X))\)?` &&
			Global.NodePref.RenameNodes[1].Replace != "$1x" {
			t.Errorf(`Expected Match=\(?((x|X)?(\d+)(\.?\d+)?)((\s?倍率?)|(x|X))\)? Replace=$1x, got %s %s`, Global.NodePref.RenameNodes[1].Match, Global.NodePref.RenameNodes[1].Replace)
		}
	}

	// Verify managed config
	if !Global.ManagedConfig.WriteManagedConfig {
		t.Error("Expected WriteManagedConfig=true")
	}
	if Global.ManagedConfig.ConfigUpdateInterval != 1800 {
		t.Errorf("Expected ConfigUpdateInterval=1800, got %d", Global.ManagedConfig.ConfigUpdateInterval)
	}
	if Global.ManagedConfig.ConfigUpdateStrict {
		t.Error("Expected ConfigUpdateStrict=false")
	}

	// Verify surge external
	if Global.SurgeExternal.ResolveHostname {
		t.Error("Expected ResolveHostname=false")
	}

	// Verify emojis
	if !Global.Emojis.AddEmoji {
		t.Error("Expected AddEmoji=true")
	}
	if Global.Emojis.RemoveOldEmoji {
		t.Error("Expected RemoveOldEmoji=false")
	}
	if len(Global.Emojis.Rules) < 2 {
		t.Errorf("Expected at least 2 emoji rules, got %d", len(Global.Emojis.Rules))
	}

	// Verify rulesets
	if !Global.Rulesets.Enabled {
		t.Error("Expected Rulesets.Enabled=true")
	}
	if !Global.Rulesets.OverwriteOriginalRules {
		t.Error("Expected OverwriteOriginalRules=true")
	}
	if len(Global.Rulesets.Rulesets) < 3 {
		t.Errorf("Expected at least 3 rulesets, got %d", len(Global.Rulesets.Rulesets))
	}

	// Verify proxy groups
	if len(Global.ProxyGroups.CustomProxyGroups) < 3 {
		t.Errorf("Expected at least 3 proxy groups, got %d", len(Global.ProxyGroups.CustomProxyGroups))
	}

	// Verify template
	if Global.Template.TemplatePath != "/path/to/templates" {
		t.Errorf("Expected TemplatePath=/path/to/templates, got %s", Global.Template.TemplatePath)
	}
	if len(Global.Template.Globals) < 2 {
		t.Errorf("Expected at least 2 template globals, got %d", len(Global.Template.Globals))
	}

	// Verify aliases
	if len(Global.Aliases) < 2 {
		t.Errorf("Expected at least 2 aliases, got %d", len(Global.Aliases))
	}

	// Verify tasks
	if len(Global.Tasks) < 1 {
		t.Errorf("Expected at least 1 task, got %d", len(Global.Tasks))
	}

	// Verify server
	if Global.Server.Port != 7777 {
		t.Errorf("Expected Port=7777, got %d", Global.Server.Port)
	}

	// Verify advanced
	if Global.Advanced.LogLevel != "warn" {
		t.Errorf("Expected LogLevel=warn, got %s", Global.Advanced.LogLevel)
	}
	if !Global.Advanced.PrintDebugInfo {
		t.Error("Expected PrintDebugInfo=true")
	}
	if Global.Advanced.MaxAllowedRules != 50000 {
		t.Errorf("Expected MaxAllowedRules=50000, got %d", Global.Advanced.MaxAllowedRules)
	}
	if Global.Advanced.SkipFailedLinks {
		t.Error("Expected SkipFailedLinks=false")
	}
}

func TestGetBasePath(t *testing.T) {
	// Reset global
	Global = &Settings{}
	Global.Common.BasePath = "testbase"

	basePath := GetBasePath()
	if !filepath.IsAbs(basePath) {
		t.Error("Expected absolute path from GetBasePath()")
	}
	if filepath.Base(basePath) != "testbase" {
		t.Errorf("Expected base path to end with 'testbase', got %s", basePath)
	}

	// Test with absolute path
	Global.Common.BasePath = "/absolute/path"
	basePath = GetBasePath()
	if basePath != "/absolute/path" {
		t.Errorf("Expected /absolute/path, got %s", basePath)
	}
}

func TestINIRulesetParsing(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected RulesetConfig
	}{
		{
			name:  "basic ruleset",
			value: "DIRECT,rules/local.list",
			expected: RulesetConfig{
				Group:   "DIRECT",
				Ruleset: "rules/local.list",
			},
		},
		{
			name:  "ruleset with type prefix",
			value: "Proxy,surge:https://example.com/proxy.list,3600",
			expected: RulesetConfig{
				Group:    "Proxy",
				Type:     "surge",
				Ruleset:  "https://example.com/proxy.list",
				Interval: 3600,
			},
		},
		{
			name:  "rule (not ruleset)",
			value: "DIRECT,[]GEOIP,CN",
			expected: RulesetConfig{
				Group: "DIRECT",
				Rule:  "[]GEOIP,CN",
			},
		},
		{
			name:  "import directive",
			value: "!!import:../base/snippets/rules.txt",
			expected: RulesetConfig{
				Import: "../base/snippets/rules.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock INI section
			tmpDir := t.TempDir()
			iniPath := filepath.Join(tmpDir, "test.ini")
			iniContent := "[test]\nruleset=" + tt.value
			os.WriteFile(iniPath, []byte(iniContent), 0644)

			cfg, _ := ini.Load(iniPath)
			sec := cfg.Section("test")
			results := parseINIRulesets(sec, "ruleset")

			if len(results) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(results))
			}

			result := results[0]
			if result.Group != tt.expected.Group {
				t.Errorf("Expected Group=%s, got %s", tt.expected.Group, result.Group)
			}
			if result.Type != tt.expected.Type {
				t.Errorf("Expected Type=%s, got %s", tt.expected.Type, result.Type)
			}
			if result.Ruleset != tt.expected.Ruleset {
				t.Errorf("Expected Ruleset=%s, got %s", tt.expected.Ruleset, result.Ruleset)
			}
			if result.Rule != tt.expected.Rule {
				t.Errorf("Expected Rule=%s, got %s", tt.expected.Rule, result.Rule)
			}
			if result.Import != tt.expected.Import {
				t.Errorf("Expected Import=%s, got %s", tt.expected.Import, result.Import)
			}
			if tt.expected.Interval > 0 && result.Interval != tt.expected.Interval {
				t.Errorf("Expected Interval=%d, got %d", tt.expected.Interval, result.Interval)
			}
		})
	}
}

func TestINIProxyGroupParsing(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected ProxyGroupConfig
	}{
		{
			name:  "select group",
			value: "Proxy`select`.*`[]DIRECT",
			expected: ProxyGroupConfig{
				Name: "Proxy",
				Type: "select",
				Rule: []string{".*", "[]DIRECT"},
			},
		},
		{
			name:  "url-test group",
			value: "Auto`url-test`.*`http://www.gstatic.com/generate_204`300,5,100",
			expected: ProxyGroupConfig{
				Name:      "Auto",
				Type:      "url-test",
				Rule:      []string{".*"},
				URL:       "http://www.gstatic.com/generate_204",
				Interval:  300,
				Timeout:   5,
				Tolerance: 100,
			},
		},
		{
			name:  "import directive",
			value: "!!import:../base/snippets/groups.txt",
			expected: ProxyGroupConfig{
				Import: "../base/snippets/groups.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			iniPath := filepath.Join(tmpDir, "test.ini")
			iniContent := "[test]\ncustom_proxy_group=" + tt.value
			os.WriteFile(iniPath, []byte(iniContent), 0644)

			cfg, _ := ini.Load(iniPath)
			sec := cfg.Section("test")
			results := parseINIProxyGroups(sec, "custom_proxy_group")

			if len(results) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(results))
			}

			result := results[0]
			if result.Name != tt.expected.Name {
				t.Errorf("Expected Name=%s, got %s", tt.expected.Name, result.Name)
			}
			if result.Type != tt.expected.Type {
				t.Errorf("Expected Type=%s, got %s", tt.expected.Type, result.Type)
			}
			if result.Import != tt.expected.Import {
				t.Errorf("Expected Import=%s, got %s", tt.expected.Import, result.Import)
			}
			if tt.expected.URL != "" && result.URL != tt.expected.URL {
				t.Errorf("Expected URL=%s, got %s", tt.expected.URL, result.URL)
			}
			if tt.expected.Interval > 0 && result.Interval != tt.expected.Interval {
				t.Errorf("Expected Interval=%d, got %d", tt.expected.Interval, result.Interval)
			}
		})
	}
}
