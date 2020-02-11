# Periodic Pull Requests Labeler

A GitHub action to automatically label all PRs **based on changed files** according to file patterns.

## Usage

Action is meant to be run as periodic job. This is needed to workaround issues regarding
[lack of write access when executed from fork](https://help.github.com/en/actions/automating-your-workflow-with-github-actions/authenticating-with-the-github_token#permissions-for-the-github_token)
which is a common problem when using https://github.com/actions/labeler.

```yaml
---
name: Pull Request Labeler
on:
  schedule:
    - cron: '*/5 * * * *'
jobs:
  labeler:
    runs-on: ubuntu-latest
    steps:
      - uses: docker://docker.pkg.github.com/ilyam8/periodic-pr-labeler/periodic-pr-labeler:master
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GITHUB_REPOSITORY: ${{ github.repository }}
          LABEL_MAPPINGS_FILE: .github/labeler.yml
```

By default action uses `.github/labeler.yml` located in repository from `GITHUB_REPOSITORY` as a source of pattern matchers.
This file uses the same schema as in https://github.com/actions/labeler

## File patterns

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
  -r, --repository=           github repository slug
  -t, --token=                github token
  -m, --label-mappings=       label mappings file on github (default: .github/labeler.yml)
  -M, --label-mappings-local= label mappings file on the local system
  -d, --dry-run               dry run, labels won't be applied, only reported

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
