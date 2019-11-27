# Commands

*Dreck* understands the following commands.

| Command | Example | Description | Who Can Use | Pull Request Only |
| --- | --- | --- | --- | --- |
| /[un]label **LABEL** | `/label build` | Adds a label | codeowners | |
| /[un]assign **USER** | `/assign bob` | Assigns or unassign to **USER** | codeowners | |
| /[un]cc **USER** | `/cc bob` | Cc **USERS on the issue | codeowners | |
| /close | `/close` | Closes the issue | anyone | |
| /reopen | `/reopen` | Opens the issue | anyone | |
| /title **TITLE** | `title New Title` | Sets the title for the issue | codeowners | |
| /duplicate **NUMBER** | `/duplicate 17` | Marks issues as duplicate | anyone | |
| /[un]lock | `/lock` | Locks or unlocks the issue | codeowners | |
| /fortune | `/fortune` | Adds comment containing a fortune (cookie) |anyone | |
| /[un]lgtm | `/lgtm` | Approves the pull request | anyone | Yes |
| /merge | `/merge` | When status is green and approved, submits pull request | code owners| Yes |
| /exec | `/exec` | Execute a command | code owners| |
| /fortune | `/fortune` | Add fortune comment | anyone | |

Extra commands may be defined via aliases, but this depends on the configuration.
