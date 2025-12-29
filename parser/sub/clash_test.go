package sub

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
)

func TestClashSubscriptionParser_UDP(t *testing.T) {
	// Input Clash config with udp: false
	content := `
proxies:
  - name: "ss-udp-false"
    type: ss
    server: server
    port: 443
    cipher: aes-256-gcm
    password: password
    udp: false
  - name: "ss-udp-true"
    type: ss
    server: server
    port: 443
    cipher: aes-256-gcm
    password: password
    udp: true
  - name: "ss-udp-missing"
    type: ss
    server: server
    port: 443
    cipher: aes-256-gcm
    password: password
`

	// Parse
	subContent, err := ParseMihomoConfig(content)
	assert.NoError(t, err)
	assert.Len(t, subContent.Proxies, 3)

	// Check parsed values
	pFalse := subContent.Proxies[0]
	assert.Equal(t, "ss-udp-false", pFalse.GetRemark())

	pFalseMihomo, ok := pFalse.(*impl.MihomoProxy)
	assert.True(t, ok)
	pFalseSS, ok := pFalseMihomo.ProxyInterface.(*impl.ShadowsocksProxy)
	assert.True(t, ok)
	assert.Equal(t, false, *pFalseSS.UDP)

	pTrue := subContent.Proxies[1]
	assert.Equal(t, "ss-udp-true", pTrue.GetRemark())
	pTrueMihomo, ok := pTrue.(*impl.MihomoProxy)
	assert.True(t, ok)
	pTrueSS, ok := pTrueMihomo.ProxyInterface.(*impl.ShadowsocksProxy)
	assert.True(t, ok)
	assert.Equal(t, true, *pTrueSS.UDP)

	pMissing := subContent.Proxies[2]
	assert.Equal(t, "ss-udp-missing", pMissing.GetRemark())
	pMissingMihomo, ok := pMissing.(*impl.MihomoProxy)
	assert.True(t, ok)
	pMissingSS, ok := pMissingMihomo.ProxyInterface.(*impl.ShadowsocksProxy)
	assert.True(t, ok)
	assert.Nil(t, pMissingSS.UDP)

	// Generate Clash config
	opts := config.ProxySetting{} // No global overrides

	// Generate for pFalse
	mixinFalse, ok := pFalse.(core.ClashConvertableMixin)
	assert.True(t, ok)
	confFalse, err := mixinFalse.ToClashConfig(&opts)
	assert.NoError(t, err)

	// Check if udp is false in generated config
	val, ok := confFalse["udp"]
	if ok {
		assert.Equal(t, false, val, "udp should be false for ss-udp-false")
	} else {
		t.Log("udp field missing for ss-udp-false")
	}
	assert.NotEqual(t, true, val, "udp should not be true for ss-udp-false")

	// Generate for pTrue
	mixinTrue, ok := pTrue.(core.ClashConvertableMixin)
	assert.True(t, ok)
	confTrue, err := mixinTrue.ToClashConfig(&opts)
	assert.NoError(t, err)
	assert.Equal(t, true, confTrue["udp"], "udp should be true for ss-udp-true")

}

func TestClashSubscriptionParser_UDP_GlobalOverride(t *testing.T) {
	// Input Clash config with udp: false
	content := `
proxies:
  - name: "ss-udp-false"
    type: ss
    server: server
    port: 443
    cipher: aes-256-gcm
    password: password
    udp: false
  - name: "ss-udp-missing"
    type: ss
    server: server
    port: 443
    cipher: aes-256-gcm
    password: password
`

	// Parse
	subContent, err := ParseMihomoConfig(content)
	assert.NoError(t, err)
	assert.Len(t, subContent.Proxies, 2)

	pFalse := subContent.Proxies[0]
	pMissing := subContent.Proxies[1]

	// Global override: udp = true
	udpTrue := true
	opts := config.ProxySetting{
		UDP: &udpTrue,
	}

	// Case 1: Source has udp: false. Global has udp: true.
	// Expectation: Source wins (false).
	mixinFalse, ok := pFalse.(core.ClashConvertableMixin)
	assert.True(t, ok)
	confFalse, err := mixinFalse.ToClashConfig(&opts)
	assert.NoError(t, err)
	assert.Equal(t, false, confFalse["udp"], "udp should be false (source) even with global override")

	// Case 2: Source missing udp. Global has udp: true.
	// Expectation: Global wins (true).
	mixinMissing, ok := pMissing.(core.ClashConvertableMixin)
	assert.True(t, ok)
	confMissing, err := mixinMissing.ToClashConfig(&opts)
	assert.NoError(t, err)
	assert.Equal(t, true, confMissing["udp"], "udp should be true (global) when source is missing")
}
