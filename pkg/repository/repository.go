package repository

import (
	"context"
	"fmt"

	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
)

func newGitHubClient(token string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return github.NewClient(tc)
}

// New creates new Repository.
func New(conf Config) *Repository {
	return &Repository{
		owner:  conf.Owner,
		name:   conf.Name,
		Client: newGitHubClient(conf.Token),
	}
}

// Config is Repository configuration.
type Config struct {
	Owner string
	Name  string
	Token string
}

// Repository represents github repository.
type Repository struct {
	owner string
	name  string
	*github.Client
}

// Name is repository owner.
func (r Repository) Owner() string {
	return r.owner
}

// Name is repository name.
func (r Repository) Name() string {
	return r.name
}

// FileContent returns content of a single file. If filepath doesnt reference to a file it returns nil.
func (r Repository) FileContent(filepath string) (*github.RepositoryContent, error) {
	content, _, _, err := r.Repositories.GetContents(context.TODO(), r.Owner(), r.Name(), filepath, nil)
	if content == nil && err == nil {
		err = fmt.Errorf("'%s' is not a file", filepath)
	}
	return content, err
}

// OpenPullRequests lists all the pull requests in the open state.
func (r Repository) OpenPullRequests() ([]*github.PullRequest, error) {
	opts := &github.PullRequestListOptions{State: "open", Sort: "updated"}
	var pulls []*github.PullRequest
	for {
		list, resp, err := r.PullRequests.List(context.TODO(), r.Owner(), r.Name(), opts)
		pulls = append(pulls, list...)
		if err != nil || resp.NextPage == 0 {
			return pulls, err
		}
		opts.Page = resp.NextPage
	}
}

// PullRequestModifiedFiles lists the files in a pull request.
func (r Repository) PullRequestModifiedFiles(number int) ([]*github.CommitFile, error) {
	files, _, err := r.PullRequests.ListFiles(context.Background(), r.Owner(), r.Name(), number, nil)
	return files, err
}

// AddLabelsToPullRequest adds labels to a pull request.
func (r Repository) AddLabelsToPullRequest(number int, labels []string) error {
	_, _, err := r.Issues.AddLabelsToIssue(context.Background(), r.Owner(), r.Name(), number, labels)
	return err
}
