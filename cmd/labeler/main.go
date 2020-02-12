package main

import (
	"errors"
	"os"
	"strings"

	"github.com/ilyam8/periodic-pr-labeler/pkg/labeling"
	"github.com/ilyam8/periodic-pr-labeler/pkg/mappings"
	"github.com/ilyam8/periodic-pr-labeler/pkg/repository"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

type options struct {
	RepoSlug           string `short:"r" long:"repository" description:"GitHub repository slug"`
	Token              string `short:"t" long:"token" description:"GitHub token"`
	LabelMappings      string `short:"m" long:"label-mappings" description:"Label mappings file on github" default:".github/labeler.yml"`
	LabelMappingsLocal string `short:"M" long:"label-mappings-local" description:"Label mappings file on the local system"`
	DryRun             bool   `short:"d" long:"dry-run" description:"Dry run, labels won't be applied, only reported"`
}

func validateOptions(opts options) error {
	if opts.RepoSlug == "" {
		return errors.New("repository slug config parameter not set")
	}
	if opts.Token == "" {
		return errors.New("token config parameter not set")
	}
	if opts.LabelMappingsLocal == "" && opts.LabelMappings == "" {
		return errors.New("label mappings config parameter not set")
	}
	return nil
}

func parseCLI() options {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "labeler"
	parser.Usage = "[OPTION]..."

	if _, err := parser.ParseArgs(os.Args); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}
	return opts
}

func applyFromEnv(opts *options) {
	if repoSlug, ok := os.LookupEnv("GITHUB_REPOSITORY"); ok && opts.RepoSlug == "" {
		opts.RepoSlug = repoSlug
	}
	if token, ok := os.LookupEnv("GITHUB_TOKEN"); ok && opts.Token == "" {
		opts.Token = token
	}
	if labelMappings, ok := os.LookupEnv("LABEL_MAPPINGS_FILE"); ok && opts.LabelMappings == "" {
		opts.LabelMappings = labelMappings
	}
}

func extractOwnerName(repoSlug string) (owner, name string, ok bool) {
	parts := strings.Split(strings.TrimSpace(repoSlug), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func newRepositoryService(opts options) *repository.Repository {
	owner, name, ok := extractOwnerName(opts.RepoSlug)
	if !ok {
		log.Fatalf("repository slug config parameter bad syntax ('%s')", opts.RepoSlug)
	}
	conf := repository.Config{
		Owner: owner,
		Name:  name,
		Token: opts.Token,
	}
	return repository.New(conf)
}

func newMappingsService(opts options, rs *repository.Repository) (ms *mappings.Mappings) {
	var err error
	if opts.LabelMappingsLocal != "" {
		ms, err = mappings.FromFile(opts.LabelMappingsLocal)
	} else {
		ms, err = mappings.FromGitHub(opts.LabelMappings, rs)
	}
	if err != nil {
		log.Fatal(err)
	}
	return ms
}

func newLabelingService(rs *repository.Repository, ms *mappings.Mappings, opts options) *labeling.Labeler {
	labSvc := labeling.New(rs, ms)
	if opts.DryRun {
		labSvc.DryRun = true
	}
	return labSvc
}

func main() {
	opts := parseCLI()
	applyFromEnv(&opts)

	if err := validateOptions(opts); err != nil {
		log.Fatal(err)
	}

	if opts.DryRun {
		log.SetLevel(log.DebugLevel)
	}

	repoSvc := newRepositoryService(opts)
	mapSvc := newMappingsService(opts, repoSvc)
	labSvc := newLabelingService(repoSvc, mapSvc, opts)

	if err := labSvc.ApplyLabels(); err != nil {
		log.Fatal(err)
	}
}
