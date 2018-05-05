# derek

[![Build Status](https://travis-ci.org/miekg/dreck.svg?branch=master)](https://travis-ci.org/miekg/dreck)

It's dreck. Nice to meet you. I'd like to help you with Pull Requests and Issues on your GitHub project.

Dreck is a fork of [Derek](https:/github.com/alexellis/derek). It adds Caddy integration, so you can
just run it.

> Please show support for the project and **Star** the repo.

# Config in caddy

dreck {
    owners NAME // owners file
    secret SECRET // webhook secret
    path PATH // when to trigger
    key PATH
}



## How to use

* Build with Caddy
* Add a webhook content/type: application/json

## What can I do?

* Check that commits are signed-off

When someone sends a PR without a sign-off, I'll apply a label `no-dco` and also send them a comment pointing them to the contributor guide. Most of the time when I've been helping the OpenFaaS project - people read my message and fix things up without you having to get involved.

* Allow users in a specified .dreck.yml file to manage issues and pull-requests

You don't have to give people full write access anymore to help you manage issues and pull-requests.
I'll do that for you, just put them in a .dreck.yml file in the root and when they comment on an
issue then I'll use my granular permissions instead.

* Wait.. doesn't the term "maintainer" mean write access in GitHub?

No this is what Derek sets out to resolve. The users in your maintainers list have granular permissions which you'll see in detail when you add the app to your repo org.

```
maintainers:
- alexellis
- rgee0
```

You can use the alias "curators" instead for the exact same behaviour:

```
curators:
- alexellis
- rgee0
```

* What about roles?

We are planning to add roles in the ROADMAP which will mean you can get even more granular and have folks who can only add labels but not close issues for instance. If you feel you need to make that distinction. It will also let you call the roles whatever you think makes sense.

> Note that the assign/unassign commands provides the shortcut `me` to assign to the commenter

### Examples:

* Update the title of a PR or issue

Let's say a user raised an issue with the title `I can't get it to work on my computer`

```
/set title: Question - does this work on Windows 10?
```
or
```
/edit title: Question - does this work on Windows 10?
```

* Triage and organise work through labels

Labels can be used to triage work or help sort it.

```
/add label: proposal
/add label: help wanted
/remove label: bug
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

### Backlog:

* [x] Derek as a managed GitHub App
* [x] Lock thread
* [x] Edit title
* [x] Toggle the DCO-feature

Future work:

* [ ] Caching .dreck.yml file
* [ ] Observability of GitHub API Token rate limit
* [ ] Add roles & actions
* [ ] Branch Checking
