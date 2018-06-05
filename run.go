package dreck

import (
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

func (d Dreck) run(req types.IssueCommentOuter, cmdType, cmdValue string) error {

	log.Infof("%s wants to run %s for issue #%d\n", req.Comment.User.Login, cmdValue, req.Issue.Number)
	/*
		client, ctx, err := d.newClient(req.Installation.ID)
		if err != nil {
			return err
		}
	*/

	// cmdValue is what is being run.

	// sanitize

	return nil
}

// isRun checks our whitelist.
var isRun = regexp.MustCompile(`^[-a-zA-Z0-9 ./]+$`).MatchString
