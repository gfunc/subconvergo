package core

import (
	"fmt"
	"sync"
)

var (
	generators = make(map[string]Generator)
	mu         sync.RWMutex
)

// RegisterGenerator registers a generator
func RegisterGenerator(g Generator) {
	mu.Lock()
	defer mu.Unlock()
	generators[g.Name()] = g
}

// GetGenerator returns a generator by name
func GetGenerator(name string) (Generator, error) {
	mu.RLock()
	defer mu.RUnlock()
	if g, ok := generators[name]; ok {
		return g, nil
	}
	return nil, fmt.Errorf("generator not found: %s", name)
}
