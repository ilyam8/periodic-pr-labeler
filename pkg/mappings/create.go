package mappings

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/gobwas/glob"
	"github.com/google/go-github/v29/github"
	"gopkg.in/yaml.v2"
)

type Repository interface {
	FileContent(filePath string) (*github.RepositoryContent, error)
}

func FromFile(filepath string) (*Mappings, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return newMappings(b)
}

func FromGitHub(filepath string, r Repository) (*Mappings, error) {
	content, err := r.FileContent(filepath)
	if err != nil {
		return nil, err
	}
	c, err := content.GetContent()
	return newMappings([]byte(c))
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

func newMappings(data []byte) (*Mappings, error) {
	var userMappings map[string]interface{}
	if err := yaml.Unmarshal(data, &userMappings); err != nil {
		return nil, fmt.Errorf("label mappings unmarshaling: %v", err)
	}
	if len(userMappings) == 0 {
		return nil, errors.New("empty label mappings")
	}

	var mappings Mappings
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
		mappings.labels = append(mappings.labels, Label{name: name, patterns: patterns})
	}
	return &mappings, nil
}

func mappingToSlice(mapping interface{}) ([]string, error) {
	val := reflect.Indirect(reflect.ValueOf(mapping))
	var rv []string
	err := mappingValueToSlice(val, &rv)
	return rv, err
}

func mappingValueToSlice(value reflect.Value, rv *[]string) error {
	if !value.IsValid() {
		return errors.New("invalid mapping value")
	}
	switch value.Kind() {
	case reflect.String:
		return convertString(rv, value)
	case reflect.Interface:
		return convertInterface(rv, value)
	case reflect.Slice:
		return convertSlice(rv, value)
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
