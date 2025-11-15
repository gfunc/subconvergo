package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
)

// Settings represents the complete application configuration
type Settings struct {
	Common         CommonConfig         `yaml:"common" toml:"common" ini:"common"`
	UserInfo       UserInfoConfig       `yaml:"userinfo" toml:"userinfo" ini:"userinfo"`
	NodePref       NodePrefConfig       `yaml:"node_pref" toml:"node_pref" ini:"node_pref"`
	ManagedConfig  ManagedConfigSection `yaml:"managed_config" toml:"managed_config" ini:"managed_config"`
	SurgeExternal  SurgeExternalConfig  `yaml:"surge_external_proxy" toml:"surge_external_proxy" ini:"surge_external_proxy"`
	Emojis         EmojiConfig          `yaml:"emojis" toml:"emojis" ini:"emojis"`
	Rulesets       RulesetSection       `yaml:"rulesets" toml:"ruleset" ini:"rulesets"`
	ProxyGroups    ProxyGroupSection    `yaml:"proxy_groups" toml:"proxy_groups" ini:"proxy_groups"`
	CustomGroups   []ProxyGroupConfig   `yaml:"custom_proxy_group" toml:"custom_groups" ini:"-"`
	CustomRulesets []RulesetConfig      `yaml:"ruleset" toml:"rulesets" ini:"-"`
	Template       TemplateConfig       `yaml:"template" toml:"template" ini:"template"`
	Aliases        []AliasConfig        `yaml:"aliases" toml:"aliases" ini:"aliases"`
	Tasks          []TaskConfig         `yaml:"tasks" toml:"tasks" ini:"tasks"`
	Server         ServerConfig         `yaml:"server" toml:"server" ini:"server"`
	Advanced       AdvancedConfig       `yaml:"advanced" toml:"advanced" ini:"advanced"`
}

// CommonConfig represents the [common] section
type CommonConfig struct {
	APIMode               bool     `yaml:"api_mode" toml:"api_mode" ini:"api_mode"`
	APIAccessToken        string   `yaml:"api_access_token" toml:"api_access_token" ini:"api_access_token"`
	DefaultURL            []string `yaml:"default_url" toml:"default_url" ini:"default_url,omitempty,allowshadow"`
	EnableInsert          bool     `yaml:"enable_insert" toml:"enable_insert" ini:"enable_insert"`
	InsertURL             []string `yaml:"insert_url" toml:"insert_url" ini:"insert_url,omitempty,allowshadow"`
	PrependInsertURL      bool     `yaml:"prepend_insert_url" toml:"prepend_insert_url" ini:"prepend_insert_url"`
	ExcludeRemarks        []string `yaml:"exclude_remarks" toml:"exclude_remarks" ini:"exclude_remarks,omitempty,allowshadow"`
	IncludeRemarks        []string `yaml:"include_remarks" toml:"include_remarks" ini:"include_remarks,omitempty,allowshadow"`
	EnableFilter          bool     `yaml:"enable_filter" toml:"enable_filter" ini:"enable_filter"`
	FilterScript          string   `yaml:"filter_script" toml:"filter_script" ini:"filter_script"`
	DefaultExternalConfig string   `yaml:"default_external_config" toml:"default_external_config" ini:"default_external_config"`
	BasePath              string   `yaml:"base_path" toml:"base_path" ini:"base_path"`
	ClashRuleBase         string   `yaml:"clash_rule_base" toml:"clash_rule_base" ini:"clash_rule_base"`
	SurgeRuleBase         string   `yaml:"surge_rule_base" toml:"surge_rule_base" ini:"surge_rule_base"`
	SurfboardRuleBase     string   `yaml:"surfboard_rule_base" toml:"surfboard_rule_base" ini:"surfboard_rule_base"`
	MellowRuleBase        string   `yaml:"mellow_rule_base" toml:"mellow_rule_base" ini:"mellow_rule_base"`
	QuanRuleBase          string   `yaml:"quan_rule_base" toml:"quan_rule_base" ini:"quan_rule_base"`
	QuanXRuleBase         string   `yaml:"quanx_rule_base" toml:"quanx_rule_base" ini:"quanx_rule_base"`
	LoonRuleBase          string   `yaml:"loon_rule_base" toml:"loon_rule_base" ini:"loon_rule_base"`
	SSSubRuleBase         string   `yaml:"sssub_rule_base" toml:"sssub_rule_base" ini:"sssub_rule_base"`
	SingBoxRuleBase       string   `yaml:"singbox_rule_base" toml:"singbox_rule_base" ini:"singbox_rule_base"`
	ProxyConfig           string   `yaml:"proxy_config" toml:"proxy_config" ini:"proxy_config"`
	ProxyRuleset          string   `yaml:"proxy_ruleset" toml:"proxy_ruleset" ini:"proxy_ruleset"`
	ProxySubscription     string   `yaml:"proxy_subscription" toml:"proxy_subscription" ini:"proxy_subscription"`
	AppendProxyType       bool     `yaml:"append_proxy_type" toml:"append_proxy_type" ini:"append_proxy_type"`
	ReloadConfOnRequest   bool     `yaml:"reload_conf_on_request" toml:"reload_conf_on_request" ini:"reload_conf_on_request"`
}

// UserInfoConfig represents the [userinfo] section
type UserInfoConfig struct {
	StreamRules []RegexRuleConfig `yaml:"stream_rule" toml:"stream_rule" ini:"stream_rule,,,allowshadow"`
	TimeRules   []RegexRuleConfig `yaml:"time_rule" toml:"time_rule" ini:"time_rule,,,allowshadow"`
}

// RegexRuleConfig represents a regex match/replace rule
type RegexRuleConfig struct {
	Match   string `yaml:"match" toml:"match" ini:"match"`
	Replace string `yaml:"replace" toml:"replace" ini:"replace"`
}

