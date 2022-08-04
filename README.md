# Periodic Pull Requests Labeler

A GitHub action to automatically label all PRs **based on changed files** according to file patterns.

Action is meant to be run as periodic job. This is needed to workaround issues regarding
[lack of write access when executed from fork](https://help.github.com/en/actions/automating-your-workflow-with-github-actions/authenticating-with-the-github_token#permissions-for-the-github_token)
which is a common problem when using https://github.com/actions/labeler.

## Workflow

```yaml
name: Pull Request Labeler
on:
  schedule:
    - cron: '*/10 * * * *'
jobs:
  labeler:
    runs-on: ubuntu-latest
    steps:
      - uses: docker://docker.io/ilyam8/periodic-pr-labeler:v0.1.1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GITHUB_REPOSITORY: ${{ github.repository }}
          LABEL_MAPPINGS_FILE: .github/labeler.yml
```

## Label mappings file

This file is in [`YaML`](https://yaml.org/) format. It contains a list of labels and patterns to match to apply the label.
By default this action uses `.github/labeler.yml` located in repository from `GITHUB_REPOSITORY` as a source of pattern matchers.

```yaml
# Add 'label1' to any changes within 'example' folder or any subfolders
label1:
  - example/**/*

# Add 'label2' to any file changes within 'example2' folder
label2: example2/*

# Add 'label3' label to any change to *.spec.js files within the source dir
label3:
  - src/**/*.spec.js

# Add 'label4' label to any change within the 'core' package
label4:
  - package/core/*
  - package/core/**/*
```

## Path exclusion

Pattern can be negated to stop searching through the remaining patterns. 
To exclude path from searching prepend pattern with `!`.

Using path exclusion keep in mind:

- negated pattern must be quoted.
- order is not relevant. Add label condition: **match at least one positive pattern AND do not match any of negated patterns.**

```yaml
# Add 'label5' to any changes within 'package' folder or any subfolders except `core` and `installer` subfolders
label5:
  - package/*
  - package/**/*
  - "!package/core/*"
  - "!package/installer/*"
```

## Pattern syntax

This action uses [`gobwas/glob`](https://github.com/gobwas/glob) library for pattern matches.

```console
pattern:
    { term }

term:
    `*`         matches any sequence of non-separator characters
    `**`        matches any sequence of characters
    `?`         matches any single non-separator character
    `[` [ `!` ] { character-range } `]`
                character class (must be non-empty)
    `{` pattern-list `}`
                pattern alternatives
    c           matches character c (c != `*`, `**`, `?`, `\`, `[`, `{`, `}`)
    `\` c       matches character c

character-range:
    c           matches character c (c != `\\`, `-`, `]`)
    `\` c       matches character c
    lo `-` hi   matches character c for lo <= c <= hi

pattern-list:
    pattern { `,` pattern }
                comma-separated (without spaces) patterns
```

To get better understanding see [examples](https://github.com/gobwas/glob#example).

## CLI

See all available options:

```console
Usage:
  labeler [OPTION]...

Application Options:
  -r, --repository=           Github repository slug
  -t, --token=                GitHub token
  -m, --label-mappings=       Label mappings file on github (default: .github/labeler.yml)
  -M, --label-mappings-local= Label mappings file on the local system
  -d, --dry-run               Dry run, labels won't be applied, only reported

Help Options:
  -h, --help                  Show this help message
```

## Dry-run mode

Labeler has `dry-run` mode. In this mode **it doesnt add any labels**.

Reported information:

- `no match` means all changed files dont match any pattern.
- `has all` means that pull request has all expected labels.
- `list of labels` means that pull request has no expected labels. List of labels to add.

Example output:

```console
DEBU[0003] found 38 open pull requests
DEBU[0003] PR netdata/netdata#6447                       labels="has all"
DEBU[0004] PR netdata/netdata#6390                       labels="has all"
DEBU[0004] PR netdata/netdata#7271                       labels="no match"
DEBU[0004] PR netdata/netdata#6145                       labels="has all"
DEBU[0007] PR netdata/netdata#7945                       labels="has all"
DEBU[0008] PR netdata/netdata#7216                       labels="has all"
DEBU[0008] PR netdata/netdata#7943                       labels="has all"
DEBU[0008] PR netdata/netdata#7185 [dry run]             labels="[area/packaging area/collectors area/external/python area/docs]"
DEBU[0009] PR netdata/netdata#6988                       labels="has all"
DEBU[0009] PR netdata/netdata#7902                       labels="has all"
DEBU[0009] PR netdata/netdata#7962 [dry run]             labels="[area/external area/packaging area/collectors area/docs]"
DEBU[0009] PR netdata/netdata#7377 [dry run]             labels="[area/collectors area/packaging area/external/python area/docs area/web]"
DEBU[0010] PR netdata/netdata#7423                       labels="has all"
DEBU[0010] PR netdata/netdata#7754                       labels="has all"
DEBU[0010] PR netdata/netdata#7942 [dry run]             labels="[area/packaging area/docs]"
DEBU[0011] PR netdata/netdata#7951                       labels="has all"
DEBU[0011] PR netdata/netdata#7988                       labels="has all"
DEBU[0011] PR netdata/netdata#6936                       labels="has all"
DEBU[0011] PR netdata/netdata#7104                       labels="has all"
DEBU[0012] PR netdata/netdata#7769                       labels="has all"
DEBU[0012] PR netdata/netdata#7979 [dry run]             labels="[area/packaging area/collectors area/docs area/web]"
DEBU[0012] PR netdata/netdata#8025                       labels="has all"
```
