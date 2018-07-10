package dreck

import (
	"os/exec"
	"regexp"

	"github.com/miekg/dreck/types"
)

var r = regexp.MustCompile("^")

// Fortune points to the fortune executable. This is the path on Debian.
var Fortune = "/usr/games/fortune"

func runFortune() (string, error) {
	cmd := exec.Command(Fortune)
	buf, err := cmd.Output()
	if err != nil {
		return "", err
	}

	buf = r.ReplaceAll(buf, []byte("> "))

	return string(buf), nil
}

func (d Dreck) fortune(req types.IssueCommentOuter, cmdType string) error {
	body, err := runFortune()
	if err != nil {
		return err
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	comment := githubIssueComment(body)
	_, resp, err := client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, comment)

	logRateLimit(resp)

	return nil
}
