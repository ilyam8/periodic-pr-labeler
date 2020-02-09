package mappings

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	validConfig, _   = ioutil.ReadFile("testdata/labeler.yaml")
	emptyConfig, _   = ioutil.ReadFile("testdata/labeler_empty.yaml")
	invalidConfig, _ = ioutil.ReadFile("testdata/labeler_invalid.yaml")
)

func TestParse_testdata(t *testing.T) {
	assert.NotEmptyf(t, validConfig, "valid data")
	assert.NotEmptyf(t, invalidConfig, "invalid data")
	assert.NotEmptyf(t, emptyConfig, "empty data")
}

func TestParse(t *testing.T) {
	tests := map[string]struct {
		input      []byte
		wantLabels map[string]int
		wantErr    bool
	}{
		"valid configuration":   {input: validConfig, wantLabels: map[string]int{"area/57": 2, "area/58": 1, "area/59": 2}},
		"invalid configuration": {input: invalidConfig, wantErr: true},
		"empty configuration":   {input: emptyConfig, wantErr: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ms, err := Parse(test.input)

			if !test.wantErr {
				require.NotNil(t, ms)
				require.NoError(t, err)
				assert.Lenf(t, ms.labels, len(test.wantLabels), "mappings has %d labels, but expected %d", len(ms.labels), len(test.wantLabels))
				for _, l := range ms.labels {
					num, ok := test.wantLabels[l.name]
					assert.Truef(t, ok, "label '%s' not in mappings", l.name)
					v, ok := l.matcher.(globOrMatcher)
					require.Truef(t, ok, "label '%s' not globOrMatcher", l.name)
					assert.Lenf(t, v, num, "label '%s' has %d matchers, but expected %d", l.name, len(v), num)
				}
			} else {
				assert.Nil(t, ms)
				assert.Error(t, err)
			}
		})
	}
}
