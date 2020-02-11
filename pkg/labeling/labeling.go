package labeling

import (
	"fmt"

	"github.com/google/go-github/v29/github"
	log "github.com/sirupsen/logrus"
)

type Repository interface {
	OpenPullRequests() ([]*github.PullRequest, error)
	PullRequestModifiedFiles(number int) ([]*github.CommitFile, error)
	AddLabelsToPullRequest(number int, labels []string) error
	Owner() string
	Name() string
}

type Mappings interface {
	MatchedLabels([]*github.CommitFile) (labels []string)
}

type Labeler struct {
	DryRun bool
	Repository
	Mappings
}

func New(r Repository, m Mappings) *Labeler {
	return &Labeler{
		Repository: r,
		Mappings:   m,
	}
}

func (l Labeler) ApplyLabels() (err error) {
	pulls, err := l.OpenPullRequests()
	if err != nil {
		return err
	}
	log.Debugf("found %d open pull requests", len(pulls))
	return l.applyLabels(pulls)
}

func (l Labeler) applyLabels(pulls []*github.PullRequest) error {
	for _, pull := range pulls {
		files, err := l.PullRequestModifiedFiles(pull.GetNumber())
		if err != nil {
			return err
		}

		expected := l.MatchedLabels(files)
		if len(expected) == 0 {
			log.WithField("labels", "no match").Info(l.fullName(pull))
			continue
		}

		if !shouldAddLabels(expected, pull.Labels) {
			log.WithField("labels", "has all").Debug(l.fullName(pull))
			continue
		}

		log.WithField("labels", expected).Debugf("%s [dry run]", l.fullName(pull))
		if l.DryRun {
			continue
		}

		log.WithField("labels", expected).Infof("%s [applying]", l.fullName(pull))
		if err := l.AddLabelsToPullRequest(pull.GetNumber(), expected); err != nil {
			return err
		}
	}
	return nil
}

func (l Labeler) fullName(pull *github.PullRequest) string {
	return fmt.Sprintf("PR %s/%s#%d", l.Owner(), l.Name(), pull.GetNumber())
}

func shouldAddLabels(expected []string, existing []*github.Label) bool {
	switch {
	case len(expected) == 0:
		return false
	case len(expected) > len(existing):
		return true
	}
	return len(difference(expected, existing)) > 0
}

func difference(expected []string, existing []*github.Label) []string {
	existingSet := make(map[string]struct{}, len(existing))
	for _, v := range existing {
		existingSet[v.GetName()] = struct{}{}
	}
	var diff []string
	for _, v := range expected {
		if _, ok := existingSet[v]; !ok {
			diff = append(diff, v)
		}
	}
	return diff
}
