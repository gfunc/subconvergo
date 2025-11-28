package core

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	"github.com/stretchr/testify/assert"
)

type MockSubconverterProxy struct {
	BaseProxy
}

func (m *MockSubconverterProxy) ToSingleConfig(ext *config.ProxySetting) (string, error) {
	return "", nil
}

func (m *MockSubconverterProxy) ToClashConfig(ext *config.ProxySetting) (map[string]interface{}, error) {
	return nil, nil
}

func TestSubconverterProxy_Implements_SingleConvertableMixin(t *testing.T) {
	var _ SingleConvertableMixin = (*MockSubconverterProxy)(nil)
	var _ ParsableProxy = (*MockSubconverterProxy)(nil)
}

func TestBaseProxy_GettersAndSetters(t *testing.T) {
	p := &BaseProxy{
		Type:    "ss",
		Remark:  "test",
		Server:  "1.2.3.4",
		Port:    8388,
		Group:   "group1",
		GroupId: 1,
	}

	assert.Equal(t, "ss", p.GetType())
	assert.Equal(t, "test", p.GetRemark())
	assert.Equal(t, "1.2.3.4", p.GetServer())
	assert.Equal(t, 8388, p.GetPort())
	assert.Equal(t, "group1", p.GetGroup())
	assert.Equal(t, 1, p.GetGroupId())

	p.SetRemark("new_remark")
	assert.Equal(t, "new_remark", p.GetRemark())

	p.SetGroup("new_group")
	assert.Equal(t, "new_group", p.GetGroup())

	p.SetGroupId(2)
	assert.Equal(t, 2, p.GetGroupId())
}