// NodePrefConfig represents the [node_pref] section
type NodePrefConfig struct {
	UDPFlag               *bool              `yaml:"udp_flag,omitempty" toml:"udp_flag" ini:"udp_flag"`
	TCPFastOpenFlag       *bool              `yaml:"tcp_fast_open_flag,omitempty" toml:"tcp_fast_open_flag" ini:"tcp_fast_open_flag"`
	SkipCertVerifyFlag    *bool              `yaml:"skip_cert_verify_flag,omitempty" toml:"skip_cert_verify_flag" ini:"skip_cert_verify_flag"`
	TLS13Flag             *bool              `yaml:"tls13_flag,omitempty" toml:"tls13_flag" ini:"tls13_flag"`
	SortFlag              bool               `yaml:"sort_flag" toml:"sort_flag" ini:"sort_flag"`
	SortScript            string             `yaml:"sort_script" toml:"sort_script" ini:"sort_script"`
	FilterDeprecatedNodes bool               `yaml:"filter_deprecated_nodes" toml:"filter_deprecated_nodes" ini:"filter_deprecated_nodes"`
	AppendSubUserinfo     bool               `yaml:"append_sub_userinfo" toml:"append_sub_userinfo" ini:"append_sub_userinfo"`
	ClashProxiesStyle     string             `yaml:"clash_proxies_style" toml:"clash_proxies_style" ini:"clash_proxies_style"`
	ClashProxyGroupsStyle string             `yaml:"clash_proxy_groups_style" toml:"clash_proxy_groups_style" ini:"clash_proxy_groups_style"`
	SingBoxAddClashModes  bool               `yaml:"singbox_add_clash_modes" toml:"singbox_add_clash_modes" ini:"singbox_add_clash_modes"`
	RenameNodes           []RenameNodeConfig `yaml:"rename_node" toml:"rename_node" ini:"rename_node,,,allowshadow"`
}

// RenameNodeConfig represents node rename rules
type RenameNodeConfig struct {
	Match   string `yaml:"match,omitempty" toml:"match" ini:"match"`
	Replace string `yaml:"replace,omitempty" toml:"replace" ini:"replace"`
	Script  string `yaml:"script,omitempty" toml:"script" ini:"script"`
	Import  string `yaml:"import,omitempty" toml:"import" ini:"import"`
}

// ManagedConfigSection represents the [managed_config] section
type ManagedConfigSection struct {
	WriteManagedConfig   bool   `yaml:"write_managed_config" toml:"write_managed_config" ini:"write_managed_config"`
	ManagedConfigPrefix  string `yaml:"managed_config_prefix" toml:"managed_config_prefix" ini:"managed_config_prefix"`
	ConfigUpdateInterval int    `yaml:"config_update_interval" toml:"config_update_interval" ini:"config_update_interval"`
	ConfigUpdateStrict   bool   `yaml:"config_update_strict" toml:"config_update_strict" ini:"config_update_strict"`
	QuanXDeviceID        string `yaml:"quanx_device_id" toml:"quanx_device_id" ini:"quanx_device_id"`
}

// SurgeExternalConfig represents the [surge_external_proxy] section
type SurgeExternalConfig struct {
	SurgeSSRPath    string `yaml:"surge_ssr_path" toml:"surge_ssr_path" ini:"surge_ssr_path"`
	ResolveHostname bool   `yaml:"resolve_hostname" toml:"resolve_hostname" ini:"resolve_hostname"`
}

// EmojiConfig represents the [emojis] section
type EmojiConfig struct {
	AddEmoji       bool              `yaml:"add_emoji" toml:"add_emoji" ini:"add_emoji"`
	RemoveOldEmoji bool              `yaml:"remove_old_emoji" toml:"remove_old_emoji" ini:"remove_old_emoji"`
	Rules          []EmojiRuleConfig `yaml:"rules" toml:"emoji" ini:"rule,,,allowshadow"`
}

// EmojiRuleConfig represents an emoji rule
type EmojiRuleConfig struct {
	Match  string `yaml:"match,omitempty" toml:"match" ini:"match"`
	Emoji  string `yaml:"emoji,omitempty" toml:"emoji" ini:"emoji"`
	Script string `yaml:"script,omitempty" toml:"script" ini:"script"`
	Import string `yaml:"import,omitempty" toml:"import" ini:"import"`
}

// RulesetSection represents the [rulesets] section
type RulesetSection struct {
	Enabled                bool            `yaml:"enabled" toml:"enabled" ini:"enabled"`
	OverwriteOriginalRules bool            `yaml:"overwrite_original_rules" toml:"overwrite_original_rules" ini:"overwrite_original_rules"`
	UpdateRulesetOnRequest bool            `yaml:"update_ruleset_on_request" toml:"update_ruleset_on_request" ini:"update_ruleset_on_request"`
	Rulesets               []RulesetConfig `yaml:"rulesets" toml:"rulesets" ini:"ruleset,,,allowshadow"`
}

// RulesetConfig represents a ruleset configuration
type RulesetConfig struct {
	Group    string `yaml:"group,omitempty" toml:"group" ini:"group"`
	Ruleset  string `yaml:"ruleset,omitempty" toml:"ruleset" ini:"ruleset"`
	Rule     string `yaml:"rule,omitempty" toml:"rule" ini:"rule"`
	Type     string `yaml:"type,omitempty" toml:"type" ini:"type"`
	Interval int    `yaml:"interval,omitempty" toml:"interval" ini:"interval"`
	Import   string `yaml:"import,omitempty" toml:"import" ini:"import"`
}

// ProxyGroupSection represents the [proxy_groups] section
type ProxyGroupSection struct {
	CustomProxyGroups []ProxyGroupConfig `yaml:"custom_proxy_group" toml:"custom_groups" ini:"custom_proxy_group,,,allowshadow"`
}

