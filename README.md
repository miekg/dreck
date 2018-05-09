# Dreck

[![Build Status](https://travis-ci.org/miekg/dreck.svg?branch=master)](https://travis-ci.org/miekg/dreck)

It's dreck. Nice to meet you. I'd like to help you with Pull Requests and Issues on your GitHub project.

Dreck is a fork of [Derek](https:/github.com/alexellis/derek). It adds Caddy integration, so you can
just run it as a plugin in Caddy.

For this all to work, you'll need to have an Github App that allows access to your repo - setting
this up is beyond scope of this documentation.

> Please show support for the project and **Star** the repo.

## Config in caddy

~~~
dreck {
    client_id ID
    private_key PATH
    owners NAME
    secret SECRET
    path PAT
    validate
}
~~~

* `client_id` is mandatory and must be the client **ID** of the Github App.
* `private_key` specifies the **PATH** of the private key of the Github App. This is mandatory.
* `secret` can optionally specify a **SECRET** for the webhook.
* `owners` can optionally specify an OWNERS file that is named differently, defaults to "OWNERS".
* `path` will trigger Dreck when the webhook hits **PATH**, defaults to "/dreck".
* `validate` will enable HMAC validation of the request.

## OWNERS File Syntax

```
approvers:
- name1
- name2

reviewers:
- name3
- name4

features:
- feature1
- feature2
```

### Features

* `comments` - allow commands (see below) in comments.
* `dco` - check if a PR has "Signed-off-by" (that literal string) and if not ask for it to be done. Needs a "no-dco" label
  in the repository.
* `reviewers` - assign reviewers for the PR based on changed files and OWNERS' reviewers.

When emailing command the email must start with the command, i.e. `/label rm: bug` and include no
lines above that.

> Note that the assign/unassign commands provides the shortcut `me` to assign to the commenter

## Supported Commands

### Comments

~~~
		/"label: ":        addLabelConst,
		/"label add: ":    addLabelConst,
		/"label remove: ": removeLabelConst,
		/"label rm: ":     removeLabelConst,
		/"assign: ":       assignConst,
		/"unassign: ":     unassignConst,
		/"close":          closeConst,
		/"reopen":         reopenConst,
		/"title: ":        setTitleConst,
		/"title set: ":    setTitleConst,
		/"title edit: ":   setTitleConst,
		/"lock":           lockConst,
        /"unlock":         unlockConst,
~~~

### Pull Requests

For pull requests all modified, addded and removed files are checked. We crawl the path upwards
until we find an OWNERS file. We will then randomly assign someone from the reviewers to review the
PR.

### Examples:

* Update the title of a PR or issue

Let's say a user raised an issue with the title `I can't get it to work on my computer`

```
/title set: Question - does this work on Windows 10?
```
or
```
/title edit: Question - does this work on Windows 10?
```

* Triage and organise work through labels

Labels can be used to triage work or help sort it.

```
/label: proposal
/label add: help wanted
/label remove: bug
/label rm: bug
```

* Assign work

You can assign work to people too

```
/assign: alexellis
/unassign: me
```

* Open and close issues and PRs

Sometimes you may want to close or re-open issues or Pull Requests:

```
/close
/reopen
```

* Lock/un-lock conversation/threads

This is useful for when conversations are going off topic or an old thread receives a lot of comments that are better placed in a new issue.

```
/lock
/unlock
```

# Bugs

Documentation can be much better. Code can also be refactored quite a bit.
