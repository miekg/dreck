package dreck

import (
	"context"
	"fmt"
	"time"

	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/github"
)

func (d Dreck) autosubmit(req types.IssueCommentOuter) error {
	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(15 * time.Second)
	stop := time.NewTimer(30 * time.Minute)
	defer ticker.Stop()
	defer stop.Stop()

	// Add autosubmit label to signal we will merge this automatically.
	client.Issues.AddLabelsToIssue(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, []string{"autosubmit"})

	log.Infof("Start autosubmit polling for PR %d", req.Issue.Number)

	for {
		select {
		case <-ticker.C:

			pull, resp, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
			if err != nil {
				return err
			}
			fmt.Printf("%v\n", resp) // bail out on 404

			ok, _ := d.pullRequestStatus(ctx, client, req, pull)
			if ok && pull.Mergeable != nil {
				err := d.pullRequestMerge(ctx, client, req, pull)
				return err
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

	log.Infof("PR %d has been autosubmitted in %s", req.Issue.Number, commit.GetSHA())

	return nil
}

func (d Dreck) pullRequestStatus(ctx context.Context, client *github.Client, req types.IssueCommentOuter, pull *github.PullRequest) (bool, error) {

	listOpts := &github.ListOptions{PerPage: 100}
	combined, _, err := client.Repositories.GetCombinedStatus(ctx, req.Repository.Owner.Login, req.Repository.Name, pull.Head.GetSHA(), listOpts)
	if err != nil {
		return false, err
	}

	log.Infof("Checking %d statuses for PR %d", combined.GetTotalCount(), pull.GetNumber())

	for _, status := range combined.Statuses {
		if status.GetState() != statusOK {
			log.Infof("Status %s is %s", status.GetContext(), status.GetState())
			return false, nil
		}
	}

	log.Infof("All %d statuses for PR %d are in state %s", combined.GetTotalCount(), pull.GetNumber(), statusOK)
	return true, nil
}

const statusOK = "success"