// ProxyGroupConfig represents a proxy group configuration
type ProxyGroupConfig struct {
	Name       string   `yaml:"name,omitempty" toml:"name" ini:"name"`
	Type       string   `yaml:"type,omitempty" toml:"type" ini:"type"`
	Rule       []string `yaml:"rule,omitempty" toml:"rule" ini:"rule"`
	URL        string   `yaml:"url,omitempty" toml:"url" ini:"url"`
	Interval   int      `yaml:"interval,omitempty" toml:"interval" ini:"interval"`
	Tolerance  int      `yaml:"tolerance,omitempty" toml:"tolerance" ini:"tolerance"`
	Timeout    int      `yaml:"timeout,omitempty" toml:"timeout" ini:"timeout"`
	Strategy   string   `yaml:"strategy,omitempty" toml:"strategy" ini:"strategy"`
	Lazy       *bool    `yaml:"lazy,omitempty" toml:"lazy" ini:"lazy"`
	DisableUDP *bool    `yaml:"disable_udp,omitempty" toml:"disable_udp" ini:"disable_udp"`
	Import     string   `yaml:"import,omitempty" toml:"import" ini:"import"`
	Script     string   `yaml:"script,omitempty" toml:"script" ini:"script"`
	Proxies    []string `yaml:"-" toml:"-" ini:"-" json:"proxies"` // helper fileld for mihomo clash meta config convertion
}

// TemplateConfig represents the [template] section
type TemplateConfig struct {
	TemplatePath string                 `yaml:"template_path" toml:"template_path" ini:"template_path"`
	Globals      []TemplateGlobalConfig `yaml:"globals" toml:"globals" ini:"-"`
}

// TemplateGlobalConfig represents a template global variable
type TemplateGlobalConfig struct {
	Key   string `yaml:"key" toml:"key" ini:"key"`
	Value string `yaml:"value" toml:"value" ini:"value"`
}

// AliasConfig represents an alias configuration
type AliasConfig struct {
	URI    string `yaml:"uri" toml:"uri" ini:"uri"`
	Target string `yaml:"target" toml:"target" ini:"target"`
}

// TaskConfig represents a task configuration
type TaskConfig struct {
	Name    string `yaml:"name" toml:"name" ini:"name"`
	CronExp string `yaml:"cronexp" toml:"cronexp" ini:"cronexp"`
	Path    string `yaml:"path" toml:"path" ini:"path"`
	Timeout int    `yaml:"timeout" toml:"timeout" ini:"timeout"`
}

// ServerConfig represents the [server] section
type ServerConfig struct {
	Listen        string `yaml:"listen" toml:"listen" ini:"listen"`
	Port          int    `yaml:"port" toml:"port" ini:"port"`
	ServeFileRoot string `yaml:"serve_file_root" toml:"serve_file_root" ini:"serve_file_root"`
}

// AdvancedConfig represents the [advanced] section
type AdvancedConfig struct {
	LogLevel               string `yaml:"log_level" toml:"log_level" ini:"log_level"`
	PrintDebugInfo         bool   `yaml:"print_debug_info" toml:"print_debug_info" ini:"print_debug_info"`
	MaxPendingConnections  int    `yaml:"max_pending_connections" toml:"max_pending_connections" ini:"max_pending_connections"`
	MaxConcurrentThreads   int    `yaml:"max_concurrent_threads" toml:"max_concurrent_threads" ini:"max_concurrent_threads"`
	MaxAllowedRulesets     int    `yaml:"max_allowed_rulesets" toml:"max_allowed_rulesets" ini:"max_allowed_rulesets"`
	MaxAllowedRules        int    `yaml:"max_allowed_rules" toml:"max_allowed_rules" ini:"max_allowed_rules"`
	MaxAllowedDownloadSize int    `yaml:"max_allowed_download_size" toml:"max_allowed_download_size" ini:"max_allowed_download_size"`
	EnableCache            bool   `yaml:"enable_cache" toml:"enable_cache" ini:"enable_cache"`
	CacheSubscription      int    `yaml:"cache_subscription" toml:"cache_subscription" ini:"cache_subscription"`
	CacheConfig            int    `yaml:"cache_config" toml:"cache_config" ini:"cache_config"`
	CacheRuleset           int    `yaml:"cache_ruleset" toml:"cache_ruleset" ini:"cache_ruleset"`
	ScriptCleanContext     bool   `yaml:"script_clean_context" toml:"script_clean_context" ini:"script_clean_context"`
	AsyncFetchRuleset      bool   `yaml:"async_fetch_ruleset" toml:"async_fetch_ruleset" ini:"async_fetch_ruleset"`
	SkipFailedLinks        bool   `yaml:"skip_failed_links" toml:"skip_failed_links" ini:"skip_failed_links"`
}

// Global settings instance
var Global = &Settings{}

func init() {
	// Set default values
	Global.Common.APIMode = true
	Global.Common.BasePath = "base"
	Global.Common.ProxyConfig = "SYSTEM"
	Global.Common.ProxyRuleset = "SYSTEM"
	Global.Common.ProxySubscription = "NONE"

	Global.NodePref.ClashProxiesStyle = "flow"
	Global.NodePref.ClashProxyGroupsStyle = "block"
	Global.NodePref.AppendSubUserinfo = true
	Global.NodePref.SingBoxAddClashModes = true

	Global.ManagedConfig.WriteManagedConfig = true
	Global.ManagedConfig.ManagedConfigPrefix = "http://127.0.0.1:25500"
	Global.ManagedConfig.ConfigUpdateInterval = 86400

	Global.SurgeExternal.ResolveHostname = true

	Global.Emojis.AddEmoji = true
	Global.Emojis.RemoveOldEmoji = true

	Global.Rulesets.Enabled = true

	Global.Server.Listen = "0.0.0.0"
	Global.Server.Port = 25500

	Global.Advanced.LogLevel = "info"
	Global.Advanced.MaxPendingConnections = 10240
	Global.Advanced.MaxConcurrentThreads = 2
	Global.Advanced.MaxAllowedRulesets = 64
	Global.Advanced.CacheSubscription = 60
	Global.Advanced.CacheConfig = 300
	Global.Advanced.CacheRuleset = 21600
	Global.Advanced.ScriptCleanContext = true
}

