package labeling

import (
	"io/ioutil"

	"github.com/gobwas/glob"
	"github.com/google/go-github/v29/github"
	"gopkg.in/yaml.v2"
)

type (
	label struct {
		name     string
		patterns []glob.Glob
	}
	labels []label
)

func (l label) match(name string) bool {
	for _, m := range l.patterns {
		if m.Match(name) {
			return true
		}
	}
	return false
}

func (ls labels) matchedLabels(files []*github.CommitFile) (matched []string) {
	set := make(map[string]bool)
	for _, file := range files {
		for _, label := range ls {
			if !set[label.name] && label.match(*file.Filename) {
				set[label.name] = true
				matched = append(matched, label.name)
			}
		}
	}
	return matched
}

func newLabelsFromGitHub(r repositoryService, filePath string) (labels, error) {
	content, err := r.fileContent(filePath)
	if err != nil {
		return nil, err
	}
	c, err := content.GetContent()
	if err != nil {
		return nil, err
	}
	return newLabelsFromConfig([]byte(c))
}

func newLabelsFromFile(filePath string) (labels, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return newLabelsFromConfig(b)
}

func newLabelsFromConfig(from []byte) (labels, error) {
	var config map[string][]string
	if err := yaml.Unmarshal(from, &config); err != nil {
		return nil, err
	}

	ls := make(labels, len(config))
	for name, values := range config {
		patterns, err := newPatterns(values)
		if err != nil {
			return nil, err
		}
		ls = append(ls, label{name: name, patterns: patterns})
	}
	return ls, nil
}

func newPatterns(values []string) ([]glob.Glob, error) {
	var patterns []glob.Glob
	for _, v := range values {
		p, err := glob.Compile(v, '/')
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, p)
	}
	return patterns, nil
}
