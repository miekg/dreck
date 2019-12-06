# Commands

*Dreck* understands the following commands.

| Command | Example | Description | Who Can Use | Pull Request Only |
| --- | --- | --- | --- | --- |
| /[un]label **LABEL** | `/label build` | Adds a label | codeowners | |
| /[un]assign **USER** | `/assign bob` | Assigns to **USER** | codeowners | |
| /[un]cc **USER** | `/cc bob` | Assign to **USER** | anyone | |
| /[un]cc **USER** | `/cc bob` | Request review from **USER** | codeowners |Yes |
| /title **TITLE** | `title New Title` | Sets the title for the issue | codeowners | |
| /[un]lock | `/lock` | Locks or unlocks the issue | codeowners | |
| /duplicate **NUMBER** | `/duplicate 17` | Marks issues as duplicate | anyone | |
| /[un]lgtm | `/lgtm` | Approves the pull request | code owners | Yes |
| /[un]approve | `/approve` | Approves the pull request | code owners | Yes |
| /merge | `/merge` | When status is green and approved, submits pull request | code owners| Yes |
| /retest | `/retest` | Run the checks again | code owners| Yes |
| /exec | `/exec` | Execute a command | code owners| |
| /close | `/close` | Closes the issue | anyone | |
| /reopen | `/reopen` | Opens the issue | anyone | |
| /fortune | `/fortune` | Adds comment containing a fortune (cookie) |anyone | |
| /[un]block **USER** | `/block bob` | Block **USER** | codeowners | |

Extra commands may be defined via aliases, but this depends on the configuration in `.dreck.yaml`.