// LoadConfig loads configuration from pref files
func LoadConfig() (string, error) {
	// Check for config configFile existence in order: yml -> toml -> ini
	configFileList := []string{
		"pref.yml",
		"pref.toml",
		"pref.ini",
		"pref.example.yml",
		"pref.example.toml",
		"pref.example.ini",
	}
	var parseError error = fmt.Errorf("no configuration file found")
	var effectiveConfig = ""
	for _, configFile := range configFileList {
		if _, err := os.Stat(configFile); err == nil {
			effectiveConfig = configFile
			if strings.Contains(configFile, "example.") {
				effectiveConfig = strings.Replace(configFile, "example.", "", 1)
				err := copyFile(configFile, effectiveConfig)
				if err != nil {
					return "", err
				}
			}
			if strings.HasSuffix(effectiveConfig, ".toml") {
				parseError = loadTOMLConfig(effectiveConfig)
			} else if strings.HasSuffix(effectiveConfig, ".yaml") || strings.HasSuffix(effectiveConfig, ".yml") {
				parseError = loadYAMLConfig(effectiveConfig)
			} else if strings.HasSuffix(effectiveConfig, ".ini") {
				parseError = loadINIConfig(effectiveConfig)
			}
			if parseError == nil {
				break
			}
		}
	}

	return effectiveConfig, parseError
}

func loadYAMLConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read YAML config: %w", err)
	}

	if err := yaml.Unmarshal(data, Global); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Merge top-level custom_proxy_group into ProxyGroups.CustomProxyGroups
	if len(Global.CustomGroups) > 0 {
		Global.ProxyGroups.CustomProxyGroups = append(Global.ProxyGroups.CustomProxyGroups, Global.CustomGroups...)
		Global.CustomGroups = nil // Clear to avoid confusion
	}

	// Merge top-level rulesets into Rulesets.Rulesets
	if len(Global.CustomRulesets) > 0 {
		Global.Rulesets.Rulesets = append(Global.Rulesets.Rulesets, Global.CustomRulesets...)
		Global.CustomRulesets = nil // Clear to avoid confusion
	}

	if err := processImports(); err != nil {
		return fmt.Errorf("failed to process imports: %w", err)
	}

	return nil
}

func loadTOMLConfig(path string) error {
	if _, err := toml.DecodeFile(path, Global); err != nil {
		return fmt.Errorf("failed to parse TOML config: %w", err)
	}

	// Merge top-level custom_groups into ProxyGroups.CustomProxyGroups
	if len(Global.CustomGroups) > 0 {
		Global.ProxyGroups.CustomProxyGroups = append(Global.ProxyGroups.CustomProxyGroups, Global.CustomGroups...)
		Global.CustomGroups = nil // Clear to avoid confusion
	}

	// Merge top-level rulesets into Rulesets.Rulesets
	if len(Global.CustomRulesets) > 0 {
		Global.Rulesets.Rulesets = append(Global.Rulesets.Rulesets, Global.CustomRulesets...)
		Global.CustomRulesets = nil // Clear to avoid confusion
	}

	if err := processImports(); err != nil {
		return fmt.Errorf("failed to process imports: %w", err)
	}

	return nil
}

