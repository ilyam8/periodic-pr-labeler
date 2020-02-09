package mappings

import "github.com/gobwas/glob"

type globOrMatcher []glob.Glob

func (gm globOrMatcher) Match(name string) bool {
	for _, m := range gm {
		if m.Match(name) {
			return true
		}
	}
	return false
}

func newGlobOrMatcher(values []string) (globOrMatcher, error) {
	var gm globOrMatcher
	for _, v := range values {
		p, err := glob.Compile(v, '/')
		if err != nil {
			return nil, err
		}
		gm = append(gm, p)
	}
	return gm, nil
}
