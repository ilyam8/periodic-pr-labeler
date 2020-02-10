package labeling

import (
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
		if pull.Number == nil {
			continue
		}

		files, err := l.PullRequestModifiedFiles(*pull.Number)
		if err != nil {
			return err
		}

		expected := l.MatchedLabels(files)
		if len(expected) == 0 {
			log.Debugf("[NO MATCH] PR %s/%s#%d '%s'", l.Owner(), l.Name(), *pull.Number, safeString(pull.Title))
			continue
		}

		if !shouldAddLabels(expected, pull.Labels) {
			log.Debugf("[HAVE ALL] PR %s/%s#%d '%s'", l.Owner(), l.Name(), *pull.Number, safeString(pull.Title))
			continue
		}

		log.Debugf("[SHOULD HAVE] PR %s/%s#%d '%s', LABELS: %v", l.Owner(), l.Name(), *pull.Number, safeString(pull.Title), expected)
		if l.DryRun {
			continue
		}

		log.Infof("[APPLYING] PR %s/%s#%d '%s', LABELS:", l.Owner(), l.Name(), *pull.Number, safeString(pull.Title), expected)
		if err := l.AddLabelsToPullRequest(*pull.Number, expected); err != nil {
			return err
		}
	}
	return nil
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
		existingSet[safeString(v.Name)] = struct{}{}
	}
	var diff []string
	for _, v := range expected {
		if _, ok := existingSet[v]; !ok {
			diff = append(diff, v)
		}
	}
	return diff
}

func safeString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}