func loadINIConfig(path string) error {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		AllowShadows:               true,
		AllowBooleanKeys:           true,
		UnparseableSections:        []string{},
		InsensitiveSections:        false,
		InsensitiveKeys:            false,
		IgnoreInlineComment:        true,
		AllowPythonMultilineValues: true,
	}, path)
	if err != nil {
		return fmt.Errorf("failed to load INI config: %w", err)
	}

	// Map INI to struct using reflection for common section
	if err := cfg.Section("common").MapTo(&Global.Common); err != nil {
		return fmt.Errorf("failed to map common section: %w", err)
	}

	// Parse userinfo section
	if sec := cfg.Section("userinfo"); sec != nil {
		Global.UserInfo.StreamRules = parseINIRules(sec, "stream_rule")
		Global.UserInfo.TimeRules = parseINIRules(sec, "time_rule")
	}

	// Parse node_pref section manually
	if sec := cfg.Section("node_pref"); sec != nil {
		if key := sec.Key("udp_flag"); key != nil && key.String() != "" {
			val := key.MustBool()
			Global.NodePref.UDPFlag = &val
		}
		if key := sec.Key("tcp_fast_open_flag"); key != nil && key.String() != "" {
			val := key.MustBool()
			Global.NodePref.TCPFastOpenFlag = &val
		}
		if key := sec.Key("skip_cert_verify_flag"); key != nil && key.String() != "" {
			val := key.MustBool()
			Global.NodePref.SkipCertVerifyFlag = &val
		}
		if key := sec.Key("tls13_flag"); key != nil && key.String() != "" {
			val := key.MustBool()
			Global.NodePref.TLS13Flag = &val
		}
		if key := sec.Key("sort_flag"); key != nil {
			Global.NodePref.SortFlag = key.MustBool()
		}
		if key := sec.Key("sort_script"); key != nil {
			Global.NodePref.SortScript = key.String()
		}
		if key := sec.Key("filter_deprecated_nodes"); key != nil {
			Global.NodePref.FilterDeprecatedNodes = key.MustBool()
		}
		if key := sec.Key("append_sub_userinfo"); key != nil {
			Global.NodePref.AppendSubUserinfo = key.MustBool()
		}
		if key := sec.Key("clash_proxies_style"); key != nil {
			Global.NodePref.ClashProxiesStyle = key.String()
		}
		if key := sec.Key("clash_proxy_groups_style"); key != nil {
			Global.NodePref.ClashProxyGroupsStyle = key.String()
		}
		if key := sec.Key("singbox_add_clash_modes"); key != nil {
			Global.NodePref.SingBoxAddClashModes = key.MustBool()
		}
		Global.NodePref.RenameNodes = parseINIRenameRules(sec, "rename_node")
	}

	// Map managed_config section
	if err := cfg.Section("managed_config").MapTo(&Global.ManagedConfig); err != nil {
		return fmt.Errorf("failed to map managed_config section: %w", err)
	}

	// Map surge_external_proxy section
	if err := cfg.Section("surge_external_proxy").MapTo(&Global.SurgeExternal); err != nil {
		return fmt.Errorf("failed to map surge_external_proxy section: %w", err)
	}

	// Parse emojis section
	if sec := cfg.Section("emojis"); sec != nil {
		if key := sec.Key("add_emoji"); key != nil {
			Global.Emojis.AddEmoji = key.MustBool(true)
		}
		if key := sec.Key("remove_old_emoji"); key != nil {
			Global.Emojis.RemoveOldEmoji = key.MustBool(true)
		}
		Global.Emojis.Rules = parseINIEmojiRules(sec, "rule")
	}

	// Parse rulesets section
	if sec := cfg.Section("rulesets"); sec != nil {
		if key := sec.Key("enabled"); key != nil {
			Global.Rulesets.Enabled = key.MustBool(true)
		}
		if key := sec.Key("overwrite_original_rules"); key != nil {
			Global.Rulesets.OverwriteOriginalRules = key.MustBool(false)
		}
		if key := sec.Key("update_ruleset_on_request"); key != nil {
			Global.Rulesets.UpdateRulesetOnRequest = key.MustBool(false)
		}
		Global.Rulesets.Rulesets = parseINIRulesets(sec, "ruleset")
	}

	// Parse proxy_groups section
	if sec := cfg.Section("proxy_groups"); sec != nil {
		Global.ProxyGroups.CustomProxyGroups = parseINIProxyGroups(sec, "custom_proxy_group")
	}

	// Parse template section
	if sec := cfg.Section("template"); sec != nil {
		if key := sec.Key("template_path"); key != nil {
			Global.Template.TemplatePath = key.String()
		}
		Global.Template.Globals = parseINITemplateGlobals(sec)
	}

	// Parse aliases section
	if sec := cfg.Section("aliases"); sec != nil {
		Global.Aliases = parseINIAliases(sec)
	}

	// Parse tasks section
	if sec := cfg.Section("tasks"); sec != nil {
		Global.Tasks = parseINITasks(sec, "task")
	}

	// Map server section
	if err := cfg.Section("server").MapTo(&Global.Server); err != nil {
		return fmt.Errorf("failed to map server section: %w", err)
	}

	// Map advanced section
	if err := cfg.Section("advanced").MapTo(&Global.Advanced); err != nil {
		return fmt.Errorf("failed to map advanced section: %w", err)
	}

	if err := processImports(); err != nil {
		return fmt.Errorf("failed to process imports: %w", err)
	}

	return nil
}

// Helper functions for parsing INI format

func parseINIRules(sec *ini.Section, keyName string) []RegexRuleConfig {
	var rules []RegexRuleConfig
	key := sec.Key(keyName)
	if key == nil {
		return rules
	}

	for _, val := range key.ValueWithShadows() {
		parts := strings.SplitN(val, "|", 2)
		if len(parts) == 2 {
			rules = append(rules, RegexRuleConfig{
				Match:   parts[0],
				Replace: parts[1],
			})
		}
	}
	return rules
}

func parseINIRenameRules(sec *ini.Section, keyName string) []RenameNodeConfig {
	var rules []RenameNodeConfig
	key := sec.Key(keyName)
	if key == nil {
		return rules
	}

	for _, val := range key.ValueWithShadows() {
		// Check for special prefixes
		if strings.HasPrefix(val, "!!import:") {
			rules = append(rules, RenameNodeConfig{
				Import: strings.TrimPrefix(val, "!!import:"),
			})
		} else if strings.HasPrefix(val, "!!script:") {
			rules = append(rules, RenameNodeConfig{
				Script: strings.TrimPrefix(val, "!!script:"),
			})
		} else {
			// Normal match@replace format
			parts := strings.SplitN(val, "@", 2)
			if len(parts) == 2 {
				rules = append(rules, RenameNodeConfig{
					Match:   parts[0],
					Replace: parts[1],
				})
			}
		}
	}
	return rules
}

func parseINIEmojiRules(sec *ini.Section, keyName string) []EmojiRuleConfig {
	var rules []EmojiRuleConfig
	key := sec.Key(keyName)
	if key == nil {
		return rules
	}

	for _, val := range key.ValueWithShadows() {
		if strings.HasPrefix(val, "!!import:") {
			rules = append(rules, EmojiRuleConfig{
				Import: strings.TrimPrefix(val, "!!import:"),
			})
		} else if strings.HasPrefix(val, "!!script:") {
			rules = append(rules, EmojiRuleConfig{
				Script: strings.TrimPrefix(val, "!!script:"),
			})
		} else {
			// Normal match,emoji format
			parts := strings.SplitN(val, ",", 2)
			if len(parts) == 2 {
				rules = append(rules, EmojiRuleConfig{
					Match: parts[0],
					Emoji: parts[1],
				})
			}
		}
	}
	return rules
}

