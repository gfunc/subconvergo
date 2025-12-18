package impl

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/utils"
)

// VLESSProxy represents a VLESS proxy
type VLESSProxy struct {
	core.BaseProxy  `yaml:",inline"`
	UUID            string   `yaml:"uuid" json:"uuid"`
	Network         string   `yaml:"network" json:"network"`
	Path            string   `yaml:"path" json:"path"`
	Host            string   `yaml:"host" json:"host"`
	TLS             bool     `yaml:"tls" json:"tls"`
	AllowInsecure   bool     `yaml:"allow_insecure" json:"allow_insecure"`
	Flow            string   `yaml:"flow" json:"flow"`
	SNI             string   `yaml:"sni" json:"sni"`
	Alpn            []string `yaml:"alpn" json:"alpn"`
	Fingerprint     string   `yaml:"fingerprint" json:"fingerprint"`
	PublicKey       string   `yaml:"public-key" json:"public-key"`
	ShortID         string   `yaml:"short-id" json:"short-id"`
	GRPCMode        string   `yaml:"grpc-mode" json:"grpc-mode"`
	GRPCServiceName string   `yaml:"grpc-service-name" json:"grpc-service-name"`
	Edge            string   `yaml:"edge" json:"edge"`
	XTLS            int      `yaml:"xtls" json:"xtls"`
}

func (p *VLESSProxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
	// Format: vless://uuid@server:port?params#remark
	link := fmt.Sprintf("vless://%s@%s:%d", p.UUID, p.Server, p.Port)

	params := []string{fmt.Sprintf("type=%s", p.Network)}

	if p.PublicKey != "" {
		params = append(params, "security=reality")
		params = append(params, fmt.Sprintf("pbk=%s", p.PublicKey))
		if p.ShortID != "" {
			params = append(params, fmt.Sprintf("sid=%s", p.ShortID))
		}
		if p.Fingerprint != "" {
			params = append(params, fmt.Sprintf("fp=%s", p.Fingerprint))
		}
		if p.SNI != "" {
			params = append(params, fmt.Sprintf("sni=%s", p.SNI))
		}
		if p.Flow != "" {
			params = append(params, fmt.Sprintf("flow=%s", p.Flow))
		}
	} else if p.TLS {
		params = append(params, "security=tls")
		if p.SNI != "" {
			params = append(params, fmt.Sprintf("sni=%s", p.SNI))
		}
		if len(p.Alpn) > 0 {
			params = append(params, fmt.Sprintf("alpn=%s", strings.Join(p.Alpn, ",")))
		}
		if p.Fingerprint != "" {
			params = append(params, fmt.Sprintf("fp=%s", p.Fingerprint))
		}
		if p.Flow != "" {
			params = append(params, fmt.Sprintf("flow=%s", p.Flow))
		}
	}

	if p.Network == "ws" {
		if p.Path != "" {
			params = append(params, fmt.Sprintf("path=%s", url.QueryEscape(p.Path)))
		}
		if p.Host != "" {
			params = append(params, fmt.Sprintf("host=%s", p.Host))
		}
	} else if p.Network == "grpc" {
		if p.GRPCServiceName != "" {
			params = append(params, fmt.Sprintf("serviceName=%s", p.GRPCServiceName))
		}
		if p.GRPCMode != "" {
			params = append(params, fmt.Sprintf("mode=%s", p.GRPCMode))
		}
	} else if p.Network == "http" || p.Network == "h2" {
		if p.Path != "" {
			params = append(params, fmt.Sprintf("path=%s", url.QueryEscape(p.Path)))
		}
		if p.Host != "" {
			params = append(params, fmt.Sprintf("host=%s", p.Host))
		}
	} else if p.Network == "tcp" {
		if p.Host != "" {
			params = append(params, fmt.Sprintf("headerType=http&host=%s", p.Host))
		}
	}

	link += "?" + strings.Join(params, "&")

	if p.Remark != "" {
		link += "#" + utils.UrlEncode(p.Remark)
	}

	return link, nil
}

