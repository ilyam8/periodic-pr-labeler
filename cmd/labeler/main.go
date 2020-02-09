package main

import (
	"os"

	"github.com/ilyam8/periodic-pr-labeler/pkg/labeling"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

type options struct {
	RepoSlug           string `short:"r" long:"repository" description:"github repository slug"`
	Token              string `short:"t" long:"token" description:"github token"`
	LabelMappings      string `short:"m" long:"label-mappings" description:"label mappings file on github"`
	LocalLabelMappings string `short:"M" long:"label-mappings-local" description:"label mappings file on the local system"`
	DryRun             bool   `short:"d" long:"dry-run" description:"dry run, labels won't be applied, only reported"`
}

func parseCLI() options {
	var opt options
	parser := flags.NewParser(&opt, flags.Default)
	parser.Name = "periodic-pr-labeler"
	parser.Usage = "[OPTION]..."

	if _, err := parser.ParseArgs(os.Args); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}
	return opt
}

func applyFromEnv(opts *options) {
	if repoSlug, ok := os.LookupEnv("GITHUB_REPOSITORY"); ok && opts.RepoSlug == "" {
		opts.RepoSlug = repoSlug
	}
	if token, ok := os.LookupEnv("GITHUB_TOKEN"); ok && opts.Token == "" {
		opts.Token = token
	}
	if mappings, ok := os.LookupEnv("LABEL_MAPPINGS_FILE"); ok && opts.LabelMappings == "" {
		opts.LabelMappings = mappings
	}
}

func main() {
	opts := parseCLI()
	applyFromEnv(&opts)

	if opts.DryRun {
		log.SetLevel(log.DebugLevel)
	}

	conf := labeling.Config{
		RepoSlug:            opts.RepoSlug,
		Token:               opts.Token,
		LabelMappingsGithub: opts.LabelMappings,
		LabelMappingsLocal:  opts.LocalLabelMappings,
		DryRun:              opts.DryRun,
	}

	labeler, err := labeling.New(conf)
	if err != nil {
		log.Fatal(err)
	}

	if err := labeler.ApplyLabels(); err != nil {
		log.Fatal(err)
	}
}
