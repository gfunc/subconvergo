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
