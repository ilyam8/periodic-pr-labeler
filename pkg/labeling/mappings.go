package labeling

import (
	"errors"
	"fmt"
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
	mappings []label
)

func (l label) match(name string) bool {
	for _, m := range l.patterns {
		if m.Match(name) {
			return true
		}
	}
	return false
}

func (ms mappings) matchedLabels(files []*github.CommitFile) (labels []string) {
	set := make(map[string]bool)
	for _, file := range files {
		for _, label := range ms {
			if !set[label.name] && label.match(*file.Filename) {
				set[label.name] = true
				labels = append(labels, label.name)
			}
		}
	}
	return labels
}

func newMappingsFromGitHub(r repositoryService, filePath string) (mappings, error) {
	content, err := r.fileContent(filePath)
	if err != nil {
		return nil, err
	}
	c, err := content.GetContent()
	if err != nil {
		return nil, err
	}
	return newMappings([]byte(c))
}

func newMappingsFromFile(filePath string) (mappings, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return newMappings(b)
}

func newMappings(data []byte) (mappings, error) {
	var userMappings map[string][]string
	if err := yaml.Unmarshal(data, &userMappings); err != nil {
		return nil, err
	}
	if err := validateUserMappings(userMappings); err != nil {
		return nil, err
	}

	ls := make(mappings, len(userMappings))
	for name, values := range userMappings {
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

func validateUserMappings(mappings map[string][]string) error {
	if len(mappings) == 0 {
		return errors.New("empty mappings mapping")
	}
	for name, values := range mappings {
		if len(values) == 0 {
			return fmt.Errorf("label '%s' has no patterns", name)
		}
		for i, v := range values {
			if v == "" {
				return fmt.Errorf("label '%s' pattern %d is empty", name, i+1)
			}
		}
	}
	return nil
}
