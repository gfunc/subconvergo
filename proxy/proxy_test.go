package proxy

import (
	"net/url"
	"strings"
	"testing"
)

func TestShadowsocksProxyOptionsAndLink(t *testing.T) {
	p := &ShadowsocksProxy{BaseProxy: BaseProxy{Remark: "SS", Server: "host", Port: 8388}, Password: "pwd", EncryptMethod: "aes-256-gcm"}
	opts := p.ProxyOptions(nil)
	if opts["type"] != "ss" || opts["password"] != "pwd" {
		t.Errorf("unexpected ss options: %#v", opts)
	}
	if link, _ := p.GenerateLink(nil); link == "" {
		t.Errorf("GenerateLink should not be empty")
	}
}

func TestVMessProxyOptionsAndLink(t *testing.T) {
	p := &VMessProxy{BaseProxy: BaseProxy{Remark: "VM", Server: "v", Port: 443}, UUID: "123", Network: "ws", Path: "/p", Host: "h", TLS: true}
	opts := p.ProxyOptions(nil)
	if opts["type"] != "vmess" || opts["uuid"] != "123" {
		t.Errorf("unexpected vmess options: %#v", opts)
	}
	if link, _ := p.GenerateLink(nil); link == "" {
		t.Errorf("GenerateLink should not be empty")
	}
}

func TestTrojanProxyOptionsAndLink(t *testing.T) {
	p := &TrojanProxy{BaseProxy: BaseProxy{Remark: "TR", Server: "t", Port: 443}, Password: "pw", Network: "ws", Path: "/p", AllowInsecure: true}
	opts := p.ProxyOptions(nil)
	if opts["type"] != "trojan" || opts["password"] != "pw" || opts["network"] != "ws" {
		t.Errorf("unexpected trojan options: %#v", opts)
	}
	if link, _ := p.GenerateLink(nil); link == "" {
		t.Errorf("GenerateLink should not be empty")
	}
}

func TestVLESSProxyOptionsAndLink(t *testing.T) {
	p := &VLESSProxy{BaseProxy: BaseProxy{Remark: "VL", Server: "v", Port: 8443}, UUID: "abc", Network: "grpc", Path: "svc", TLS: true, AllowInsecure: true}
	opts := p.ProxyOptions(nil)
	if opts["type"] != "vless" || opts["uuid"] != "abc" {
		t.Errorf("unexpected vless options: %#v", opts)
	}
	if link, _ := p.GenerateLink(nil); link == "" {
		t.Errorf("GenerateLink should not be empty")
	}
}

func TestHysteriaProxyOptionsAndLink(t *testing.T) {
	params := url.Values{}
	params.Set("sni", "example.com")
	p := &HysteriaProxy{BaseProxy: BaseProxy{Type: "hysteria", Remark: "HY", Server: "h", Port: 8443}, Password: "auth", Params: params}
	opts := p.ProxyOptions(nil)
	if opts["type"] != "hysteria" || opts["sni"] != "example.com" {
		t.Errorf("unexpected hysteria options: %#v", opts)
	}
	if link, _ := p.GenerateLink(nil); link == "" {
		t.Errorf("GenerateLink should not be empty")
	}
}

func TestTUICProxyOptionsAndLink(t *testing.T) {
	params := url.Values{}
	params.Set("alpn", "h3")
	p := &TUICProxy{BaseProxy: BaseProxy{Remark: "TC", Server: "tu", Port: 443}, UUID: "id", Password: "pw", Params: params}
	opts := p.ProxyOptions(nil)
	if opts["type"] != "tuic" || opts["uuid"] != "id" {
		t.Errorf("unexpected tuic options: %#v", opts)
	}
	if link, _ := p.GenerateLink(nil); link == "" {
		t.Errorf("GenerateLink should not be empty")
	}
}

func TestBaseProxyMethods(t *testing.T) {
	proxy := &BaseProxy{Type: "ss", Remark: "Test", Server: "example.com", Port: 443, Group: "TestGroup"}
	if proxy.GetType() != "ss" || proxy.GetRemark() != "Test" || proxy.GetServer() != "example.com" || proxy.GetPort() != 443 || proxy.GetGroup() != "TestGroup" {
		t.Error("BaseProxy getters failed")
	}
	proxy.SetRemark("New")
	proxy.SetGroup("NewGroup")
	if proxy.GetRemark() != "New" || proxy.GetGroup() != "NewGroup" {
		t.Error("BaseProxy setters failed")
	}
}

func TestShadowsocksRProxy(t *testing.T) {
	proxy := &ShadowsocksRProxy{BaseProxy: BaseProxy{Type: "ssr", Remark: "SSR", Server: "ssr.com", Port: 443}, Password: "pass", EncryptMethod: "aes-256-cfb", Protocol: "origin", Obfs: "plain"}
	opts := proxy.ProxyOptions(nil)
	if opts["type"] != "ssr" {
		t.Error("SSR ProxyOptions failed")
	}
	link, _ := proxy.GenerateLink(nil)
	if !strings.HasPrefix(link, "ssr://") {
		t.Error("SSR GenerateLink failed")
	}
}

func TestUrlEncode(t *testing.T) {
	if urlEncode("Hello World") != "Hello%20World" {
		t.Error("urlEncode failed")
	}
}
