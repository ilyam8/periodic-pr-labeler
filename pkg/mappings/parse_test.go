package mappings

import (
	"os"
	"sort"
	"testing"

	"github.com/gobwas/glob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	validConfig, _   = os.ReadFile("testdata/labeler.yaml")
	emptyConfig, _   = os.ReadFile("testdata/labeler_empty.yaml")
	invalidConfig, _ = os.ReadFile("testdata/labeler_invalid.yaml")
)

func TestParse_testdata(t *testing.T) {
	assert.NotEmptyf(t, validConfig, "valid data")
	assert.NotEmptyf(t, invalidConfig, "invalid data")
	assert.NotEmptyf(t, emptyConfig, "empty data")
}

func TestParse(t *testing.T) {
	tests := map[string]struct {
		input      []byte
		wantLabels []*label
		wantErr    bool
	}{
		"valid configuration": {input: validConfig, wantLabels: []*label{
			{name: "build", patterns: patterns{
				{positive: true, raw: "build/**/*", Glob: globMust("build/**/*")},
			}},
			{name: "collectors", patterns: patterns{
				{positive: false, raw: "collectors/apps.plugin/*", Glob: globMust("collectors/apps.plugin/*")},
				{positive: false, raw: "collectors/README.md", Glob: globMust("collectors/README.md")},
				{positive: true, raw: "collectors/*", Glob: globMust("collectors/*")},
				{positive: true, raw: "collectors/**/*", Glob: globMust("collectors/**/*")},
			}},
			{name: "github", patterns: patterns{
				{positive: true, raw: ".github/*", Glob: globMust(".github/*")},
				{positive: true, raw: ".github/**/*", Glob: globMust(".github/**/*")},
			}},
		}},
		"invalid configuration": {input: invalidConfig, wantErr: true},
		"empty configuration":   {input: emptyConfig, wantErr: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ms, err := Parse(test.input)

			if !test.wantErr {
				require.NotNil(t, ms)
				require.NoError(t, err)

				sort.Slice(ms.labels, func(i, j int) bool { return ms.labels[i].name < ms.labels[j].name })
				assert.Equal(t, test.wantLabels, ms.labels)
			} else {
				assert.Nil(t, ms)
				assert.Error(t, err)
			}
		})
	}
}

func globMust(pattern string) glob.Glob {
	return glob.MustCompile(pattern, '/')
}
