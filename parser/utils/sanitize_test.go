package utils

import (
	"strings"
	"testing"

	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for the parse-boundary sanitization and the source proxy-group
// filter guard.

func TestSanitizeScalarField_StripsLineBreaks(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"LF", "legit\n[rewrite_local]\n^https://attacker url reject", "legit[rewrite_local]^https://attacker url reject"},
		{"CR", "legit\r[rewrite_local]", "legit[rewrite_local]"},
		{"CRLF", "legit\r\n[rewrite_local]", "legit[rewrite_local]"},
		{"NUL", "pass\x00word", "password"},
		{"other C0 controls", "a\x01b\x07c\x1fd", "abcd"},
		{"DEL", "a\x7fb", "ab"},
		{"clean string unchanged", "HK 节点 01, \"vip\" = premium 🚀", "HK 节点 01, \"vip\" = premium 🚀"},
		{"unicode line separators kept", "a b", "a b"},
		{"empty", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, SanitizeScalarField(c.input))
		})
	}
}

// The funnel: every parser/proxy Parse* method returns via ToMihomoProxy*, so
// sanitizing there covers all protocols and subscription formats.
func TestToMihomoProxy_SanitizesAllStringFields(t *testing.T) {
	dirty := &impl.ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:            "ss",
			Remark:          "legit\n[rewrite_local]\n^https://attacker url reject",
			Server:          "1.2.3.4\n",
			Port:            8388,
			Group:           "grp\r\nX",
			UnderlyingProxy: "up\nX",
		},
		Password:      "pw\n[rewrite_local]",
		EncryptMethod: "aes-128-gcm",
		Plugin:        "obfs-local",
		PluginOpts: map[string]interface{}{
			"obfs":      "http",
			"obfs-host": "evil.com\n[rewrite_local]",
		},
	}

	out, err := ToMihomoProxy(dirty)
	require.NoError(t, err)

	ss, ok := out.(*impl.MihomoProxy).ProxyInterface.(*impl.ShadowsocksProxy)
	require.True(t, ok, "expected underlying *ShadowsocksProxy, got %T", out.(*impl.MihomoProxy).ProxyInterface)

	assert.Equal(t, "legit[rewrite_local]^https://attacker url reject", ss.Remark)
	assert.Equal(t, "1.2.3.4", ss.Server)
	assert.Equal(t, "grpX", ss.Group)
	assert.Equal(t, "upX", ss.UnderlyingProxy)
	assert.Equal(t, "pw[rewrite_local]", ss.Password)
	assert.Equal(t, "evil.com[rewrite_local]", ss.PluginOpts["obfs-host"])
	// Clean values untouched.
	assert.Equal(t, "aes-128-gcm", ss.EncryptMethod)
	assert.Equal(t, "http", ss.PluginOpts["obfs"])
}

func TestToMihomoProxy_SanitizesSlicesAndParams(t *testing.T) {
	dirty := &impl.VLESSProxy{
		BaseProxy: core.BaseProxy{
			Type:   "vless",
			Remark: "ok",
			Server: "1.2.3.4",
			Port:   443,
		},
		UUID: "id\nx",
		Alpn: []string{"h2", "http/1.1\n"},
	}

	out, err := ToMihomoProxy(dirty)
	require.NoError(t, err)
	v, ok := out.(*impl.MihomoProxy).ProxyInterface.(*impl.VLESSProxy)
	require.True(t, ok)
	assert.Equal(t, "idx", v.UUID)
	assert.Equal(t, []string{"h2", "http/1.1"}, v.Alpn)
}

func TestToMihomoProxyFromClash_SanitizesOptionsMap(t *testing.T) {
	dirty := &impl.ShadowsocksProxy{
		BaseProxy: core.BaseProxy{
			Type:   "ss",
			Remark: "name\n[rewrite_local]",
			Server: "1.2.3.4",
			Port:   8388,
		},
		Password:      "pw",
		EncryptMethod: "aes-128-gcm",
	}
	options := map[string]interface{}{
		"name":     "name\n[rewrite_local]",
		"server":   "1.2.3.4",
		"port":     8388,
		"password": "pw\r\nX",
		"ws-opts": map[string]interface{}{
			"path": "/p\nath",
		},
	}

	out, err := ToMihomoProxyFromClash(dirty, options)
	require.NoError(t, err)
	mp, ok := out.(*impl.MihomoProxy)
	require.True(t, ok)
	assert.Equal(t, "name[rewrite_local]", mp.Options["name"])
	assert.Equal(t, "pwX", mp.Options["password"])
	assert.Equal(t, "/path", mp.Options["ws-opts"].(map[string]interface{})["path"])
	assert.Equal(t, 8388, mp.Options["port"], "non-string values untouched")
}

// Filter guard (compile-level seam)

func TestIsSafeGroupFilter(t *testing.T) {
	safe := []string{
		"^(HK|SG)",
		"HK",
		"HK|香港",
		"港|台|HK|SG",
		"(?i)hk|sg",
		`^.* Premium .*$`,
		"a+",
		"(ab)+",        // single quantifier over a literal group is fine
		"(HK|SG|TW)",   // plain alternation, no quantifier
		"x{0,1}y{2}",   // sibling quantifiers are not nested
		"^HK(a|b){2}$", // quantified alternation with disjoint branches
	}
	for _, f := range safe {
		assert.True(t, IsSafeGroupFilter(f), "expected safe: %q", f)
	}

	unsafe := []string{
		"^(a+)+$",                // nested quantifier
		"(a|a)*",                 // quantifier over alternation with identical branches
		"(a*)+",                  // nested quantifier
		"(a+)*",                  // nested quantifier
		"((a|b)+)+",              // nested quantifier over alternation
		"(a?){2}",                // nested quantifier
		"[a-z",                   // does not compile
		"(",                      // does not compile
		strings.Repeat("a", 257), // over the length cap
		"",                       // empty
	}
	for _, f := range unsafe {
		assert.False(t, IsSafeGroupFilter(f), "expected unsafe: %q", f)
	}
}

func TestIsSafeGroupFilter_LengthBoundary(t *testing.T) {
	assert.True(t, IsSafeGroupFilter(strings.Repeat("a", MaxGroupFilterLength)))
	assert.False(t, IsSafeGroupFilter(strings.Repeat("a", MaxGroupFilterLength+1)))
}
