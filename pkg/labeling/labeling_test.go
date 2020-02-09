package labeling

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-github/v29/github"
)

func TestNew(t *testing.T) {

}

func TestLabeler_ApplyLabels(t *testing.T) {
	rs := prepareRepository()
	ms := &mockMappings{}

	l := New(rs, ms)
	_ = l.ApplyLabels()
}

type mockMappings struct{}

func (mockMappings) MatchedLabels(files []*github.CommitFile) (labels []string) {
	set := make(map[string]bool)

	for _, f := range files {
		switch {
		case strings.HasPrefix(*f.Filename, "collectors"):
			set["collectors"] = true
		}
	}
	for v := range set {
		labels = append(labels, v)

	}
	return labels
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

func (r mockRepository) OpenPullRequests() ([]*github.PullRequest, error) {
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

func (r mockRepository) PullRequestModifiedFiles(number int) ([]*github.CommitFile, error) {
	if r.errOnPullRequestModifiedFiles {
		return nil, errors.New("mock PullRequestModifiedFiles error")
	}
	files, ok := r.pullsFiles[number]
	if !ok {
		return nil, fmt.Errorf("couldnt find pr#%d commit files", number)
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

func (r mockRepository) findPullRequest(num int) (*github.PullRequest, error) {
	for _, p := range r.pulls {
		if *p.Number == num {
			return p, nil
		}
	}
	return nil, fmt.Errorf("pull request %d not found", num)
}

func (r mockRepository) Owner() string {
	return r.owner
}

func (r mockRepository) Name() string {
	return r.name
}

func prepareRepository() *mockRepository {
	r := mockRepository{
		owner:      "owner",
		name:       "name",
		pullsFiles: make(map[int][]*github.CommitFile),
	}

	const open = "open"
	prData := []struct {
		title string
		state string
		files []string
	}{
		{
			title: "Title",
			state: open,
			files: []string{"collectors/charts.d.plugin/charts.d.plugin.in"},
		},
	}

	for i, data := range prData {
		pr, files := preparePullRequestCommitFiles(i, data.title, data.state, data.files...)
		r.pulls = append(r.pulls, pr)
		r.pullsFiles[i] = files
	}

	return &r
}

func preparePullRequestCommitFiles(number int, title, state string, files ...string) (*github.PullRequest, []*github.CommitFile) {
	pr := &github.PullRequest{
		Number: &number,
		Title:  &title,
		State:  &state,
	}

	var cf []*github.CommitFile
	for _, name := range files {
		name := name
		cf = append(cf, &github.CommitFile{Filename: &name})
	}
	return pr, cf
}
