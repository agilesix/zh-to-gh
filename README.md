# ZenHub to GitHub migrator

A CLI tool to migrate ZenHub issue hierarchies to GitHub sub-issues.

## Quickstart

### Prerequisites

- Golang, version 1.22.3
- A GitHub [Personal Access Token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens) with the following permissions:
  - `repo`
- Belonging to a GitHub organization that has access to the sub-issues private beta

### Installation

1. Clone the repo
2. Change directory into the repo: `cd zh-to-gh`
3. Set your GitHub access token as an environment variable: `export GITHUB_TOKEN=...`
4. Build the target: `go build -o target/zhtogh ./cmd`
5. Add the target to your path: `export $PATH:./target`

### Usage

Once you have a JSON file that maps epic-level issues to the corresponding child issues in ZenHub (see [examples/input.json](examples/input.json) for reference) then you can pass the path to that file to the config flag:

```
zhtogh -config examples/input.json
```

And it should print something like this:

```
- Adding sub-issues for: https://github.com/agilesix/simpler-grants-sandbox/issues/30
- Adding sub-issues for: https://github.com/agilesix/simpler-grants-sandbox/issues/29


### Results for parent issue: https://github.com/agilesix/simpler-grants-sandbox/issues/29

Added:	2
-  https://github.com/agilesix/simpler-grants-sandbox/issues/15
-  https://github.com/agilesix/simpler-grants-sandbox/issues/33
Failed:	0


### Results for parent issue: https://github.com/agilesix/simpler-grants-sandbox/issues/30

Added:	2
-  https://github.com/agilesix/simpler-grants-sandbox/issues/13
-  https://github.com/agilesix/simpler-grants-sandbox/issues/12
Failed:	0
```

> [!NOTE]
> You must have write permissions on the repo the issues belong to 
> and the organization that owns the repo must be participating in the 
> [sub-issues beta](https://github.com/orgs/community/discussions/131957)

## TODO

In the future, I'd like to also add entry points to:

- Export an existing list of epics from ZenHub
- Transform and filter the list of epics from ZenHub into the format described above
