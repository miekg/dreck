# Dreck

[![Build Status](https://travis-ci.org/miekg/dreck.svg?branch=master)](https://travis-ci.org/miekg/dreck)

Dreck is a fork of [Derek](https://github.com/alexellis/derek). It adds Caddy integration, so you
can "just" run it as a plugin in Caddy. It also massively expands on the number of features.

Dreck can help you with managing pull requests and issues in your GitHub project. Dreck currently
can:

* Label/close/lock etc. issues.
* Assign reviewers to a pull request based on *OWNERS* files, taking into account Work-in-Progress
  status.
* Delete the branch when a pull request is merged.
* Merge a pull request when the status is green (/autosubmit).
* LGTM a pull request with a comment.
* Merge a pull request, but only when the checks are OK and with at least 1 LGTM (/merge).
* Define (shorter) alias for often used commands.
* Execute (whitelisted) commands on the dreck server.

The commands must be given as the first word(s) on a line, multiple commands (up to 10) are allowed
but we return on the first error encountered. This holds true for comments that are submitted via
email.

Commands are detected in a case insensitive manner.

For this all to work, you'll need to have an Github App that allows access to your repository
- setting this up is beyond scope of this documentation. And need to recompile Caddy and have
a functional Go setup; again: all beyond the scope of this document.

## Config in caddy

If configuring Caddy, you need to recompile it with the *dreck* plugin enabled. After that the
following configuration is available.

~~~
dreck {
    client_id ID
    private_key PATH
    owners NAME
    secret SECRET
    path PATH
    merge STRATEGY
    validate
    user USER
    env NAME VALUE
}
~~~

* `client_id` is mandatory and must be the client **ID** of the Github App.
* `private_key` specifies the **PATH** of the private key of the Github App. This is also mandatory.
* `secret` can optionally specify a **SECRET** for the webhook.
* `owners` can optionally specify a **NAME** for the OWNERS files, defaults to "OWNERS".
* `path` trigger Dreck when the webhook hits **PATH**, defaults to "/dreck".
* `merge` defines the **STRATEGY** for merging, possible values are `merge`, `squash` or `rebase`,
  it defaults to `squash`.
* `validate` enable HMAC validation of the request.
* `user` specifies the **USER** to be used for executing commands. This defaults to the user running
  Caddy.
* `env` defines environment variable with **NAME** and assign it **VALUE**. These may be repeated.
  Any executed command will have these variables in their environment.

## OWNERS File Syntax

The OWNERS file syntax is borrowed from Kubernetes and extended with a `features` and `aliases`
section that allows you to configure dreck. This file should live in the top level directory of the
repository. Other OWNERS files may exist in deeper directories. These are used to assign reviewers
from for pull requests.

``` yaml
approvers:
    - name1
    - name2

reviewers:
    - name3
    - name4

features:
    - feature1
    - feature2

aliases:
    - |
      alias1
    - |
      alias2
```

An example:

~~~ yaml
approvers:
    - miek
reviewers:
    - miek
features:
    - comments
    - exec
    - aliases
aliases:
    - |
      /plugin: (.*) -> /label add: plugin/$1
    - |
      /release: (.*) -> /exec: /opt/bin/release $1
~~~

### Features

The following features are available.

* `comments` - allow commands (see below) in comments.
* `reviewers` - assign reviewers for the pull request based on changed files and reviewers in the
  relevant OWNERS files.
* `dco` - check if a pull request has "Signed-off-by" (that literal string) and if not ask for it to
  be done. Needs a "no-dco" label in the repository for it to work.
* `aliases` - enable alias expansion.
* `branches` - enables the deletion of branches after a merge of a pull request.
* `autosubmit` - enables `/autosubmit`.
* `merge` - enables `/merge`.
* `exec` - enables `/exec`.

## Supported Commands

### Comments

The following commands are supported in issue comments.

* `/label add: LABEL`, label an issue with **LABEL**.
* `/label: LABEL`,  short for "label add".
* `/label remove: LABEL`, remove **LABEL**.
* `/label rm: LABEL`, short for "label remove",
* `/assign: ASSIGNEE`, assign issue to **ASSIGNEE**, `me` can be used as a shortcut for the
  commenter
* `/unassign: ASSIGNEE`, unassigns **ASSIGNEE**.
* `/close`, close issue.
* `/reopen`, reopen issue.
* `/title set: TITLE`, set the title to **TITLE**.
* `/title: TITLE`: short for "title set".
* `/title edit: TITLE`, set the title to **TITLE**.
* `/lock`, lock the issue.
* `/unlock`, unlock the issue.
* `/exec COMMAND`, executes **COMMAND** on the dreck server. Only commands via an expanded alias are
  allowed.
* `/test`, a noop used for testing dreck.
* `/duplicate: NUMBER`, mark this issue as a duplicate of NUMBER. This is done by closing the issue
  and adding the 'duplicate' label.

### Pull Requests

When a pull request is submitted dreck will check which files are modified, removed or changed. For
a subset of these it will search for the nearest OWNERS file. We will then randomly assign someone
from the reviewers to review the pull request. This is only done when the pull request does not have
any reviewers, nor is a work-in-progress.

If pull requests have `WIP` (case insensitive) as a prefix in the title and this title is changed to
remove that prefix we will search (again) for a reviewer. The prefixes allowed are: `WIP`, `WIP:`,
`[WIP]` and `[WIP]:`.

Further more the following extra commands are supported for pull request issues comments (ignored for
issues).

* `/lgtm`, approve the pull request.
* `/autosubmit`, when all checks are OK, automatically merge the pull request. This will wait for 30
  minutes for all tests to complete. The label 'autosubmit' is added to the pull request. Note that
  the command `/autosubmit` can *also be given in the pull request body*. If we dreck sees this it
  will perform the same checks and, if allowed, we start submitting.
* `/exec`, executing commands is supported for pull requests.
* `/merge`, merge this pull request if the checks are green and we have approval (and no
  explicit changes requested).

## Aliases

The `aliases` sections of the OWNERS file allows you to specify alias for other commands. It's
a regular expression based format and looks like this: `alias -> command`. Note the this is:
`<space>-><space>`, e.g.:

~~~
/plugin: (.*) -> /label add: plugin/$1
~~~

This defines a new command `/plugin: forward` that translates into `/label add: plugin/forward`. The
regular expression `(.*)` catches the argument after `/plugin: ` and `$1` is the first expression
match group.

Note this entire string needs to be taken literal in the OWNERS file to be valid yaml:

~~~ yaml
aliases:
    - |
      /plugin: (.*) -> /label add: plugin/$1
~~~

### Exec

Exec allows for processes be started on the dreck server. For this the `exec` feature *and* the
`aliases` feature must be enabled. Only commands *expanded* by an alias are allowed to execute, this
is to prevent things like `/exec: /bin/cat /etc/passwd` to be run accidentally. The standard output
of the command will be picked up and put in the new comment under the issue or pull request.

If `user` is specified dreck will run the command under that user.

The command executed can not be a script, it must be a real executable.

Apart from the environment set in the configuration all command well have access to GITHUB_TRIGGER.
If the command is given in an issue dreck will set GITHUB_TRIGGER to `issue/NUMBER`, if done for
a pull request that value will be `pull/NUMBER`.

If the command is run for a pull request dreck will update the status with 'pending' when the
execution is in progress and either 'failed' or 'success' when the execution ends.

For example, if you want to execute `/opt/bin/release ARGUMENT` on the server, the following alias
must be defined:

~~~
/release: (.*) -> /exec: /opt/bin/release $1
~~~

If you then call the command with `/release 0.1` in issue 42. Dreck will run:

~~~
/opt/bin/release 0.1
~~~

And GITHUB_TRIGGER will be issue/42.

Note that in this case `/cat -> /exec: /bin/cat /etc/resolv.conf`, running `cat /etc/passwd` *still*
yields in an (unwanted?) disclosure because the final command being run is `/bin/cat
/etc/resolv.conf /etc/passwd`. In other words be careful of what commands you white list.

Dreck enforces a very restrictive white list on the allowed characters in the command. The
white list currently is this regular expression: `^[-a-zA-Z0-9 ./]+$`. Note that two dots in a row
is not allowed.

## Branches

With this enabled, *dreck* will, after each closed pull request, look to see if the branch is
merged, but not deleted. If this is true, it will delete the branch. The *master* branch is always
excluded from this.

# Examples

Set a label on an issue, on Github (or via email), create a reply that contains:

~~~
/label: bug
~~~

And dreck will apply that label if it exists. Text can freely intermixed, but each command should be
on its own line and start on the left most position.

~~~
This is good question.
/label: question
~~~

While the following will not be detected as a command:

~~~
This is good question. /label: question
~~~

# Also See

See [Derek](https://github.com/alexellis/derek) of which dreck is a fork.
