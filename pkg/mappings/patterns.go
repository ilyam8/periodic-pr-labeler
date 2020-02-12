package mappings

import (
	"strings"

	"github.com/gobwas/glob"
)

type (
	pattern struct {
		positive bool
		raw      string
		glob.Glob
	}
	patterns []*pattern
)

func (ps patterns) match(name string) bool {
	for _, p := range ps {
		if p.Match(name) {
			return p.positive
		}
	}
	return false
}

func newPatterns(values []string) (patterns, error) {
	var ps patterns
	for _, value := range values {
		p, err := newPattern(value)
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}
	return ps, nil
}

func newPattern(value string) (*pattern, error) {
	positive := !(value[0] == '!' && len(value) > 1)
	if !positive {
		value = value[1:]
	}
	value = strings.TrimSpace(value)

	g, err := glob.Compile(value, '/')
	if err != nil {
		return nil, err
	}

	p := pattern{
		positive: positive,
		raw:      value,
		Glob:     g,
	}
	return &p, nil
}
