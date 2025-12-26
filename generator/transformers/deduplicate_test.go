package transformers

import (
	"testing"

	"github.com/gfunc/subconvergo/config"
	pc "github.com/gfunc/subconvergo/proxy/core"
)

func TestDeduplicateTransformer(t *testing.T) {
	transformer := NewDeduplicateTransformer()

	proxies := []pc.ProxyInterface{
		&pc.BaseProxy{Remark: "Node"},
		&pc.BaseProxy{Remark: "Node"},
		&pc.BaseProxy{Remark: "Node"},
		&pc.BaseProxy{Remark: "Other"},
	}

	result, err := transformer.Transform(proxies, config.Global)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result) != 4 {
		t.Errorf("Expected 4 proxies, got %d", len(result))
	}

	expectedNames := []string{"Node", "Node_1", "Node_2", "Other"}
	for i, p := range result {
		if p.GetRemark() != expectedNames[i] {
			t.Errorf("Proxy %d: expected name %q, got %q", i, expectedNames[i], p.GetRemark())
		}
	}
}

func TestDeduplicateTransformer_NoDuplicates(t *testing.T) {
	transformer := NewDeduplicateTransformer()

	proxies := []pc.ProxyInterface{
		&pc.BaseProxy{Remark: "Node A"},
		&pc.BaseProxy{Remark: "Node B"},
		&pc.BaseProxy{Remark: "Node C"},
	}

	result, err := transformer.Transform(proxies, config.Global)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 proxies, got %d", len(result))
	}

	expectedNames := []string{"Node A", "Node B", "Node C"}
	for i, p := range result {
		if p.GetRemark() != expectedNames[i] {
			t.Errorf("Proxy %d: expected name %q, got %q", i, expectedNames[i], p.GetRemark())
		}
	}
}

func TestDeduplicateTransformer_ExistingSuffix(t *testing.T) {
	transformer := NewDeduplicateTransformer()

	proxies := []pc.ProxyInterface{
		&pc.BaseProxy{Remark: "Node"},
		&pc.BaseProxy{Remark: "Node_1"},
		&pc.BaseProxy{Remark: "Node"},
	}

	result, err := transformer.Transform(proxies, config.Global)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 proxies, got %d", len(result))
	}

	// Node -> Node
	// Node_1 -> Node_1
	// Node -> Node_2 (because Node_1 is taken)
	expectedNames := []string{"Node", "Node_1", "Node_2"}
	for i, p := range result {
		if p.GetRemark() != expectedNames[i] {
			t.Errorf("Proxy %d: expected name %q, got %q", i, expectedNames[i], p.GetRemark())
		}
	}
}