func parseINIRulesets(sec *ini.Section, keyName string) []RulesetConfig {
	var rulesets []RulesetConfig
	key := sec.Key(keyName)
	if key == nil {
		return rulesets
	}

	for _, val := range key.ValueWithShadows() {
		if strings.HasPrefix(val, "!!import:") {
			rulesets = append(rulesets, RulesetConfig{
				Import: strings.TrimPrefix(val, "!!import:"),
			})
		} else {
			// Format: Group,URL[,interval] or Group,[]Rule
			parts := strings.SplitN(val, ",", 3)
			if len(parts) >= 2 {
				rs := RulesetConfig{
					Group: parts[0],
				}

				// Check if it's a rule or ruleset
				if strings.HasPrefix(parts[1], "[]") {
					// For inline rules like []GEOIP,CN - combine remaining parts
					if len(parts) > 2 {
						rs.Rule = parts[1] + "," + parts[2]
					} else {
						rs.Rule = parts[1]
					}
				} else {
					// Extract type prefix if exists
					rulesetPart := parts[1]
					if idx := strings.Index(rulesetPart, ":"); idx > 0 {
						rs.Type = rulesetPart[:idx]
						rs.Ruleset = rulesetPart[idx+1:]
					} else {
						rs.Ruleset = rulesetPart
					}

					// Parse interval if exists (only for rulesets, not rules)
					if len(parts) == 3 {
						fmt.Sscanf(parts[2], "%d", &rs.Interval)
					}
				}

				rulesets = append(rulesets, rs)
			}
		}
	}
	return rulesets
}

func parseINIProxyGroups(sec *ini.Section, keyName string) []ProxyGroupConfig {
	var groups []ProxyGroupConfig
	key := sec.Key(keyName)
	if key == nil {
		return groups
	}

	for _, val := range key.ValueWithShadows() {
		if strings.HasPrefix(val, "!!import:") {
			groups = append(groups, ProxyGroupConfig{
				Import: strings.TrimPrefix(val, "!!import:"),
			})
		} else if strings.HasPrefix(val, "!!script:") {
			groups = append(groups, ProxyGroupConfig{
				Script: strings.TrimPrefix(val, "!!script:"),
			})
		} else {
			// Format: Name`type`rules...
			parts := strings.Split(val, "`")
			if len(parts) >= 2 {
				g := ProxyGroupConfig{
					Name: parts[0],
					Type: parts[1],
					Rule: parts[2:],
				}

				// Parse URL test parameters if present
				if g.Type == "url-test" || g.Type == "fallback" || g.Type == "load-balance" {
					if len(parts) >= 4 {
						g.URL = parts[len(parts)-2]
						// Parse interval,timeout,tolerance from last part
						lastPart := parts[len(parts)-1]
						params := strings.Split(lastPart, ",")
						if len(params) > 0 {
							fmt.Sscanf(params[0], "%d", &g.Interval)
						}
						if len(params) > 1 {
							fmt.Sscanf(params[1], "%d", &g.Timeout)
						}
						if len(params) > 2 {
							fmt.Sscanf(params[2], "%d", &g.Tolerance)
						}
						// Remove URL and params from rules
						g.Rule = parts[2 : len(parts)-2]
					}
				}

				groups = append(groups, g)
			}
		}
	}
	return groups
}

func parseINITemplateGlobals(sec *ini.Section) []TemplateGlobalConfig {
	var globals []TemplateGlobalConfig
	for _, key := range sec.Keys() {
		if key.Name() != "template_path" {
			globals = append(globals, TemplateGlobalConfig{
				Key:   key.Name(),
				Value: key.String(),
			})
		}
	}
	return globals
}

func parseINIAliases(sec *ini.Section) []AliasConfig {
	var aliases []AliasConfig
	for _, key := range sec.Keys() {
		aliases = append(aliases, AliasConfig{
			URI:    key.Name(),
			Target: key.String(),
		})
	}
	return aliases
}

func parseINITasks(sec *ini.Section, keyName string) []TaskConfig {
	var tasks []TaskConfig
	keys := sec.KeysHash()
	for key, val := range keys {
		if key == keyName {
			// Format: Name`Cron`Path`Timeout
			parts := strings.Split(val, "`")
			if len(parts) >= 3 {
				t := TaskConfig{
					Name:    parts[0],
					CronExp: parts[1],
					Path:    parts[2],
				}
				if len(parts) >= 4 {
					fmt.Sscanf(parts[3], "%d", &t.Timeout)
				}
				tasks = append(tasks, t)
			}
		}
	}
	return tasks
}

// processImports processes import directives in config sections
func processImports() error {
	// Process proxy group imports - iterate backwards to safely remove items
	var indicesToRemove []int
	for i := range Global.ProxyGroups.CustomProxyGroups {
		group := &Global.ProxyGroups.CustomProxyGroups[i]
		if group.Import != "" {
			importPath := resolveImportPath(group.Import)
			if err := loadProxyGroupImport(importPath); err != nil {
				return fmt.Errorf("failed to load proxy group import %s: %w", group.Import, err)
			}
			// Mark for removal after processing
			indicesToRemove = append(indicesToRemove, i)
		}
	}

	// Remove groups that were only import directives
	for i := len(indicesToRemove) - 1; i >= 0; i-- {
		idx := indicesToRemove[i]
		Global.ProxyGroups.CustomProxyGroups = append(
			Global.ProxyGroups.CustomProxyGroups[:idx],
			Global.ProxyGroups.CustomProxyGroups[idx+1:]...,
		)
	}

	// Process ruleset imports
	indicesToRemove = nil
	for i := range Global.Rulesets.Rulesets {
		ruleset := &Global.Rulesets.Rulesets[i]
		if ruleset.Import != "" {
			importPath := resolveImportPath(ruleset.Import)
			if err := loadRulesetImport(importPath); err != nil {
				return fmt.Errorf("failed to load ruleset import %s: %w", ruleset.Import, err)
			}
			indicesToRemove = append(indicesToRemove, i)
		}
	}
	for i := len(indicesToRemove) - 1; i >= 0; i-- {
		idx := indicesToRemove[i]
		Global.Rulesets.Rulesets = append(
			Global.Rulesets.Rulesets[:idx],
			Global.Rulesets.Rulesets[idx+1:]...,
		)
	}

	// Process rename node imports
	indicesToRemove = nil
	for i := range Global.NodePref.RenameNodes {
		rename := &Global.NodePref.RenameNodes[i]
		if rename.Import != "" {
			importPath := resolveImportPath(rename.Import)
			if err := loadRenameImport(importPath); err != nil {
				return fmt.Errorf("failed to load rename import %s: %w", rename.Import, err)
			}
			indicesToRemove = append(indicesToRemove, i)
		}
	}
	for i := len(indicesToRemove) - 1; i >= 0; i-- {
		idx := indicesToRemove[i]
		Global.NodePref.RenameNodes = append(
			Global.NodePref.RenameNodes[:idx],
			Global.NodePref.RenameNodes[idx+1:]...,
		)
	}

	// Process emoji imports
	indicesToRemove = nil
	for i := range Global.Emojis.Rules {
		emoji := &Global.Emojis.Rules[i]
		if emoji.Import != "" {
			importPath := resolveImportPath(emoji.Import)
			if err := loadEmojiImport(importPath); err != nil {
				return fmt.Errorf("failed to load emoji import %s: %w", emoji.Import, err)
			}
			indicesToRemove = append(indicesToRemove, i)
		}
	}
	for i := len(indicesToRemove) - 1; i >= 0; i-- {
		idx := indicesToRemove[i]
		Global.Emojis.Rules = append(
			Global.Emojis.Rules[:idx],
			Global.Emojis.Rules[idx+1:]...,
		)
	}

	return nil
}

