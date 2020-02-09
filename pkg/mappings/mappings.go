package mappings

import (
	"github.com/gobwas/glob"
	"github.com/google/go-github/v29/github"
)

type (
	Label struct {
		name     string
		patterns []glob.Glob
	}
	Mappings struct {
		labels []Label
	}
)

func (ms Mappings) MatchedLabels(files []*github.CommitFile) (labels []string) {
	set := make(map[string]bool)
	for _, file := range files {
		for _, label := range ms.labels {
			if !set[label.name] && label.match(*file.Filename) {
				set[label.name] = true
				labels = append(labels, label.name)
			}
		}
	}
	return labels
}

func (l Label) match(name string) bool {
	for _, m := range l.patterns {
		if m.Match(name) {
			return true
		}
	}
	return false
}
