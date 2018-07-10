# Commands

*Dreck* understands the following commands.

| Command | Example | Description | Who Can Use | Pull Request Only |
| --- | --- | --- | --- | --- |
| /label [add]: **LABEL** | `/label: build` | Adds a label | anyone | |
| /label [remove,rm]: **LABEL** | `/label rm: build` | Remove a label | anyone | |
| /[un]assign: **ASSIGNEE** | `/assign: bob` | Assigns or unassign to *ASSIGNEE* | anyone | |
| /close | `/close` | Closes the issue | anyone | |
| /reopen | `/reopen` | Opens the issue | anyone | |
| /title [set,edit]: **TITLE** | `title set: New Title` | Sets the title for the issue | anyone | |
| /duplicate: **NUMBER** | `/deplicate: 17` | Marks issues as duplicate | anyone | |
| /[un]lock | `/lock` | Locks or unlocks the issue | Approvers, Reviewers | |
| /fortune | `/fortune` | Adds comment containing a fortune (cookie) |anyone | |
| /lgtm | `/lgmt` | Approves the pull request |anyone | Yes |
| /autosubmit | `/autosubmit` | When status is green, submits pull request | anyone | Yes |
| /merge | `/merge` | When status is green and approved, submits pull request | Approvers, Reviewers | Yes |
