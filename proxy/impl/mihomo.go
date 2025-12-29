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

func (m *MihomoProxy) ToSingleConfig(opts *config.ProxySetting) (string, error) {
	if m.ProxyInterface == nil {
		return "", fmt.Errorf("Plain proxy is not set")
	}
	switch p := m.ProxyInterface.(type) {
	case core.ParsableProxy:
		return p.ToSingleConfig(opts)
	default:
		clashConfig, err := m.ToClashConfig(opts)
		if err != nil {
			return "", fmt.Errorf("error ToSingleConfig for proxy %s of type %s, %v", m.GetRemark(), m.GetType(), err)
		}
		b, err := json.Marshal(clashConfig)
		if err != nil {
			return "", fmt.Errorf("error ToSingleConfig for proxy %s of type %s, %v", m.GetRemark(), m.GetType(), err)
		}
		content := base64.StdEncoding.EncodeToString(b)
		return fmt.Sprintf("%s://%s", m.GetType(), content), nil
	}
}

func (m *MihomoProxy) ToClashConfig(opts *config.ProxySetting) (map[string]interface{}, error) {
	var generated map[string]interface{}
	var err error

	if p, ok := m.ProxyInterface.(core.ClashConvertableMixin); ok {
		generated, err = p.ToClashConfig(opts)
		if err != nil {
			return nil, err
		}
	}

	// If we have original options, merge them
	if m.Options != nil {
		// Start with a copy of Options to preserve unknown fields
		merged := make(map[string]interface{})
		for k, v := range m.Options {
			merged[k] = v
		}

		// Overwrite with generated fields (which are likely more up-to-date or processed)
		if generated != nil {
			for k, v := range generated {
				merged[k] = v
			}
		}

		// Ensure name matches remark
		merged["name"] = m.GetRemark()

		// Apply global overrides
		if _, ok := merged["udp"]; !ok && opts.UDP != nil {
			merged["udp"] = *opts.UDP
		}
		if _, ok := merged["tfo"]; !ok && opts.TFO != nil {
			merged["tfo"] = *opts.TFO
		}
		if _, ok := merged["skip-cert-verify"]; !ok && opts.SCV != nil {
			merged["skip-cert-verify"] = *opts.SCV
		}
		if _, ok := merged["tls13"]; !ok && opts.TLS13 != nil {
			merged["tls13"] = *opts.TLS13
		}

		return merged, nil
	}

	if generated != nil {
		return generated, nil
	}

	options, err := m.proxyOptions()
	if err != nil {
		log.Printf("failed to get proxy options: %v", err)
		return nil, err
	}
	return options, nil
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

func (m *MihomoProxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	if p, ok := m.ProxyInterface.(core.SurgeConvertableMixin); ok {
		return p.ToSurgeConfig(ext)
	}
	return "", fmt.Errorf("ToSurgeConfig not supported for proxy type %s", m.GetType())
}

func (m *MihomoProxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	if p, ok := m.ProxyInterface.(core.LoonConvertableMixin); ok {
		return p.ToLoonConfig(ext)
	}
	return "", fmt.Errorf("ToLoonConfig not supported for proxy type %s", m.GetType())
}

func (m *MihomoProxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	if p, ok := m.ProxyInterface.(core.QuantumultXConvertableMixin); ok {
		return p.ToQuantumultXConfig(ext)
	}
	return "", fmt.Errorf("ToQuantumultXConfig not supported for proxy type %s", m.GetType())
}

func (m *MihomoProxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	if p, ok := m.ProxyInterface.(core.SingboxConvertableMixin); ok {
		return p.ToSingboxConfig(ext)
	}
	return nil, fmt.Errorf("ToSingboxConfig not supported for proxy type %s", m.GetType())
}
