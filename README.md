# Reviewpad 

[![x-ray-badge](https://img.shields.io/badge/Time%20to%20Merge-Fair%20team-bb3e03?link=https://xray.reviewpad.com/analysis?repository=https%3A%2F%2Fgithub.com%2Freviewpad%2Freviewpad&style=plastic.svg)](https://xray.reviewpad.com/analysis?repository=https%3A%2F%2Fgithub.com%2Freviewpad%2Freviewpad) [![Build](https://github.com/reviewpad/reviewpad/actions/workflows/build.yml/badge.svg)](https://github.com/reviewpad/reviewpad/actions/workflows/build.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/reviewpad/reviewpad)](https://goreportcard.com/report/github.com/reviewpad/reviewpad)

## Welcome to Reviewpad!

For **questions**, check out the [discussions](https://github.com/reviewpad/reviewpad/discussions).

For **documentation**, check out this document and the [official documentation](https://docs.reviewpad.com).

**If you think Reviewpad is or could be useful for you, come say hi on [Discord](https://reviewpad.com/discord).**

To start using Reviewpad, check out the [Reviewpad GitHub Action](https://github.com/reviewpad/action).

## What is Reviewpad?
Reviewpad is an open-source tool to automate pull request workflows.

The workflows are specified in a YML-based configuration language described in the [official documentation](https://docs.reviewpad.com/docs/intro-to-reviewpad).

In Reviewpad, you can automate actions over the pull requests such as:
1. Automated [comments](https://docs.reviewpad.com/docs/comment-on-pull-requests);
2. Add or remove [labels](https://docs.reviewpad.com/docs/;automatic-pull-request-labelling);
3. Specify [reviewer assignments](https://docs.reviewpad.com/docs/automatic-reviewer-assignment);
4. Automate [close/merge actions](https://docs.reviewpad.com/docs/auto-merge).

As an example, the following workflow:

```yml
api-version: reviewpad.com/v2.x

labels:
  ship:
    description: Ship mode
    color: 76dbbe

rules:
  - name: changesToMDFiles
    kind: patch
    description: Patch only contains changes to files with extension .md
    spec: $hasFileExtensions([".md"])

workflow:
  - name: ship
    description: Ship process - bypass the review and merge with rebase
    if:
      - rule: changesToMDFiles
    then:
      - $addLabel("ship")
      - $merge()
```

Automatically adds a label `ship` and merges pull requests that only change `.md` files.

You can execute Reviewpad through the CLI or through the Reviewpad [GitHub action](https://github.com/reviewpad/action).

## Architecture
This repository generates two artifacts:

1. CLI [cmd/cli](cmd/cli/main.go) that runs reviewpad open source edition.
2. Reviewpad library packages:  
    ├ github.com/reviewpad/reviewpad/collector  
    ├ github.com/reviewpad/reviewpad/engine  
    ├ github.com/reviewpad/reviewpad/lang/aladino  
    ├ github.com/reviewpad/reviewpad/plugins/aladino  
    ├ github.com/reviewpad/reviewpad/utils/fmtio

Conceptually, the packages are divided into four categories:

1. Engine: The engine is the package responsible of processing the YML file. This process is divided into two stages:
    - Process the YML file to determine which workflows are enabled. The outcome of this phase is a program with the actions that will be executed over the pull request.
    - Execution of the synthesised program.
2. [Aladino Language](https://docs.reviewpad.com/docs/aladino): This is the language that is used in the `spec` property of the rules and also the actions of the workflows. The engine of Reviewpad is not specific to Aladino - this means that it is possible to add the support for a new language such as `Javascript` or `Golang` in these specifications.
3. Plugins: The plugin package contains the built-in functions and actions that act as an abstraction to the 3rd party services such as GitHub, Jira, etc. This package is specific to each supported specification language. In the case of `plugins/aladino`, it contains the implementations of the [built-ins](https://docs.reviewpad.com/docs/aladino-builtins) of the team edition.
4. Utilities: packages such as the collector and the fmtio that provide utilities that are used in multiple places.

## Development

### Setup

This project uses the Goyacc, a version of yacc for Go, used to create the Aladino parser, to install goyacc run:

```sh
go install golang.org/x/tools/cmd/goyacc@master
```

### Compilation

We use [Taskfile](https://taskfile.dev/). To compile the packages simply run:

```sh
task build
```

To generate the CLI of the team edition run:

```sh
task build-cmd
```

This command generate the Reviewpad CLI `main` which you can run to try Reviewpad directly. The CLI has the following argument list:

```sh
./main --help
Usage of ./main:
  -dry-run bool
        Dry run mode
  -event-payload string (optional)
        File path to github action event in JSON format
  -github-token string
        GitHub token
  -mixpanel-token string (optional)
        Mixpanel token
  -pull-request string
        Pull request GitHub url
  -reviewpad string
        File path to reviewpad.yml
```

### VSCode

We strongly recommend the use of VSCode but feel free to use the IDE of your choice. For the case of VSCode we also recommend the installation of the following extensions:

* [EditorConfig](https://marketplace.visualstudio.com/items?itemName=EditorConfig.EditorConfig) - To helps maintaining consistent coding styles.
* [licenser](https://marketplace.visualstudio.com/items?itemName=ymotongpoo.licenser) - For adding license headers.
* [YAML](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) - For enabling `reviewpad.yml` JSON schema.

On your VSCode settings please add the following settings:

```json
{
    // Format files on save
    "editor.formatOnSave": true,
    // Licenser configuration
    "licenser.license": "Custom",
    "licenser.author": "Explore.dev, Unipessoal Lda",
    "licenser.customHeader": "Copyright (C) @YEAR@ @AUTHOR@ - All Rights Reserved\nUse of this source code is governed by a license that can be\nfound in the LICENSE file.",
    // reviewpad.yml JSON schema
    "yaml.schemas": {
        "https://raw.githubusercontent.com/reviewpad/schemas/main/latest/schema.json": [
            "reviewpad.yml",
        ]
    },
    // File nesting
    "explorer.fileNesting.patterns": {
    // Next go and go test files
    "*.go": "${capture}_test.go",
    },
}
```

### Debugging on VSCode

Add the following to your `.vscode/launch.json`.

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch CLI",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "args": [
                // Flag to run on dry run
                "-dry-run",
                // Absolute path to reviewpad.yml file to run
                "-reviewpad=_PATH_TO_REVIEWPAD_FILE_",
                // GitHub url to run the reviewpad.yml against to
                // e.g. https://github.com/reviewpad/action-demo/pull/1
                "-pull-request=_PULL_REQUEST_URL_",
                // GiHub personal access token
                // https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
                "-github-token=_GIT_HUB_TOKEN_",
                // Absolute path to JSON file with GitHub event payload
                "-event-payload=_PATH_TO_EVENT_JSON",
            ],
            "program": "${workspaceFolder}/cmd/cli/main.go"
        }
    ]
}
```

### Contributing

We welcome contributions to Reviewpad from the community!

See the [Contributing Guide](CONTRIBUTING.md).

If you need any assistance, please join [discord](https://reviewpad.com/discord) to reach the core contributors.

**Take a look at the [X-Ray for Reviewpad](https://xray.reviewpad.com/analysis?repository=https%3A%2F%2Fgithub.com%2Freviewpad%2Freviewpad) to see how we are doing!**

## License

Reviewpad is available under the GNU Lesser General Public License v3.0 license.

See LICENSE for the full license text.
