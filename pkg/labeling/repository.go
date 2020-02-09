package labeling

import (
	"context"
	"fmt"

	"github.com/google/go-github/v29/github"
)

type repository struct {
	Owner string
	Name  string
	*github.Client
}

func (r repository) owner() string {
	return r.Owner
}

func (r repository) name() string {
	return r.Name
}

func (r *repository) fileContent(filePath string) (*github.RepositoryContent, error) {
	content, _, _, err := r.Repositories.GetContents(context.TODO(), r.owner(), r.name(), filePath, nil)
	if content == nil && err == nil {
		err = fmt.Errorf("%s is not a file", filePath)
	}
	return content, err
}

func (r *repository) openPullRequests() ([]*github.PullRequest, error) {
	opts := &github.PullRequestListOptions{State: "open", Sort: "updated"}
	var pulls []*github.PullRequest
	for {
		list, resp, err := r.PullRequests.List(context.Background(), r.owner(), r.name(), opts)
		pulls = append(pulls, list...)
		if err != nil || resp.NextPage == 0 {
			return pulls, err
		}
		opts.Page = resp.NextPage
	}
}

func (r *repository) pullRequestModifiedFiles(prNum int) ([]*github.CommitFile, error) {
	files, _, err := r.PullRequests.ListFiles(context.Background(), r.owner(), r.name(), prNum, nil)
	return files, err
}

func (r *repository) addLabelsToPullRequest(prNum int, labels []string) error {
	_, _, err := r.Issues.AddLabelsToIssue(context.Background(), r.owner(), r.name(), prNum, labels)
	return err
}