// resolveImportPath resolves relative import paths to absolute paths
func resolveImportPath(importPath string) string {
	if filepath.IsAbs(importPath) {
		return importPath
	}
	// First try relative to current working directory
	cwd, _ := os.Getwd()
	cwdPath := filepath.Join(cwd, importPath)
	if _, err := os.Stat(cwdPath); err == nil {
		return cwdPath
	}
	// Fall back to base path
	return filepath.Join(GetBasePath(), importPath)
}

// loadProxyGroupImport loads proxy groups from an import file
func loadProxyGroupImport(path string) error {
	ext := filepath.Ext(path)

	var imported struct {
		CustomGroups []ProxyGroupConfig `yaml:"custom_proxy_group" toml:"custom_groups" ini:"custom_proxy_group"`
	}

	switch ext {
	case ".toml":
		if _, err := toml.DecodeFile(path, &imported); err != nil {
			return fmt.Errorf("failed to parse TOML import: %w", err)
		}
	case ".yaml", ".yml":
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read YAML import: %w", err)
		}
		if err := yaml.Unmarshal(data, &imported); err != nil {
			return fmt.Errorf("failed to parse YAML import: %w", err)
		}
	case ".ini":
		// For INI files, we need to parse differently
		return loadProxyGroupImportINI(path)
	case ".txt":
		// Plain text format with backtick separators
		return loadProxyGroupImportTXT(path)
	default:
		return fmt.Errorf("unsupported import file format: %s", ext)
	}

	// Append imported groups to existing groups
	Global.ProxyGroups.CustomProxyGroups = append(Global.ProxyGroups.CustomProxyGroups, imported.CustomGroups...)
	return nil
}

// loadProxyGroupImportINI loads proxy groups from an INI file
func loadProxyGroupImportINI(path string) error {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		AllowShadows:               true,
		AllowBooleanKeys:           true,
		IgnoreInlineComment:        true,
		AllowPythonMultilineValues: true,
	}, path)
	if err != nil {
		return fmt.Errorf("failed to load INI import: %w", err)
	}

	// Parse custom_proxy_group keys
	section := cfg.Section("proxy_groups")
	if section == nil {
		section = cfg.Section("")
	}

	for _, key := range section.Keys() {
		if key.Name() == "custom_proxy_group" {
			// INI format: custom_proxy_group=name`type`rule1`rule2...
			parts := strings.Split(key.String(), "`")
			if len(parts) >= 2 {
				group := ProxyGroupConfig{
					Name: parts[0],
					Type: parts[1],
					Rule: parts[2:],
				}
				Global.ProxyGroups.CustomProxyGroups = append(Global.ProxyGroups.CustomProxyGroups, group)
			}
		}
	}

	return nil
}

// loadProxyGroupImportTXT loads proxy groups from a plain text file
func loadProxyGroupImportTXT(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read TXT import: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments (lines starting with ; or #)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Format: name`type`rule1`rule2...
		// Some fields may have additional parameters like: url-test`.*`http://...`300
		parts := strings.Split(line, "`")
		if len(parts) < 2 {
			continue
		}

		group := ProxyGroupConfig{
			Name: parts[0],
			Type: parts[1],
		}

		// Parse remaining parts as rules or special parameters
		for i := 2; i < len(parts); i++ {
			part := parts[i]
			// Check if it's a URL parameter (starts with http:// or https://)
			if strings.HasPrefix(part, "http://") || strings.HasPrefix(part, "https://") {
				group.URL = part
			} else if i == len(parts)-1 && group.URL != "" {
				// If we have a URL and this is the last part, it might be interval
				interval := 0
				if _, err := fmt.Sscanf(part, "%d", &interval); err == nil && interval > 0 {
					group.Interval = interval
				} else {
					// Otherwise it's a rule
					group.Rule = append(group.Rule, part)
				}
			} else {
				// It's a rule (proxy selector pattern)
				group.Rule = append(group.Rule, part)
			}
		}

		Global.ProxyGroups.CustomProxyGroups = append(Global.ProxyGroups.CustomProxyGroups, group)
	}

	return nil
}

