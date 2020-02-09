package labeling

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v29/github"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type repositoryService interface {
	fileContent(filePath string) (*github.RepositoryContent, error)
	openPullRequests() ([]*github.PullRequest, error)
	pullRequestModifiedFiles(prNum int) ([]*github.CommitFile, error)
	addLabelsToPullRequest(prNum int, labels []string) error
	owner() string
	name() string
}

type (
	Config struct {
		RepoSlug            string
		Token               string
		LabelMappingsGithub string
		LabelMappingsLocal  string
		DryRun              bool
	}
	Labeler struct {
		labelMappingsGithub string
		labelMappingsLocal  string
		dryRun              bool
		repo                repositoryService
		mappings            mappings
	}
)

func validateConfig(conf Config) error {
	if conf.RepoSlug == "" {
		return errors.New("empty repository slug")
	}
	if conf.Token == "" {
		return errors.New("empty token")
	}
	if conf.LabelMappingsLocal == "" && conf.LabelMappingsGithub == "" {
		return errors.New("empty label mappings path")
	}
	return nil
}

func extractOwnerName(repoSlug string) (owner, name string, err error) {
	parts := strings.Split(strings.TrimSpace(repoSlug), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("bad format repository slug")
	}
	return parts[0], parts[1], nil
}

func newGitHubClient(token string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return github.NewClient(tc)
}

func New(conf Config) (*Labeler, error) {
	if err := validateConfig(conf); err != nil {
		return nil, fmt.Errorf("error on config validation: %v", err)
	}
	owner, name, err := extractOwnerName(conf.RepoSlug)
	if err != nil {
		return nil, fmt.Errorf("error extracting repository owner/name: %v", err)
	}

	r := repository{
		Owner:  owner,
		Name:   name,
		Client: newGitHubClient(conf.Token),
	}
	l := &Labeler{
		repo:                &r,
		labelMappingsGithub: conf.LabelMappingsGithub,
		labelMappingsLocal:  conf.LabelMappingsLocal,
		dryRun:              conf.DryRun,
	}

	return l, nil
}

func (l *Labeler) ApplyLabels() error {
	if err := l.initLabels(); err != nil {
		return err
	}
	pulls, err := l.repo.openPullRequests()
	if err != nil {
		return err
	}
	return l.applyLabels(pulls)
}

func (l *Labeler) initLabels() (err error) {
	if l.labelMappingsLocal != "" {
		l.mappings, err = newMappingsFromFile(l.labelMappingsLocal)
	} else {
		l.mappings, err = newMappingsFromGitHub(l.repo, l.labelMappingsGithub)
	}
	return err
}

func (l *Labeler) applyLabels(pulls []*github.PullRequest) error {
	for _, pr := range pulls {
		expected, err := l.expectedLabels(pr)
		if err != nil {
			return err
		}
		if !shouldApplyLabels(expected, pr.Labels) {
			continue
		}
		log.Infof("PR %s/%s#%d should have following labels: %v (%s)", l.repo.owner(), l.repo.name(), *pr.Number, expected, *pr.Title)
		if l.dryRun {
			continue
		}
		if err := l.repo.addLabelsToPullRequest(*pr.Number, expected); err != nil {
			return err
		}
	}
	return nil
}

func (l *Labeler) expectedLabels(pr *github.PullRequest) ([]string, error) {
	files, err := l.repo.pullRequestModifiedFiles(*pr.Number)
	if err != nil {
		return nil, err
	}
	return l.mappings.matchedLabels(files), nil
}

func shouldApplyLabels(expected []string, existing []*github.Label) bool {
	switch {
	case len(expected) == 0:
		return false
	case len(expected) > len(existing):
		return true
	}
	return hasDifference(expected, existing)
}

func hasDifference(expected []string, existing []*github.Label) bool {
	existingSet := make(map[string]struct{}, len(existing))
	for _, v := range existing {
		existingSet[*v.Name] = struct{}{}
	}
	for _, v := range expected {
		if _, ok := existingSet[v]; !ok {
			return true
		}
	}
	return false
}
