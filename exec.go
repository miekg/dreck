package dreck

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/github"
)

// sanitize checks the exec command s to see if a respects our white list.
// It is also check for a maximum length of 64, allow what isExec matches, but disallow ..
func sanitize(s string) bool {
	if len(s) > 64 {
		return false
	}
	ok := isExec(s)
	if !ok {
		return false
	}

	// Extra check for .. because the regexp doesn't catch that.
	if strings.Contains(s, "..") {
		return false
	}

	return true
}

func (d Dreck) exec(req types.IssueCommentOuter, conf *types.DreckConfig, cmdType, cmdValue string) error {
	// Due to $reasons cmdValue may be prefixed with spaces and a :, strip those off, cmdValue should
	// then start with a slash.
	pos := strings.Index(cmdValue, "/")
	if pos < 0 {
		return fmt.Errorf("illegal exec command %s", cmdValue)
	}
	run := cmdValue[pos:]

	log.Infof("%s wants to execute %s for #%d\n", req.Comment.User.Login, run, req.Issue.Number)

	parts := strings.Fields(run) // simple split
	if len(parts) == 0 {
		return fmt.Errorf("illegal exec command %s", run)
	}

	// Ok so run needs to come about from an expanded alias, that means it must be a prefix from one of those.
	ok := false
	for _, a := range conf.Aliases {
		r, err := NewAlias(a)
		if err != nil {
			log.Warningf("Failed to parse alias: %s, %v", a, err)
			continue
		}
		if strings.HasPrefix(r.replace, Trigger+execConst+": "+parts[0]) {
			log.Infof("Executing %s, because it is defined in alias expansion %s", run, r.replace)
			ok = true
			break
		}
	}

	if !ok {
		return fmt.Errorf("The command %s is not defined in any alias", run)
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	typ := "pull"
	pull, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
	// 404 error when not found
	if err != nil {
		typ = "issue"
	}

	// Add pull:<NUM> or issue:<NUM> as the first arg.
	arg := fmt.Sprintf("%s/%d", typ, req.Issue.Number)

	log.Infof("About to execute '%s %s %s' for #%d\n", parts[0], arg, strings.Join(parts[1:], " "), req.Issue.Number)
	cmd := exec.Command(parts[0], append([]string{arg}, parts[1:]...)...)

	if typ == "pull" {
		stat := newStatus(statusPending, "In progess", cmd)
		client.Repositories.CreateStatus(ctx, req.Repository.Owner.Login, req.Repository.Name, pull.Head.GetSHA(), stat)
	}

	// Get stdout, errors will go to Caddy log.
	buf, err := cmd.Output()
	if err != nil {
		if typ == "pull" {
			stat := newStatus(statusFail, fmt.Sprintf("Failed: %s", err), cmd)
			client.Repositories.CreateStatus(ctx, req.Repository.Owner.Login, req.Repository.Name, pull.Head.GetSHA(), stat)
		}
		return err
	}

	body := fmt.Sprintf("The command `%s` ran successfully. Its standard output is", run)
	body += "\n~~~\n" + string(buf) + "\n~~~\n"

	comment := githubIssueComment(body)
	client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, comment)

	if typ == "pull" {
		stat := newStatus(statusOK, "Successful", cmd)
		client.Repositories.CreateStatus(ctx, req.Repository.Owner.Login, req.Repository.Name, pull.Head.GetSHA(), stat)
	}

	return nil
}

func newStatus(s, desc string, cmd *exec.Cmd) *github.RepoStatus {
	context := fmt.Sprintf("/exec %s", strings.Join(cmd.Args, " "))

	return &github.RepoStatus{State: &s, Description: &desc, Context: &context}
}

// isExec checks our whitelist.
var isExec = regexp.MustCompile(`^[-a-zA-Z0-9 ./]+$`).MatchString
