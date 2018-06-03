package dreck

import (
	"fmt"
	"strings"

	"github.com/miekg/dreck/types"
)

func hasWIPPrefix(s string) bool {
	for _, w := range wip {
		w = strings.ToLower(w)
		if strings.HasPrefix(s, w) {
			return true
		}
	}
	return false
}

func (d Dreck) pullRequestTitle(req types.PullRequestOuter) (string, error) {
	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return "", err
	}

	pull, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number)
	if err != nil {
		return "", fmt.Errorf("getting PR %d: %s", req.PullRequest.Number, err.Error())
	}
	return pull.GetTitle(), nil
}

var wip = []string{"WIP", "[WIP]", "WIP:", "[WIP]:"}
