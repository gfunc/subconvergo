package utils

import (
	"encoding/base64"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/metacubex/mihomo/adapter"
)

func ToMihomoProxy(pObj core.ParsableProxy) (core.ParsableProxy, error) {
	return ToMihomoProxyWithSetting(pObj, &config.ProxySetting{})
}

func ToMihomoProxyWithSetting(pObj core.ParsableProxy, config *config.ProxySetting) (core.ParsableProxy, error) {
	if _, ok := pObj.(*impl.MihomoProxy); ok {
		return pObj, nil
	}
	if oObj, ok := pObj.(core.ClashConvertableMixin); ok {

		option, err := oObj.ToClashConfig(config)
		if err != nil {
			log.Printf("[toMihomoProxy] Failed to convert proxy to Clash config: %v", err)
			return pObj, nil
		}

		mihomoProxy, err := adapter.ParseProxy(option)
		if err != nil {
			log.Printf("[toMihomoProxy] Converted proxy: %+v to Mihomo format: %+v, err: %v", pObj, mihomoProxy, err)
			return pObj, nil
		} else {
			return &impl.MihomoProxy{
				ProxyInterface: pObj,
				Clash:          mihomoProxy,
				Options:        option,
			}, nil
		}
	}
	return pObj, nil
}

func GetStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case float64:
			return strconv.FormatFloat(val, 'f', -1, 64)
		case int:
			return strconv.Itoa(val)
		}
	}
	return ""
}

// Helper functions (duplicated for now, should be in a shared utils package)
func UrlDecode(s string) string {
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return s
	}
	return decoded
}

func UrlSafeBase64Decode(s string) string {
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(s)
		if err != nil {
			return s
		}
	}
	return string(decoded)
}

func ParsePluginOpts(opts string) map[string]interface{} {
	result := make(map[string]interface{})
	pairs := strings.Split(opts, ";")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = UrlDecode(kv[1])
		} else {
			result[kv[0]] = "true"
		}
	}
	return result
}

func GetIntField(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		case string:
			i, _ := strconv.Atoi(val)
			return i
		}
	}
	return 0
}

func GetBoolField(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case bool:
			return val
		case string:
			return val == "true"
		}
	}
	return false
}

func ToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	default:
		return ""
	}
}

func ToInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}
