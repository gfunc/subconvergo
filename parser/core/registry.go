package core

import (
	"fmt"
	"sync"

	"github.com/gfunc/subconvergo/proxy/core"
)

var (
	parsers []LineParser
	mu      sync.RWMutex
)

// RegisterParser adds a parser to the registry
func RegisterParser(p LineParser) {
	mu.Lock()
	defer mu.Unlock()
	parsers = append(parsers, p)
}

// ParseLine tries to parse a line using registered parsers
func ParseLine(line string) (core.SubconverterProxy, error) {
	mu.RLock()
	defer mu.RUnlock()

	for _, p := range parsers {
		if p.CanParse(line) {
			return p.Parse(line)
		}
	}
	return nil, fmt.Errorf("no parser found for line: %s", line)
}
