# Contribution Guideline

The go-nebulas project welcomes all contributors. The process of contributing to the Go project may be different than many projects you are used to. This document is intended as a guide to help you through the contribution process. This guide assumes you have a basic understanding of Git and Go.

## Becoming a contributor

Before you can contribute to the go-nebulas project you need to setup a few prerequisites.

### Contributor License Agreement

TBD.

Reference:
 - https://github.com/cla-assistant/cla-assistant
 - https://golang.org/doc/contribute.html


## Preparing a Development Environment for Contributing

### Setting up dependent tools

#### 1. Go dependency management tool

[dep](https://github.com/golang/dep) is an (not-yet) official dependency management tool for Go. go-nebulas project use it to management all dependencies.

For more information, please visit https://github.com/golang/dep

#### 2. Linter for Go source code

[Golint](https://github.com/golang/lint) is official linter for Go source code. Every Go source file in go-nebulas must be satisfied the style guideline. The mechanically checkable items in style guideline are listed in [Effective Go](https://golang.org/doc/effective_go.html) and the [CodeReviewComments wiki page](https://golang.org/wiki/CodeReviewComments).

For more information about Goline, please visit https://github.com/golang/lint.

#### 3. XUnit output for Go Test

[Go2xunit](https://github.com/tebeka/go2xunit) could convert go test output to XUnit compatible XML output used in Jenkins/Hudson.

## Making a Contribution

### Discuss your design

The project welcomes submissions but please let everyone know what you're working on if you want to change or add to the go-nebulas project.

Before undertaking to write something new for the go-nebulas, please [file an issue](https://github.com/nebulasio/go-nebulas/issues/new) (or claim an [existing issue](https://github.com/nebulasio/go-nebulas/issues)). Significant changes must go through the [change proposal process](https://github.com/nebulasio/wiki/change_proposal_process.md) before they can be accepted.

This process gives everyone a chance to validate the design, helps prevent duplication of effort, and ensures that the idea fits inside the goals for the language and tools. It also checks that the design is sound before code is written; the code review tool is not the place for high-level discussions.

Besides that, you can have an instant discussion with core developers in **developers** channel of [Nebulas.IO on Slack](https://nebulasio.herokuapp.com).

### Making a change

#### Getting Go Source

First you need to fork and have a local copy of the source checked out from the forked repository.

You should checkout the go-nebulas source repo inside your $GOPATH. Go to $GOPATH run the following command in a terminal.

```
$ mkdir -p src/github.com/nebulasio
$ cd src/github.com/nebulasio
$ git clone git@github.com:{your_github_id}/go-nebulas.git
$ cd go-nebulas
```

#### Contributing to the main repo

Most Go installations project use a release branch, but new changes should only be made based on the **develop** branch.
(They may be applied later to a release branch as part of the [release process](https://github.com/nebulasio/wiki/release_process.md), but most contributors won't do this themselves.) Before making a change, make sure you start on the **develop** branch:

```
$ git checkout develop
$ git pull
```

### Make your changes

The entire checked-out tree is editable. Make your changes as you see fit ensuring that you create appropriate tests along with your changes. Test your changes as you go.

#### Copyright

Files in the go-nebulas repository don't list author names, both to avoid clutter and to avoid having to keep the lists up to date. Instead, your name will appear in the change log and in the CONTRIBUTORS file and perhaps the AUTHORS file. These files are automatically generated from the commit logs perodically. The AUTHORS file defines who “The go-nebulas Authors”—the copyright holders—are.

New files that you contribute should use the standard copyright header:

```
// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//
```

Files in the repository are copyright the year they are added. Do not update the copyright year on files that you change.

### Gofmt, Golint and Govet

Every Go source file in go-nebulas must pass Gofmt, Golint and Govet check. Golint check the style mistakes, we should fix all style mistakes, including comments/docs. Govet reports suspicious constructs, we should fix all issues as well.

Run following command to check your code:

```
$ make fmt lint vet
```
**lint.report** text file is the Golint report, **vet.report** text file is the Govet report.


### Testing

You've written [test code](https://golang.org/pkg/testing/), tested your code before sending code out for review, run all the tests for the whole tree to make sure the changes don't break other packages or programs:

```
$ make test
```

**test.report** text file or **test.report.xml** XML file is the testing report.


### Commit your changes

The most importance of commiting changes is the commit message. Git will open an editor for a commit message. The file will look like:

```
# Please enter the commit message for your changes. Lines starting
# with '#' will be ignored, and an empty message aborts the commit.
# On branch foo
# Changes not staged for commit:
#	modified:   editedfile.go
#
```
At the beginning of this file is a blank line; replace it with a thorough description of your change. The first line of the change description is conventionally a one-line summary of the change, prefixed by the primary affected package, and is used as the subject for code review email. It should complete the sentence "This change modifies Go to _____." The rest of the description elaborates and should provide context for the change and explain what it does. Write in complete sentences with correct punctuation, just like for your comments in Go. If there is a helpful reference, mention it here. If you've fixed an issue, reference it by number with a # before it.

After editing, the template might now read:

```
math: improve Sin, Cos and Tan precision for very large arguments

The existing implementation has poor numerical properties for
large arguments, so use the McGillicutty algorithm to improve
accuracy above 1e10.

The algorithm is described at http://wikipedia.org/wiki/McGillicutty_Algorithm

Fixes #159

# Please enter the commit message for your changes. Lines starting
# with '#' will be ignored, and an empty message aborts the commit.
# On branch foo
# Changes not staged for commit:
#	modified:   editedfile.go
#
```

The commented section of the file lists all the modified files in your client. It is best to keep unrelated changes in different commits, so if you see a file listed that should not be included, abort the command and move that file to a different branch.

The special notation "Fixes #159" associates the change with issue 159 in the [go-nebulas issue tracker](https://github.com/nebulasio/go-nebulas/issues/159). When this change is eventually applied, the issue tracker will automatically mark the issue as fixed. (There are several such conventions, described in detail in the [GitHub Issue Tracker documentation](https://help.github.com/articles/closing-issues-via-commit-messages/).)

### Creating a Pull Request

For more information about creating a pull request, please refer to the [Create a Pull Request in Github](https://help.github.com/articles/creating-a-pull-request/) page.
