package impl

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/metacubex/mihomo/constant"

	"gopkg.in/yaml.v3"
)

type MihomoProxy struct {
	core.ProxyInterface
	Clash   constant.Proxy         `yaml:"-" json:"-"`
	Options map[string]interface{} `yaml:"-" json:"-"`
}

func (m *MihomoProxy) ToShareLink(opts *config.ProxySetting) (string, error) {
	if m.ProxyInterface == nil {
		return "", fmt.Errorf("Plain proxy is not set")
	}
	switch p := m.ProxyInterface.(type) {
	case core.SubconverterProxy:
		return p.ToShareLink(opts)
	default:
		b, err := json.Marshal(m.ToClashConfig(opts))
		if err != nil {
			return "", fmt.Errorf("error ToShareLink for proxy %s of type %s, %v", m.GetRemark(), m.GetType(), err)
		}
		content := base64.StdEncoding.EncodeToString(b)
		return fmt.Sprintf("%s://%s", m.GetType(), content), nil
	}
}

func (m *MihomoProxy) ToClashConfig(opts *config.ProxySetting) map[string]interface{} {
	options, err := m.proxyOptions()
	if err != nil {
		log.Printf("failed to get proxy options: %v", err)
		return nil
	}
	return options
}

func (m *MihomoProxy) proxyOptions() (map[string]interface{}, error) {
	if m.Options != nil {
		m.Options["name"] = m.GetRemark()
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
