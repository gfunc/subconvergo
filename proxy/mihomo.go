package proxy

import (
	"encoding/json"
	"fmt"

	"log"

	"github.com/metacubex/mihomo/constant"

	"gopkg.in/yaml.v3"
)

type MihomoProxy struct {
	ProxyInterface
	Clash   constant.Proxy         `yaml:"-" json:"-"`
	Options map[string]interface{} `yaml:"-" json:"-"`
}

func (m *MihomoProxy) GenerateLink() (string, error) {
	if m.ProxyInterface == nil {
		return "", fmt.Errorf("Plain proxy is not set")
	}
	switch p := m.ProxyInterface.(type) {
	case SubconverterProxy:
		return p.GenerateLink()
	default:
		return "", fmt.Errorf("GenerateLink not supported for this proxy type")
	}
}

func (m *MihomoProxy) ProxyOptions() map[string]interface{} {
	options, err := m.proxyOptions()
	if err != nil {
		log.Printf("failed to get proxy options: %v", err)
		return nil
	}
	return options
}

func (m *MihomoProxy) proxyOptions() (map[string]interface{}, error) {
	if m.Options != nil {
		return m.Options, nil
	}
	// Fallback to marshalling the Clash proxy
	b, err := m.Clash.MarshalJSON()
	if err != nil {
		return nil, err
	}
	options := make(map[string]interface{})
	err = json.Unmarshal(b, &options)
	if err != nil {
		return nil, err
	}
	return options, nil
}

func (m *MihomoProxy) MarshalYAML() (interface{}, error) {
	options, err := m.proxyOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to get proxy options for YAML marshal: %w", err)
	}
	return yaml.Marshal(options)
}

func (m *MihomoProxy) GetVmessProxy() (*VMessProxy, error) {
	p, e := m.ProxyInterface.(*VMessProxy)
	if !e {
		return nil, fmt.Errorf("not a vmess proxy")
	}
	return p, nil
}

func (m *MihomoProxy) GetShadowsocksProxy() (*ShadowsocksProxy, error) {
	p, e := m.ProxyInterface.(*ShadowsocksProxy)
	if !e {
		return nil, fmt.Errorf("not a shadowsocks proxy")
	}
	return p, nil
}

func (m *MihomoProxy) GetShadowsocksRProxy() (*ShadowsocksRProxy, error) {
	p, e := m.ProxyInterface.(*ShadowsocksRProxy)
	if !e {
		return nil, fmt.Errorf("not a shadowsocksr proxy")
	}
	return p, nil
}

func (m *MihomoProxy) GetTrojanProxy() (*TrojanProxy, error) {
	p, e := m.ProxyInterface.(*TrojanProxy)
	if !e {
		return nil, fmt.Errorf("not a trojan proxy")
	}
	return p, nil
}

func (m *MihomoProxy) GetVLESSProxy() (*VLESSProxy, error) {
	p, e := m.ProxyInterface.(*VLESSProxy)
	if !e {
		return nil, fmt.Errorf("not a vless proxy")
	}
	return p, nil
}

func (m *MihomoProxy) GetHysteriaProxy() (*HysteriaProxy, error) {
	p, e := m.ProxyInterface.(*HysteriaProxy)
	if !e {
		return nil, fmt.Errorf("not a hysteria proxy")
	}
	return p, nil
}

func (m *MihomoProxy) GetTUICProxy() (*TUICProxy, error) {
	p, e := m.ProxyInterface.(*TUICProxy)
	if !e {
		return nil, fmt.Errorf("not a tuic proxy")
	}
	return p, nil
}
