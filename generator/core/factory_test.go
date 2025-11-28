package core

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	pc "github.com/gfunc/subconvergo/proxy/core"
	"github.com/stretchr/testify/assert"
)

type mockGenerator struct{}

func (m *mockGenerator) Name() string {
	return "mock"
}

func (m *mockGenerator) Generate(proxies []pc.ProxyInterface, groups []config.ProxyGroupConfig, rules []string, global *config.Settings, opts GeneratorOptions) (string, error) {
	return "mock output", nil
}

func TestRegisterAndGetGenerator(t *testing.T) {
	mock := &mockGenerator{}
	RegisterGenerator(mock)

	g, err := GetGenerator("mock")
	assert.NoError(t, err)
	assert.Equal(t, mock, g)

	_, err = GetGenerator("nonexistent")
	assert.Error(t, err)
}
