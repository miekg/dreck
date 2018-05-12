# Dreck

[![Build Status](https://travis-ci.org/miekg/dreck.svg?branch=master)](https://travis-ci.org/miekg/dreck)

Dreck can help you with Pull Requests and Issues on your GitHub project

Dreck is a fork of [Derek](https:/github.com/alexellis/derek). It adds Caddy integration, so you can
"just" run it as a plugin in Caddy and a bunch of other features.

For this all to work, you'll need to have an Github App that allows access to your repo - setting
this up is beyond scope of this documentation. And need to recompile Caddy and have a functional Go
setup; again: all beyond the scope of this document.

## Config in caddy

If you configuring Caddy, you need to recompile it with the *dreck* plugin enabled. After that the
following configuration is available.

~~~
dreck {
    client_id ID
    private_key PATH
    owners NAME
    secret SECRET
    path PATH
    validate
}
~~~

* `client_id` is mandatory and must be the client **ID** of the Github App.
* `private_key` specifies the **PATH** of the private key of the Github App. This is also mandatory.
* `secret` can optionally specify a **SECRET** for the webhook.
* `owners` can optionally specify a **NAME** for the ONWERS files, defaults to "OWNERS".
* `path` trigger Dreck when the webhook hits **PATH**, defaults to "/dreck".
* `validate` enable HMAC validation of the request.

## OWNERS File Syntax

The OWNERS file syntax is borrowed from Kubernetes and extended with a `features` and `aliases`
section that allows you to configure dreck. This file should live in the top level directory of the
repository.

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
aliases:
    - |
      /plugin: (.*) -> /label add: plugin/$1
~~~

### Features

The following features are available.

* `comments` - allow commands (see below) in comments.
* `reviewers` - assign reviewers for the PR based on changed files and reviewers in the relevant
  OWNERS files.
* `dco` - check if a PR has "Signed-off-by" (that literal string) and if not ask for it to be done.
  Needs a "no-dco" label in the repository for it to work.

When using email to reply to an issue, the email *must* start with the command, i.e. `/label rm: bug`
and include no lines above that.

Multiple command in one message/issue are not supported.

## Supported Commands

### Comments

The following commands are supported.

* `/label add: LABEL`, label an issue with LABEL.
* `/label: LABEL`,  short for "label add".
* `/label remove: LABEL`, remove LABEL.
* `/label rm: LABEL`, short for "label remove",
* `/assign: ASSIGNEE`, assign issue to ASSIGNEE, `me` can be used as a shortcut for the commenter
* `/unassign: ASSIGNEE`, unassign ASSIGNEE.
* `/close`, close issue.
* `/reopen`, reopen issue.
* `/title set: TITLE`, set the title to TITLE.
* `/title: TITLE`: short for "title set".
* `/title edit: TITLE`, set the title to TITLE.
* `/lock`, lock the issue.
* `/unlock`, unlock the issue.

### Pull Requests

For pull requests all modified, addded and removed files are checked. We crawl the path upwards
until we find an OWNERS file. We will then randomly assign someone from the reviewers to review the
PR.

Further more the following command is support for PR issues comments (ignored for issues).

* `/lgtm`, approve the PR.

## Aliases

The `aliases` sections of the OWNERS file allows you to specify alias for other commands. It's
a regular expression based format and looks like this: `alias -> command`. Note the this is:
`<space>-><space>`, e.g.:

~~~
/plugin: (.*) -> /label add: plugin/$1
~~~

This defines a new command `/plugin: forward` that translates into `/label add: plugin/forward`.
The regular expression `(.*)` catches the argument after `/plugin: ` and `$1` is the first expression
match group.

Note this entire string needs to be taken literal in the OWNERS file to be valid yaml:

~~~ yaml
aliases:
    - |
    /plugin: (.*) -> /label add: plugin/$1
~~~ 

## Bugs

We don't support multiple commands in an issue.
