package labeling

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v45/github"
)

func prepareRepository() *mockRepository {
	return &mockRepository{
		owner:      "owner",
		name:       "name",
		pullsFiles: make(map[int][]*github.CommitFile),
	}
}

type mockRepository struct {
	owner                         string
	name                          string
	errOnOpenPullRequests         bool
	errOnPullRequestModifiedFiles bool
	errOnAddLabelsToPullRequest   bool
	pulls                         []*github.PullRequest
	pullsFiles                    map[int][]*github.CommitFile
}

func (r *mockRepository) Owner() string {
	return r.owner
}

func (r *mockRepository) Name() string {
	return r.name
}

func (r *mockRepository) OpenPullRequests() ([]*github.PullRequest, error) {
	if r.errOnOpenPullRequests {
		return nil, errors.New("mock OpenPullRequests error")
	}
	var pulls []*github.PullRequest
	for _, p := range r.pulls {
		if *p.State == "open" {
			pulls = append(pulls, p)
		}
	}
	return pulls, nil
}

func (r *mockRepository) PullRequestModifiedFiles(number int) ([]*github.CommitFile, error) {
	if r.errOnPullRequestModifiedFiles {
		return nil, errors.New("mock PullRequestModifiedFiles error")
	}
	files, ok := r.pullsFiles[number]
	if !ok {
		return nil, fmt.Errorf("couldnt find PR#%d commit files", number)
	}
	return files, nil
}

func (r *mockRepository) AddLabelsToPullRequest(prNum int, labels []string) error {
	if r.errOnAddLabelsToPullRequest {
		return errors.New("mock AddLabelsToPullRequest error")
	}
	pr, err := r.findPullRequest(prNum)
	if err != nil {
		return err
	}
	diff := difference(labels, pr.Labels)
	for _, name := range diff {
		name := name
		pr.Labels = append(pr.Labels, &github.Label{Name: &name})
	}
	return nil
}

func (r *mockRepository) findPullRequest(num int) (*github.PullRequest, error) {
	for _, p := range r.pulls {
		if *p.Number == num {
			return p, nil
		}
	}
	return nil, fmt.Errorf("pull request %d not found", num)
}

func (r *mockRepository) addPullRequest(pull *github.PullRequest, files []*github.CommitFile) {
	i := len(r.pulls)
	pull.Number = &i
	r.pulls = append(r.pulls, pull)
	r.pullsFiles[i] = files
}

func prepareMappings() *mockMappings {
	return &mockMappings{}
}

type mockMappings struct{}

func (mockMappings) MatchedLabels(files []*github.CommitFile) (labels []string) {
	set := make(map[string]bool)
	for _, f := range files {
		if strings.HasPrefix(*f.Filename, "collectors/") {
			set["collectors"] = true
		}
		if strings.HasPrefix(*f.Filename, "collectors/python.d.plugin/") {
			set["python.d"] = true
		}
		if strings.HasPrefix(*f.Filename, "collectors/python.d.plugin/apache/") {
			set["python.d/apache"] = true
		}
		if strings.HasPrefix(*f.Filename, "collectors/charts.d.plugin/") {
			set["charts.d"] = true
		}
		if strings.HasPrefix(*f.Filename, "collectors/charts.d.plugin/apache/") {
			set["charts.d/apache"] = true
		}
	}
	for v := range set {
		labels = append(labels, v)

	}
	return labels
}
