We love and appreciate contributions to Go-SIRIS.

### To contribute code, use the Github "Pull Request" mechanism

#### Overview

1. fork, by clicking the Fork button on the [GitHub Project Page](https://github.com/go-siris/siris)
2. checkout a copy
3. create a branch
4. make changes
5. push changes to your fork
6. submit Pull Request

#### Detailed Example

```sh
export GHUSERNAME=CHANGE_THIS
git clone https://github.com/$GHUSERNAME/siris.git
cd siris
git checkout -b new_branch
$EDITOR file_to_change
git add file_to_change
git commit # with a useful message
git push origin new_branch
```

The `git commit` step(s) will launch you into `$EDITOR` where the first line should be a summary of the change(s) in less than 50 characters. Additional paragraphs can be added starting on line 3.

To submit new_branch as a Pull Request, visit the [Go-SIRIS project page](https://github.com/go-siris/siris) where your recently pushed branches will appear with a green "Pull Request" button.

### Rebase

On branches with more than a couple commits, it's usually best to squash the commits (condense them into one) before submitting the change(s) as a PR. Notable exceptions to the single commit guideline are:

* where there are multiple logical changes, put each in a commit (easier to review and revert)
* whitespace changes belong in their own commit
* no-op code refactoring is separate from functional changes

To rebase:

```sh
git remote add siris https://github.com/go-siris/siris.git
git remote update siris
git rebase -i siris/master
```

Change all but the first "pick" lines to "s" and save your changes. Your $EDITOR will then present you with all of the commit messages. Edit them and save. Then force push your branch:

`git push -f`

### General Guidelines

* New features **must** be documented
* New features **should** include tests

### Style conventions

* run go fmt path/to/project

## Tests

* run go test -v -cover path/to/project