func (p *VLESSProxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	options := map[string]interface{}{
		"type":    "vless",
		"name":    p.Remark,
		"server":  p.Server,
		"port":    p.Port,
		"uuid":    p.UUID,
		"network": p.Network,
	}

	if p.TLS {
		options["tls"] = true
		if p.SNI != "" {
			options["servername"] = p.SNI
		}
		if len(p.Alpn) > 0 {
			options["alpn"] = p.Alpn
		}
	}

	if p.Fingerprint != "" {
		options["client-fingerprint"] = p.Fingerprint
	}

	if p.XTLS == 2 {
		options["flow"] = "xtls-rprx-vision"
	} else if p.Flow != "" {
		options["flow"] = p.Flow
	}

	if p.PublicKey != "" || p.ShortID != "" {
		realityOpts := map[string]interface{}{}
		if p.PublicKey != "" {
			realityOpts["public-key"] = p.PublicKey
		}
		if p.ShortID != "" {
			realityOpts["short-id"] = p.ShortID
		}
		options["reality-opts"] = realityOpts

		if _, ok := options["client-fingerprint"]; !ok {
			options["client-fingerprint"] = "chrome"
		}
	}

	var udp, tfo, scv, tls13 *bool
	udp = p.UDP
	tfo = p.TFO
	scv = p.SCV
	tls13 = p.TLS13

	if p.AllowInsecure {
		b := true
		scv = &b
	}

	if ext != nil {
		if ext.UDP != nil {
			udp = ext.UDP
		}
		if ext.TFO != nil {
			tfo = ext.TFO
		}
		if ext.SCV != nil {
			scv = ext.SCV
		}
		if ext.TLS13 != nil {
			tls13 = ext.TLS13
		}
	}

	if scv != nil && *scv {
		options["skip-cert-verify"] = true
	}
	if udp != nil && *udp {
		options["udp"] = true
		options["packet-encoding"] = "xudp"
	}
	if tfo != nil && *tfo {
		options["tfo"] = true
	}
	if tls13 != nil && *tls13 {
		options["tls13"] = true
	}

	switch p.Network {
	case "tcp":
		if p.Host != "" {
			options["host"] = p.Host
		}
		if p.Path != "" {
			options["path"] = p.Path
		}
	case "ws":
		wsOpts := make(map[string]interface{})
		wsOpts["path"] = "/"
		if p.Path != "" {
			wsOpts["path"] = p.Path
		}
		headers := make(map[string]string)
		if p.Host != "" {
			headers["Host"] = p.Host
		}
		if p.Edge != "" {
			headers["Edge"] = p.Edge
		}
		if len(headers) > 0 {
			wsOpts["headers"] = headers
		}
		options["ws-opts"] = wsOpts

	case "grpc":
		grpcOpts := make(map[string]interface{})
		if p.GRPCServiceName != "" {
			grpcOpts["grpc-service-name"] = p.GRPCServiceName
		} else if p.Path != "" {
			grpcOpts["grpc-service-name"] = p.Path
		}
		if p.GRPCMode != "" {
			grpcOpts["grpc-mode"] = p.GRPCMode
		}
		options["grpc-opts"] = grpcOpts

	case "http":
		httpOpts := make(map[string]interface{})
		httpOpts["method"] = "GET"
		path := "/"
		if p.Path != "" {
			path = p.Path
		}
		httpOpts["path"] = []string{path}
		headers := make(map[string][]string)
		if p.Host != "" {
			headers["Host"] = []string{p.Host}
		}
		if p.Edge != "" {
			headers["Edge"] = []string{p.Edge}
		}
		if len(headers) > 0 {
			httpOpts["headers"] = headers
		}
		options["http-opts"] = httpOpts

	case "h2":
		h2Opts := make(map[string]interface{})
		path := "/"
		if p.Path != "" {
			path = p.Path
		}
		h2Opts["path"] = path
		if p.Host != "" {
			h2Opts["host"] = []string{p.Host}
		}
		options["h2-opts"] = h2Opts
	}

	return options, nil
}

func (p *VLESSProxy) ToSurgeConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("VLESS not supported in Surge")
}

func (p *VLESSProxy) ToLoonConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("ToLoonConfig not supported for proxy type vless")
}

func (p *VLESSProxy) ToQuantumultXConfig(ext *config.ProxySetting) (string, error) {
	return "", fmt.Errorf("VLESS not supported in Quantumult X")
}

func (p *VLESSProxy) ToSingboxConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	outbound := map[string]interface{}{
		"type":        "vless",
		"tag":         p.Remark,
		"server":      p.Server,
		"server_port": p.Port,
		"uuid":        p.UUID,
	}

	if p.Flow != "" {
		outbound["flow"] = p.Flow
	}

	if p.TLS {
		tls := map[string]interface{}{
			"enabled": true,
		}
		if p.SNI != "" {
			tls["server_name"] = p.SNI
		}
		if p.AllowInsecure {
			tls["insecure"] = true
		}
		outbound["tls"] = tls
	}

	if p.Network == "ws" {
		transport := map[string]interface{}{
			"type": "ws",
		}
		if p.Path != "" {
			transport["path"] = p.Path
		}
		if p.Host != "" {
			transport["headers"] = map[string]string{
				"Host": p.Host,
			}
		}
		outbound["transport"] = transport
	} else if p.Network == "grpc" {
		transport := map[string]interface{}{
			"type": "grpc",
		}
		if p.Path != "" {
			transport["service_name"] = p.Path
		}
		outbound["transport"] = transport
	} else if p.Network == "http" || p.Network == "h2" {
		transport := map[string]interface{}{
			"type": "http",
		}
		if p.Path != "" {
			transport["path"] = p.Path
		}
		if p.Host != "" {
			transport["host"] = []string{p.Host}
		}
		outbound["transport"] = transport
	}

	return outbound, nil
}
