# Commands

*Dreck* understands the following commands.

| Command | Example | Description | Who Can Use | Pull Request Only | Feedback
| --- | --- | --- | --- | --- | --- |
| /[un]label **LABEL** | `/label build` | Adds a label | codeowners | | |
| /[un]assign **USER** | `/assign bob` | Assigns to **USER** | codeowners | | |
| /[un]cc **USER** | `/cc bob` | Assign to **USER** | anyone | | |
| /[un]cc **USER** | `/cc bob` | Request review from **USER** | codeowners |Yes | |
| /title **TITLE** | `title New Title` | Sets issue title | codeowners | | |
| /[un]lock **COMMENT** | `/lock` | Locks issue | codeowners | | Uses **COMMENT** as the last comment before locking
| /duplicate **#NUMBER** | `/duplicate #17` | Marks issues as duplicate | anyone | Adds comment and then closes the current issue|
| /[un]lgtm | `/lgtm` | Approves the pull request | code owners | Yes |
| /[un]approve | `/approve` | Approves the pull request | code owners | Yes |
| /merge | `/merge` | When status is green and approved, submits pull request | code owners| Yes | Adds comment with status before merge
| /exec | `/exec` | Execute a command | code owners| | Failure or success is put in a comment
| /close | `/close` | Closes the issue | anyone | |
| /reopen | `/reopen` | Opens the issue | anyone | |
| /fortune | `/fortune` | Adds comment containing a fortune (cookie) |anyone | | Adds comment
| /[un]block **USER** | `/block bob` | Block **USER** | codeowners | | Adds comment that user is blocked

Extra commands may be defined via aliases, but this depends on the configuration in `.dreck.yaml`.
