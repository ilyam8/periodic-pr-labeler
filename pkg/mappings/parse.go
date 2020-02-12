package mappings

import (
	"errors"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v2"
)

func Parse(conf []byte) (*Mappings, error) {
	var userMappings map[string]interface{}
	if err := yaml.Unmarshal(conf, &userMappings); err != nil {
		return nil, fmt.Errorf("label mappings unmarshaling: %v", err)
	}
	if len(userMappings) == 0 {
		return nil, errors.New("empty label mappings")
	}

	var mappings Mappings
	for name, value := range userMappings {
		l, err := parseLabel(name, value)
		if err != nil {
			return nil, err
		}
		mappings.labels = append(mappings.labels, l)
	}
	return &mappings, nil
}

func parseLabel(name string, value interface{}) (*label, error) {
	values, err := mappingToSlice(value)
	if err != nil {
		return nil, fmt.Errorf("mapping label '%s': %v", name, err)
	}
	if values = removeEmpty(values); len(values) == 0 {
		return nil, fmt.Errorf("mapping label '%s' has no pattern(s)", name)
	}

	ps, err := newPatterns(values)
	if err != nil {
		return nil, fmt.Errorf("mapping label '%s': %v", name, err)
	}
	return &label{name: name, patterns: ps}, nil
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

func removeEmpty(values []string) []string {
	var i int
	for _, v := range values {
		if v != "" {
			values[i] = v
			i++
		}
	}
	return values[:i]
}
