# Dreck

[![Build
Status](https://travis-ci.org/miekg/dreck.svg?branch=master)](https://travis-ci.org/miekg/dreck)

*dreck* is a fork of [Derek](https://github.com/alexellis/derek). It adds Caddy integration,
so you can "just" run it as a plugin in Caddy v1. It also massively expands on the number
of features. *Dreck* depends on the GitHub CODEOWNERS feature and it will check if that
file exist. A separate `.dreck.yaml` contains various things that are not captured in the
[CODEOWNERS](https://help.github.com/en/github/creating-cloning-and-archiving-repositories/about-code-owners)
file. Dreck doesn't support the email address syntax, so be sure to use GitHub usernames here.

*Dreck* can help you with managing pull requests and issues in your GitHub project. Dreck currently
can:

*  Label/close/lock etc. issues and pull requests comments.

*  LGTM a pull request with a comment.

*  Merge a pull request, but only when the checks are OK and with at least 1 LGTM (/merge).

*  Define (shorter) alias for often used commands.

*  Execute (whitelisted) commands on the dreck server.

The commands must be given as the first word(s) on a line, multiple commands (up to 10) are allowed
but we return on the first error encountered. This holds true for comments that are submitted via
email. Opening issues and pull requests is handled differently in Github, but all commands are
supported in there as well.

Commands are detected in a case insensitive manner.

For this all to work, you'll need to have an Github App that allows access to your repository.
You'll need:

* Issues
* Pull Requests
* Issue Comments

And need to compile Caddy with this plugin.

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

*  `client_id` is mandatory and must be the client **ID** of the Github App.

*  `private_key` specifies the **PATH** of the private key of the Github App. This is also
   mandatory.

*  `secret` specify a **SECRET** for the webhook. This is mandatory.

*  `validate` enable HMAC validation of the request, using the secret. This is mandatory.

*  `owners` can optionally specify a **NAME** for the .dreck.yaml file, defaults to ".dreck.yaml".

*  `path` trigger Dreck when the webhook hits **PATH**, defaults to "/dreck".

*  `merge` defines the **STRATEGY** for merging, possible values are `merge`, `squash` or `rebase`,
   it defaults to `squash`.

*  `user` specifies the **USER** to be used for executing commands. This defaults to the user
   running Caddy.

*  `env` defines environment variable with **NAME** and assign it **VALUE**. These may be repeated.
   Any executed command will have these variables in their environment.

## .dreck.yaml File Syntax

The `.dreck.yaml` file has `features` and `aliases` section that allows you to configure dreck. This
file should live in the top level directory of the repository.

~~~ yaml
features:
    - feature1
    - feature2

aliases:
    - |
      alias1
    - |
      alias2
~~~

An example:

~~~ yaml
features:
    - exec
    - aliases
aliases:
    - |
      /plugin (.*) -> /label add plugin/$1
    - |
      /release (.*) -> /exec /opt/bin/release $1
~~~

### Features

The comment handling feature is always enabled, so you will most like only used this for,
`aliaseses` and `exec`.

The following features are available.

*  `aliases` - enable alias expansion.

*  `exec` - enables `/exec`.

## Supported Commands

The following commands are supported in issue comments and pull requests. When referencing
a user you can use **USER** or **@USER**. Most commands are only available to user referenced in the
CODEOWNERS file. Commands comming from bots (`[bot]` as suffix) are ignored.

*  `/[un]label LABEL`, add/remove a label.

*  `/[un]assign USER`, [un]assign issue to **USER**, the empty string can be used as a
    shortcut for the current user.

*  `/close`, close issue.

*  `/reopen`, reopen issue.

*  `/title TITLE`, short for "title set".

*  `/[un]lock`, [un]lock the issue.

* `/[un]cc USER` [un]cc **USER** request a review from this user, empty string means the current user.

*  `/[un]lgtm`, [un]approve the pull request, this adds a comment that it was LGTM-ed by the user issuing
   this command and adds an approve by the bot.

*  `/[un]approve`, alternative for lgtm.

*  `/merge`, merge this pull request if the checks are green and we have approval (and no explicit
   changes requested). Any pending reviews are deleted.

*  `/exec COMMAND`, executes **COMMAND** on the dreck server. Only commands via an expanded alias
   are allowed.

*  `/duplicate NUMBER`, mark this issue as a duplicate of **NUMBER**. This is done by closing the issue
   and adding the 'duplicate' label.

*  `/fortune`, adds a comment containing text obtained from running "fortune".

## Aliases

The `aliases` sections of the .dreck.yaml allows you to specify alias for other commands. It's
a regular expression based format and looks like this: `alias -> command`. Note the this is:
`<space>-><space>`, e.g.:

    /plugin (.*) -> /label plugin/$1

This defines a new command `/plugin forward` that translates into `/label plugin/forward`.
The regular expression `(.*)` catches the argument after `/plugin` and `$1` is the first expression
match group.

Note this entire string needs to be taken literal to be valid yaml:

~~~ yaml
aliases:
    - |
      /plugin (.*) -> /label plugin/$1
~~~

## Exec

Exec allows for processes be started on the dreck server. For this the `exec` feature *and* the
`aliases` feature must be enabled. Only commands *expanded* by an alias are allowed to execute, this
is to prevent things like `/exec /bin/cat /etc/passwd` to be run accidentally. The standard output
of the command will be picked up and put in the new comment under the issue or pull request.

If `user` is specified dreck will run the command under that user.

Apart from the environment set in the configuration all command well have access to GITHUB\_TRIGGER.
If the command is given in an issue dreck will set GITHUB\_TRIGGER to `issue/NUMBER`, if done for a
pull request that value will be `pull/NUMBER`.

If the command is run for a pull request dreck will update the status with 'pending' when the
execution is in progress and either 'failed' or 'success' when the execution ends.

For example, if you want to execute `/opt/bin/release ARGUMENT` on the server, the following alias
must be defined:

~~~
/release (.*) -> /exec /opt/bin/release $1
~~~

If you then call the command with `/release 0.1` in issue 42. *dreck* will run:

~~~
/opt/bin/release 0.1
~~~

And GITHUB\_TRIGGER will be issue/42.

Note that in this case `/cat -> /exec /bin/cat /etc/resolv.conf`, running `cat /etc/passwd`
*still* yields in an (unwanted?) disclosure because the final command being run is `/bin/cat
/etc/resolv.conf /etc/passwd`. In other words be careful of what commands you white list.

*dreck* enforces a very restrictive white list on the allowed characters in the command. The white
list currently is this regular expression: `^[-a-zA-Z0-9 ./]+$`. Note that two dots in a row is not
allowed.

# Examples

Set a label on an issue, on Github (or via email), create a reply that contains:

~~~
/label bug
~~~

And *dreck* will apply that label if it exists. Text can freely intermixed, but each command should
be on its own line and start on the left most position.

~~~
This is good question.
/label question
~~~

While the following will not be detected as a command:

~~~
This is good question. /label question
~~~

# Also See

See [Derek](https://github.com/alexellis/derek) of which dreck is a fork.

# Adding in Caddy

Apply a patch like the following:

~~~ patch
diff --git caddyhttp/caddyhttp.go caddyhttp/caddyhttp.go
index 7ca8b874..34b8b405 100644
--- caddyhttp/caddyhttp.go
+++ caddyhttp/caddyhttp.go
@@ -46,4 +46,6 @@ import (
        _ "github.com/caddyserver/caddy/caddyhttp/timeouts"
        _ "github.com/caddyserver/caddy/caddyhttp/websocket"
        _ "github.com/caddyserver/caddy/onevent"
+
+       _ "github.com/miekg/dreck"
 )
diff --git caddyhttp/httpserver/plugin.go caddyhttp/httpserver/plugin.go
index 378e9cb8..79498579 100644
--- caddyhttp/httpserver/plugin.go
+++ caddyhttp/httpserver/plugin.go
@@ -701,6 +701,7 @@ var directives = []string{
        "restic",    // github.com/restic/caddy
        "wkd",       // github.com/emersion/caddy-wkd
        "dyndns",    // github.com/linkonoid/caddy-dyndns
+       "dreck",     // "github.com/miekg/dreck"
 }

 const (
~~~