// loadRulesetImport loads rulesets from an import file
func loadRulesetImport(path string) error {
	ext := filepath.Ext(path)

	var imported struct {
		Rulesets []RulesetConfig `yaml:"rulesets" toml:"rulesets" ini:"ruleset"`
	}

	switch ext {
	case ".toml":
		if _, err := toml.DecodeFile(path, &imported); err != nil {
			return fmt.Errorf("failed to parse TOML import: %w", err)
		}
	case ".yaml", ".yml":
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read YAML import: %w", err)
		}
		if err := yaml.Unmarshal(data, &imported); err != nil {
			return fmt.Errorf("failed to parse YAML import: %w", err)
		}
	case ".ini":
		return loadRulesetImportINI(path)
	case ".txt":
		return loadRulesetImportTXT(path)
	default:
		return fmt.Errorf("unsupported import file format: %s", ext)
	}

	// Post-process: convert ruleset="[]..." to rule="[]..."
	for i := range imported.Rulesets {
		rs := &imported.Rulesets[i]
		if strings.HasPrefix(rs.Ruleset, "[]") {
			rs.Rule = rs.Ruleset
			rs.Ruleset = ""
		}
	}

	// Append imported rulesets to existing rulesets
	Global.Rulesets.Rulesets = append(Global.Rulesets.Rulesets, imported.Rulesets...)
	return nil
}

// loadRulesetImportINI loads rulesets from an INI file
func loadRulesetImportINI(path string) error {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		AllowShadows:               true,
		AllowBooleanKeys:           true,
		IgnoreInlineComment:        true,
		AllowPythonMultilineValues: true,
	}, path)
	if err != nil {
		return fmt.Errorf("failed to load INI import: %w", err)
	}

	section := cfg.Section("rulesets")
	if section == nil {
		section = cfg.Section("")
	}

	for _, key := range section.Keys() {
		if key.Name() == "ruleset" {
			// INI format: ruleset=group,url[,type]
			parts := strings.Split(key.String(), ",")
			if len(parts) >= 2 {
				ruleset := RulesetConfig{
					Group:   parts[0],
					Ruleset: parts[1],
				}
				if len(parts) >= 3 {
					ruleset.Type = parts[2]
				}
				Global.Rulesets.Rulesets = append(Global.Rulesets.Rulesets, ruleset)
			}
		}
	}

	return nil
}

// loadRulesetImportTXT loads rulesets from a plain text file
func loadRulesetImportTXT(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read TXT import: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Format: group,url[,type] or group,[]RULE,params
		parts := strings.SplitN(line, ",", 3)
		if len(parts) < 2 {
			continue
		}

		ruleset := RulesetConfig{
			Group: parts[0],
		}

		// Check if it's a rule (starts with []) or a ruleset URL
		if strings.HasPrefix(parts[1], "[]") {
			// It's a rule like []GEOIP,CN or []FINAL
			if len(parts) >= 3 {
				ruleset.Rule = parts[1] + "," + parts[2]
			} else {
				ruleset.Rule = parts[1]
			}
		} else {
			// It's a ruleset URL
			ruleset.Ruleset = parts[1]
			if len(parts) >= 3 {
				ruleset.Type = parts[2]
			}
		}

		Global.Rulesets.Rulesets = append(Global.Rulesets.Rulesets, ruleset)
	}

	return nil
}

// loadRenameImport loads rename rules from an import file
func loadRenameImport(path string) error {
	ext := filepath.Ext(path)

	var imported struct {
		RenameNodes []RenameNodeConfig `yaml:"rename_node" toml:"rename_node" ini:"rename_node"`
	}

	switch ext {
	case ".toml":
		if _, err := toml.DecodeFile(path, &imported); err != nil {
			return fmt.Errorf("failed to parse TOML import: %w", err)
		}
	case ".yaml", ".yml":
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read YAML import: %w", err)
		}
		if err := yaml.Unmarshal(data, &imported); err != nil {
			return fmt.Errorf("failed to parse YAML import: %w", err)
		}
	case ".ini":
		return fmt.Errorf("INI import for rename_node not yet implemented")
	case ".txt":
		return loadRenameImportTXT(path)
	default:
		return fmt.Errorf("unsupported import file format: %s", ext)
	}

	Global.NodePref.RenameNodes = append(Global.NodePref.RenameNodes, imported.RenameNodes...)
	return nil
}

// loadRenameImportTXT loads rename rules from a plain text file
func loadRenameImportTXT(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read TXT import: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Format: match@replace
		parts := strings.SplitN(line, "@", 2)
		if len(parts) == 2 {
			rename := RenameNodeConfig{
				Match:   parts[0],
				Replace: parts[1],
			}
			Global.NodePref.RenameNodes = append(Global.NodePref.RenameNodes, rename)
		}
	}

	return nil
}

// loadEmojiImport loads emoji rules from an import file
func loadEmojiImport(path string) error {
	ext := filepath.Ext(path)

	var imported struct {
		Rules []EmojiRuleConfig `yaml:"rules" toml:"emoji" ini:"rule"`
	}

	switch ext {
	case ".toml":
		if _, err := toml.DecodeFile(path, &imported); err != nil {
			return fmt.Errorf("failed to parse TOML import: %w", err)
		}
	case ".yaml", ".yml":
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read YAML import: %w", err)
		}
		if err := yaml.Unmarshal(data, &imported); err != nil {
			return fmt.Errorf("failed to parse YAML import: %w", err)
		}
	case ".ini":
		return fmt.Errorf("INI import for emoji not yet implemented")
	case ".txt":
		return loadEmojiImportTXT(path)
	default:
		return fmt.Errorf("unsupported import file format: %s", ext)
	}

	Global.Emojis.Rules = append(Global.Emojis.Rules, imported.Rules...)
	return nil
}

// loadEmojiImportTXT loads emoji rules from a plain text file
func loadEmojiImportTXT(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read TXT import: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Format: match,emoji
		parts := strings.SplitN(line, ",", 2)
		if len(parts) == 2 {
			emoji := EmojiRuleConfig{
				Match: parts[0],
				Emoji: parts[1],
			}
			Global.Emojis.Rules = append(Global.Emojis.Rules, emoji)
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// GetBasePath returns the absolute path to the base directory
func GetBasePath() string {
	if filepath.IsAbs(Global.Common.BasePath) {
		return Global.Common.BasePath
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, Global.Common.BasePath)
}
