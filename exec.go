package dreck

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/v28/github"
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

func (d Dreck) exec(ctx context.Context, client *github.Client, req types.IssueCommentOuter, conf *types.DreckConfig, cmdType, cmdValue string) error {
	// Due to $reasons cmdValue may be prefixed with spaces and a :, strip those off, cmdValue should
	// then start with a slash.
	run, err := stripValue(cmdValue)
	if err != nil {
		return fmt.Errorf("illegal exec command %s", run)
	}

	log.Infof("%s wants to execute %s for #%d", req.Comment.User.Login, run, req.Issue.Number)

	parts := strings.Fields(run) // simple split
	if len(parts) == 0 {
		return fmt.Errorf("illegal exec command %s", run)
	}

	if !isValidExec(conf, parts, run) {
		return fmt.Errorf("The command %s is not defined in any alias", run)
	}

	typ := "pull"
	pull, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
	// 404 error when not found
	if err != nil {
		typ = "issue"
	}

	log.Infof("Assembling command '%s %s' for #%d", parts[0], strings.Join(parts[1:], " "), req.Issue.Number)

	// Add pull:<NUM> or issue:<NUM> as the first arg.
	trigger := fmt.Sprintf("%s/%d", typ, req.Issue.Number)
	cmd, err := d.execCmd(parts, trigger)
	if err != nil {
		return err
	}

	if typ == "pull" {
		stat := newStatus(statusPending, "In progress", cmd)
		client.Repositories.CreateStatus(ctx, req.Repository.Owner.Login, req.Repository.Name, pull.Head.GetSHA(), stat)
	}

	log.Infof("Executing '%s %s' for #%d", parts[0], strings.Join(parts[1:], " "), req.Issue.Number)

	// Get all output
	buf, err := cmd.CombinedOutput()
	if err != nil {
		if typ == "pull" {
			stat := newStatus(statusFail, fmt.Sprintf("Failed: %s", err), cmd)
			client.Repositories.CreateStatus(ctx, req.Repository.Owner.Login, req.Repository.Name, pull.Head.GetSHA(), stat)
		}
		body := fmt.Sprintf("The command `%s` did **not** run **successfully**. The status returned is `%s`\n\n", run, err.Error())
		if len(buf) > 0 {
			body += "Its standard and error output is"
			body += "\n~~~\n" + string(buf) + "\n~~~\n"
		}

		comment := githubIssueComment(body)
		client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, comment)
		return err
	}

	body := fmt.Sprintf("The command `%s` ran **successfully**. Its standard and error output is", run)
	body += "\n~~~\n" + string(buf) + "\n~~~\n"

	comment := githubIssueComment(body)
	client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, comment)

	if typ == "pull" {
		stat := newStatus(statusOK, "Successful", cmd)
		client.Repositories.CreateStatus(ctx, req.Repository.Owner.Login, req.Repository.Name, pull.Head.GetSHA(), stat)
	}

	return nil
}

// execCmd creates an exec.Cmd with the right attributes such as the environment and user to run as.
func (d Dreck) execCmd(parts []string, trigger string) (*exec.Cmd, error) {

	cmd := exec.Command(parts[0], parts[1:]...)
	for i := range parts {
		if !sanitize(parts[i]) {
			return nil, fmt.Errorf("exec: %q doesn't adhere to the sanitize checks", parts[i])
		}
	}

	// run as d.user, if not empty
	if d.user != "" {
		uid, gid, err := userID(d.user)
		if err != nil {
			return nil, err
		}

		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid}
	}

	// extend environment
	env := os.Environ()
	env = append(env, fmt.Sprintf("GITHUB_TRIGGER=%s", trigger))
	for e, v := range d.env {
		env = append(env, fmt.Sprintf("%s=%s", e, v))
	}
	cmd.Env = env

	return cmd, nil
}

func newStatus(s, desc string, cmd *exec.Cmd) *github.RepoStatus {
	context := fmt.Sprintf("/exec %s", strings.Join(cmd.Args, " "))

	return &github.RepoStatus{State: &s, Description: &desc, Context: &context}
}

func stripValue(s string) (string, error) {
	pos := strings.Index(s, "/")
	if pos < 0 {
		return "", fmt.Errorf("illegal exec command %s", s)
	}
	return s[pos:], nil
}

func isValidExec(conf *types.DreckConfig, parts []string, run string) bool {
	// Ok so run needs to come about from an expanded alias, that means it must be a prefix from one of those.
	for _, a := range conf.Aliases {
		r, err := NewAlias(a)
		if err != nil {
			log.Warningf("Failed to parse alias: %s, %v", a, err)
			continue
		}
		if strings.HasPrefix(r.replace, Trigger+execConst+" "+parts[0]) {
			log.Infof("Executing %s, because it is defined in alias expansion %s", run, r.replace)
			return true
		}
	}
	return false
}

// isExec checks our whitelist.
var isExec = regexp.MustCompile(`^[-a-zA-Z0-9 ./]+$`).MatchString
