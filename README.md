# Simple CSV parser looking for telemetry data

## Demonstrates use of Golang, Github Actions, Linting

The .golangci.yml file is used to customize and control the behavior of golangci-lint when it analyzes the Go code herein.
The file contains:

Linter Configuration:
Use of this file to enable or disable specific linters. In this configuration, I have disabled all linters by default and then explicitly enabled a set of linters.
Run Settings:

You can specify how golangci-lint should run, including things like concurrency, timeout, and whether to include test files.

Issue Management:
You can configure how issues are reported, including which directories or files to exclude from analysis.

Custom Rules:
You can set custom rules for certain linters. For instance, you've customized rules for the revive linter.

Project-Specific Settings:
You can tailor the linting process to your project's needs. For example, I've allowed the FileParser interface to be returned in the ireturn linter settings.

To use this golangci.yml file:

Place it in the root directory of your Go project.
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

Run golangci-lint from the command line in your project directory:
golangci-lint run  



GITHUB CI/CD Pipeline - Golangci-Lint
Look at the file in .github/workflows golangci-lint.yml

GitHub will automatically run golangci-lint on every push to main/master and on every pull request targeting these branches.

