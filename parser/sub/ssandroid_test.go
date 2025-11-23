package sub

import (
	"testing"

	"github.com/gfunc/subconvergo/proxy/impl"
)

func TestSSAndroidSubscriptionParser_Parse(t *testing.T) {
	parser := &SSAndroidSubscriptionParser{}

	// Example content from subconverter tests or constructed
	// {"nodes":[{"server":"1.2.3.4","server_port":8388,"password":"pass","method":"aes-256-gcm","remarks":"test","plugin":"obfs-local","plugin_opts":"obfs=http;obfs-host=example.com"}]}
	// But SSAndroid format is usually just the list of nodes inside "proxy_apps" or similar?
	// Wait, CanParse checks for "proxy_apps".
	// But Parse wraps content in {"nodes": ...}.
	// So the input content should be `[{"server":...}, ...]`?
	// Let's check subconverter logic.
	// subconverter: `std::string content = "{\"nodes\":" + body + "}";`
	// So body is expected to be a JSON array `[...]`.
	
	content := `[
		{
			"server": "1.2.3.4",
			"server_port": 8388,
			"password": "pass",
			"method": "aes-256-gcm",
			"remarks": "test",
			"plugin": "obfs-local",
			"plugin_opts": "obfs=http;obfs-host=example.com"
		}
	]`

	sub, err := parser.Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(sub.Proxies) != 1 {
		t.Fatalf("Expected 1 proxy, got %d", len(sub.Proxies))
	}

	p := sub.Proxies[0].(*impl.ShadowsocksProxy)
	if p.Server != "1.2.3.4" {
		t.Errorf("Expected server 1.2.3.4, got %s", p.Server)
	}
	if p.Plugin != "obfs-local" {
		t.Errorf("Expected plugin obfs-local, got %s", p.Plugin)
	}
	
	if len(p.PluginOpts) != 2 {
		t.Errorf("Expected 2 plugin opts, got %d", len(p.PluginOpts))
	}
	if p.PluginOpts["obfs"] != "http" {
		t.Errorf("Expected obfs=http, got %v", p.PluginOpts["obfs"])
	}
	if p.PluginOpts["obfs-host"] != "example.com" {
		t.Errorf("Expected obfs-host=example.com, got %v", p.PluginOpts["obfs-host"])
	}
}
