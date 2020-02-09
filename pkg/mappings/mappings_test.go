package mappings

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-github/v29/github"
	"github.com/stretchr/testify/assert"
)

func TestFromFile(t *testing.T) {
	tests := map[string]struct {
		input   string
		wantErr bool
	}{
		"valid configuration":       {input: "testdata/labeler.yaml"},
		"invalid configuration":     {input: "testdata/labeler_invalid.yaml", wantErr: true},
		"empty configuration":       {input: "testdata/labeler_empty.yaml", wantErr: true},
		"nonexistent configuration": {input: "testdata/labeler_nonexistent.yaml", wantErr: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ms, err := FromFile(test.input)

			if !test.wantErr {
				assert.NotNil(t, ms)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, ms)
				assert.Error(t, err)
			}
		})
	}
}

func TestFromGitHub(t *testing.T) {
	tests := map[string]struct {
		input   string
		wantErr bool
	}{
		"valid configuration":       {input: "testdata/labeler.yaml"},
		"invalid configuration":     {input: "testdata/labeler_invalid.yaml", wantErr: true},
		"empty configuration":       {input: "testdata/labeler_empty.yaml", wantErr: true},
		"nonexistent configuration": {input: "testdata/labeler_nonexistent.yaml", wantErr: true},
	}

	r := &mockRepository{}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ms, err := FromGitHub(test.input, r)

			if !test.wantErr {
				assert.NotNil(t, ms)
				assert.NoError(t, err)
			} else {
				assert.Nil(t, ms)
				assert.Error(t, err)
			}
		})
	}
}

func TestMappings_MatchedLabels(t *testing.T) {
	tests := []struct {
		input      []string
		wantLabels []string
	}{
		{
			input:      []string{".navarro/vertibird.plans"},
			wantLabels: []string{"area/57"},
		},
		{
			input:      []string{".navarro/main_base/power.armor"},
			wantLabels: []string{"area/57"},
		},
		{
			input:      []string{"arroyo/temple/test.room"},
			wantLabels: []string{"area/58"},
		},
		{
			input:      []string{"newreno/jungle.gym"},
			wantLabels: []string{"area/59"},
		},
		{
			input:      []string{".navarro/vertibird.plans", "arroyo/temple/test.room", "newreno/jungle.gym"},
			wantLabels: []string{"area/57", "area/58", "area/59"},
		},
		{
			input: []string{"newreno/westside/drunk.chapel"},
		},
		{
			input: []string{"brokenhils/down.town"},
		},
	}

	ms := prepareValidConfigurationMappings(t)

	for i, test := range tests {
		t.Run(fmt.Sprintf("test case #%d", i+1), func(t *testing.T) {
			files := prepareGithubCommitFiles(test.input)
			assert.Equal(t, test.wantLabels, ms.MatchedLabels(files))
		})
	}
}

type mockRepository struct{}

func (r mockRepository) FileContent(filePath string) (*github.RepositoryContent, error) {
	switch filePath {
	case "testdata/labeler.yaml":
		content := string(validConfig)
		return &github.RepositoryContent{Content: &content}, nil
	case "testdata/labeler_invalid.yaml":
		content := string(invalidConfig)
		return &github.RepositoryContent{Content: &content}, nil
	case "testdata/labeler_empty.yaml":
		content := string(emptyConfig)
		return &github.RepositoryContent{Content: &content}, nil
	}
	return nil, errors.New("mock FileContent error")
}

func prepareValidConfigurationMappings(t *testing.T) *Mappings {
	ms, err := FromFile("testdata/labeler.yaml")
	assert.NoError(t, err)
	return ms
}

func prepareGithubCommitFiles(names []string) []*github.CommitFile {
	files := make([]*github.CommitFile, 0, len(names))
	for _, name := range names {
		name := name
		files = append(files, &github.CommitFile{Filename: &name})
	}
	return files
}
