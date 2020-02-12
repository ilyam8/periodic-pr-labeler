package mappings

import (
	"github.com/google/go-github/v29/github"
	"io/ioutil"
)

type (
	label struct {
		name string
		patterns
	}
	Mappings struct {
		labels []*label
	}
)

type Repository interface {
	FileContent(filePath string) (*github.RepositoryContent, error)
}

func FromFile(filepath string) (*Mappings, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

func FromGitHub(filepath string, r Repository) (*Mappings, error) {
	content, err := r.FileContent(filepath)
	if err != nil {
		return nil, err
	}
	c, err := content.GetContent()
	return Parse([]byte(c))
}

func (ms Mappings) MatchedLabels(files []*github.CommitFile) (labels []string) {
	set := make(map[string]bool)
	for _, file := range files {
		for _, l := range ms.labels {
			if !set[l.name] && l.match(file.GetFilename()) {
				set[l.name] = true
				labels = append(labels, l.name)
			}
		}
	}
	return labels
}
