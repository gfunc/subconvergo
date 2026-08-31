package sub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleSubscriptionParser_DeduplicateLines(t *testing.T) {
	content := `
ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ@127.0.0.1:8388#Node1
ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ@127.0.0.1:8388#Node1
ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ@127.0.0.1:8388#Node2
`
	parser := &SingleSubscriptionParser{}
	sub, err := parser.Parse(content)
	assert.NoError(t, err)
	assert.NotNil(t, sub)

	// Should only have 2 proxies, as the first two lines are identical
	assert.Equal(t, 2, len(sub.Proxies))
	assert.Equal(t, "Node1", sub.Proxies[0].GetRemark())
	assert.Equal(t, "Node2", sub.Proxies[1].GetRemark())
}
