package dreck

import (
	"fmt"
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
	if len(buf) == 0 {
		return "", fmt.Errorf("no output returned")
	}

	buf = r.ReplaceAll(buf, []byte("> "))
	return "Fortune cookie\n\n" + string(buf), nil
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
	_, _, err = client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, comment)
	return err
}
