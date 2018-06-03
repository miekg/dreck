package dreck

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/github"
	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"
)

func (d Dreck) autosubmit(req types.IssueCommentOuter, cmdType string) error {
	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(15 * time.Second)
	stop := time.NewTimer(30 * time.Minute)
	defer ticker.Stop()
	defer stop.Stop()

	log.Infof("Start autosubmit polling for PR %d", req.Issue.Number)

	for {
		select {
		case <-ticker.C:

			pull, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
			if err != nil {
				return err
			}

			d.pullRequestStatus(ctx, client, req, pull)

			if pull.Mergeable != nil {
				continue
				//return d.pullRequestMerge(ctx, client, req, pull)
			}

		case <-stop.C:

			return fmt.Errorf("timeout while waiting for PR %d", req.Issue.Number)
		}
	}
}

func (d Dreck) pullRequestMerge(ctx context.Context, client *github.Client, req types.IssueCommentOuter, pull *github.PullRequest) error {

	opt := &github.PullRequestOptions{MergeMethod: d.strategy}
	msg := "Automatically submitted."
	commit, _, err := client.PullRequests.Merge(ctx, req.Repository.Owner.Login, req.Repository.Name, *pull.Number, msg, opt)

	if err != nil {
		return fmt.Errorf("failed merge of PR %d: %s", *pull.Number, err.Error())
	}

	body := fmt.Sprintf("This pull request has been automatically merged in %s.", commit.GetSHA())

	comment := githubIssueComment(body)
	client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, *pull.Number, comment)

	return nil
}

func (d Dreck) pullRequestStatus(ctx context.Context, client *github.Client, req types.IssueCommentOuter, pull *github.PullRequest) (bool, error) {

	listOpts := &github.ListOptions{PerPage: 100}
	statuses, _, err := client.Repositories.ListStatuses(ctx, req.Repository.Owner.Login, req.Repository.Name, pull.Head.GetSHA(), listOpts)
	if err != nil {
		return false, err
	}

	for _, status := range statuses {
		println(status.GetState())
		println(status.String())
		println(status.GetContext())
	}

	return false, fmt.Errorf("no status found for %d", pull.GetNumber())
}

const statusOK = "ok"
