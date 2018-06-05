package dreck

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"
)

// sanitize checks the run command s to see if a respects our white list.
// It is also check for a maximum length of 64, allow what isRun matches, but disallow ..
func sanitize(s string) bool {
	if len(s) > 64 {
		return false
	}
	ok := isRun(s)
	if !ok {
		return false
	}

	// Extra check for .. because the regexp doesn't catch that.
	if strings.Contains("..", s) {
		return false
	}

	return true
}

func (d Dreck) run(req types.IssueCommentOuter, conf *types.DreckConfig, cmdType, cmdValue string) error {

	// Due to $reasons cmdValue may be prefixed with spaces and a :, strip those off, cmdValue should
	// then start with a slash.
	pos := strings.Index(cmdValue, "/")
	if pos < 0 {
		return fmt.Errorf("illegal run command %s", cmdValue)
	}
	run := cmdValue[pos:]

	log.Infof("%s wants to run %s for issue #%d\n", req.Comment.User.Login, run, req.Issue.Number)

	parts := strings.Fields(run) // simple split
	if len(parts) == 0 {
		return fmt.Errorf("illegal run command %s", run)
	}

	// Ok so run needs to come about from an expanded alias, that means it must be a prefix from one of those.
	ok := false
	for _, a := range conf.Aliases {
		r, err := NewAlias(a)
		if err != nil {
			log.Warningf("Failed to parse alias: %s, %v", a, err)
			continue
		}
		if strings.HasPrefix(r.replace, Trigger+runConst+": "+parts[0]) {
			log.Infof("Running %s, because it is defined in alias expansion %s", run, r.replace)
			ok = true
			break
		}
	}

	if !ok {
		return fmt.Errorf("The command %s is not defined in any alias", run)
	}

	cmd := exec.Command(parts[0], parts[1:]...)

	// Get stdout, errors will go to Caddy log.
	buf, err := cmd.Output()
	if err != nil {
		return err
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	body := fmt.Sprintf("The command `%s` ran successfully. Its standard output is", run)
	body += "\n~~~\n" + string(buf) + "\n~~~\n"

	comment := githubIssueComment(body)
	client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, comment)

	return nil
}

// isRun checks our whitelist.
var isRun = regexp.MustCompile(`^[-a-zA-Z0-9 ./]+$`).MatchString
