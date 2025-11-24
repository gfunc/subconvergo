package proxy

import (
	"testing"

	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/stretchr/testify/assert"
)

func TestAnyTLSParser_Parse(t *testing.T) {
	parser := &AnyTLSParser{}

	// Test case 1: Full link
	link := "anytls://password@1.2.3.4:443?peer=example.com&alpn=h2,http/1.1&hpkp=chrome&tfo=1&insecure=1&idle_session_check_interval=30&idle_session_timeout=60&min_idle_session=5#anytls-proxy"
	proxy, err := parser.ParseSingle(link)
	assert.NoError(t, err)
	var anytlsProxy *impl.AnyTLSProxy
	if mp, ok := proxy.(*impl.MihomoProxy); ok {
		anytlsProxy = mp.ProxyInterface.(*impl.AnyTLSProxy)
	} else {
		anytlsProxy = proxy.(*impl.AnyTLSProxy)
	}
	assert.Equal(t, "anytls", anytlsProxy.Type)
	assert.Equal(t, "anytls-proxy", anytlsProxy.Remark)
	assert.Equal(t, "1.2.3.4", anytlsProxy.Server)
	assert.Equal(t, 443, anytlsProxy.Port)
	assert.Equal(t, "password", anytlsProxy.Password)
	assert.Equal(t, "example.com", anytlsProxy.SNI)
	assert.Equal(t, []string{"h2", "http/1.1"}, anytlsProxy.Alpn)
	assert.Equal(t, "chrome", anytlsProxy.Fingerprint)
	assert.Equal(t, 30, anytlsProxy.IdleSessionCheckInterval)
	assert.Equal(t, 60, anytlsProxy.IdleSessionTimeout)
	assert.Equal(t, 5, anytlsProxy.MinIdleSession)
	assert.True(t, anytlsProxy.TFO)
	assert.True(t, anytlsProxy.AllowInsecure)

	// Test case 2: Minimal link
	link = "anytls://password@1.2.3.4:443"
	proxy, err = parser.ParseSingle(link)
	assert.NoError(t, err)

	var anytlsProxy2 *impl.AnyTLSProxy
	if mp, ok := proxy.(*impl.MihomoProxy); ok {
		anytlsProxy2 = mp.ProxyInterface.(*impl.AnyTLSProxy)
	} else {
		anytlsProxy2 = proxy.(*impl.AnyTLSProxy)
	}
	assert.NotNil(t, anytlsProxy2)
	assert.Equal(t, "1.2.3.4", anytlsProxy2.Server)
	assert.Equal(t, 443, anytlsProxy2.Port)
	assert.Equal(t, "password", anytlsProxy2.Password)
	assert.Equal(t, "1.2.3.4:443", anytlsProxy2.Remark)

	// Test case 3: Invalid link
	link = "invalid://link"
	assert.False(t, parser.CanParseLine(link))
	_, err = parser.ParseSingle(link)
	assert.Error(t, err)
}
