package dreck

import (
	"fmt"
	"strings"

	"github.com/miekg/dreck/types"
)

func (d Dreck) isAutosubmit(req types.PullRequestOuter, conf *types.DreckConfig) (bool, error) {
	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return false, err
	}

	pull, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number)
	if err != nil {
		return false, fmt.Errorf("getting PR %d: %s", req.PullRequest.Number, err.Error())
	}

	println("checking for autosmit", pull.User.GetName())

	permitted := permittedUserFeature(featureAutosubmit, conf, pull.User.GetName())
	if !permitted {
		return false, nil
	}

	println("checking autosubmit", pull.GetBody())

	return isautosubmit(pull.GetBody()), nil
}

func isautosubmit(msg string) bool { return strings.Contains(msg, Trigger+autosubmitConst) }

// PullRequestAutosubmit will kick off autosubmit, by calling d.autosubmit.
func (d Dreck) pullRequestAutosubmit(req types.PullRequestOuter) error {
	reqComment := pullRequestOuterToIssueCommentOuter(req)

	return d.autosubmit(reqComment)
}

// pullRequestOuterToIssueCommentOuter converts one type to another. This is not a full copy, but copies
// enough elements to make d.autosubmit work from a pull request.
func pullRequestOuterToIssueCommentOuter(pr types.PullRequestOuter) types.IssueCommentOuter {
	ico := types.IssueCommentOuter{}
	ico.Repository = pr.Repository
	ico.Issue.Number = pr.PullRequest.Number
	ico.InstallationRequest = pr.InstallationRequest

	return ico
}
