package labeling

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"

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
	var userMappings map[string]interface{}
	if err := yaml.Unmarshal(data, &userMappings); err != nil {
		return nil, fmt.Errorf("label mappings unmarshaling: %v", err)
	}
	if len(userMappings) == 0 {
		return nil, errors.New("empty label mappings")
	}

	ls := make(mappings, len(userMappings))
	for name, value := range userMappings {
		values, err := mappingToSlice(value)
		if err != nil {
			return nil, fmt.Errorf("mapping label '%s': %v", name, err)
		}
		if len(values) == 0 {
			return nil, fmt.Errorf("mapping label '%s' has no pattern(s)", name)
		}

		patterns, err := newPatterns(values)
		if err != nil {
			return nil, fmt.Errorf("mapping label '%s': %v", name, err)
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

func mappingToSlice(i interface{}) ([]string, error) {
	val := reflect.Indirect(reflect.ValueOf(i))
	var rv []string
	err := mappingValueToSlice(val, &rv)
	return rv, err
}

func mappingValueToSlice(value reflect.Value, slice *[]string) error {
	if !value.IsValid() {
		return errors.New("invalid mapping value")
	}
	switch value.Kind() {
	case reflect.String:
		return convertString(slice, value)
	case reflect.Interface:
		return convertInterface(slice, value)
	case reflect.Slice:
		return convertSlice(slice, value)
	default:
		return fmt.Errorf("unsupported mapping value type: %v", value.Kind())
	}
}

func convertString(rv *[]string, value reflect.Value) error {
	*rv = append(*rv, value.String())
	return nil
}

func convertInterface(rv *[]string, value reflect.Value) error {
	return mappingValueToSlice(reflect.ValueOf(value.Interface()), rv)
}

func convertSlice(rv *[]string, value reflect.Value) error {
	for i := 0; i < value.Len(); i++ {
		if err := mappingValueToSlice(value.Index(i), rv); err != nil {
			return err
		}
	}
	return nil
}
