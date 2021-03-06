package dreck

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/google/go-github/v28/github"
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
	return "Cookie:\n\n" + string(buf), nil
}

func (d Dreck) fortune(ctx context.Context, client *github.Client, req IssueCommentOuter, _ *Action) (*github.Response, error) {
	body, err := runFortune()
	if err != nil {
		return nil, err
	}

	comment := githubIssueComment(body)
	_, resp, err := client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, comment)
	return resp, err
}
