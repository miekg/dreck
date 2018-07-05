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

			pull, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
			if err != nil {
				return err
			}
			if pull.ClosedAt != nil {
				// Pr has been closed or deleted. Don't merge!
				return fmt.Errorf("PR %d has been deleted at %s", req.Issue.Number, pull.GetClosedAt())
			}

			ok, _ := d.pullRequestStatus(client, req, pull)
			if ok && pull.Mergeable != nil {
				err := d.pullRequestMerge(client, req, pull)
				if err != nil {
					return err
				}
			}

			d.pullRequestDeletePendingReviews(client, req, pull)
			return

		case <-stop.C:

			return fmt.Errorf("timeout while waiting for PR %d", req.Issue.Number)
		}
	}
}

func (d Dreck) pullRequestMerge(client *github.Client, req types.IssueCommentOuter, pull *github.PullRequest) error {

	ctx := context.Background()
	opt := &github.PullRequestOptions{MergeMethod: d.strategy}
	msg := "Automatically submitted."
	commit, _, err := client.PullRequests.Merge(ctx, req.Repository.Owner.Login, req.Repository.Name, *pull.Number, msg, opt)

	if err != nil {
		return fmt.Errorf("failed merge of PR %d: %s", *pull.Number, err.Error())
	}

	log.Infof("PR %d has been autosubmitted in %s", req.Issue.Number, commit.GetSHA())

	return nil
}

func (d Dreck) pullRequestStatus(client *github.Client, req types.IssueCommentOuter, pull *github.PullRequest) (bool, error) {

	ctx := context.Background()
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

func (d Dreck) pullRequestReviewed(client *github.Client, req types.IssueCommentOuter, pull *github.PullRequest) (bool, error) {

	ctx := context.Background()
	listOpts := &github.ListOptions{PerPage: 100}
	reviews, _, err := client.PullRequests.ListReviews(ctx, req.Repository.Owner.Login, req.Repository.Name, pull.GetNumber(), listOpts)

	if err != nil {
		return false, err
	}

	ok := false
	for _, review := range reviews {
		if review.GetState() == reviewChanges {
			ok = false
			break
		}
		if review.GetState() == reviewOK {
			ok = true
		}
	}
	if !ok {
		log.Infof("PR %d has not been approved", pull.GetNumber())
		return false, fmt.Errorf("PR %d is not reviews or has a %s", pull.GetNumber(), reviewChanges)
	}

	log.Infof("PR %d has been approved", pull.GetNumber())
	return true, nil
}

func (d Dreck) pullRequestDeletePendingReviews(client *github.Client, req types.IssueCommentOuter, pull *github.PullRequest) error {
	ctx := context.Background()
	listOpts := &github.ListOptions{PerPage: 100}
	reviews, _, err := client.PullRequests.ListReviews(ctx, req.Repository.Owner.Login, req.Repository.Name, pull.GetNumber(), listOpts)

	if err != nil {
		return false, err
	}

	for _, review := range reviews {
		// don't care about return code here.
		client.PullRequests.DeletePendingReview(ctx, req.Repository.Owner.Login, rep.Repository.Name, pull.GetNumber(), review.GetId())
	}

}

func (d Dreck) merge(req types.IssueCommentOuter) error {
	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	pull, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
	if err != nil {
		// Pr does not exist, noop.
		return err
	}
	if pull.ClosedAt != nil {
		// Pr has been closed or deleted. Don't merge!
		return fmt.Errorf("PR %d has been deleted at %s", req.Issue.Number, pull.GetClosedAt())
	}

	statusOK, _ := d.pullRequestStatus(client, req, pull)
	reviewOK, _ := d.pullRequestReviewed(client, req, pull)
	if statusOK && reviewOK && pull.Mergeable != nil {
		err := d.pullRequestMerge(client, req, pull)
		if err != nil {
			return err
		}
	}

	d.pullRequestDeletePendingReviews(client, req, pull)

	return nil
}

const (
	statusOK      = "success"
	statusPending = "pending"
	statusFail    = "failure"
	statusError   = "error"

	reviewOK      = "APPROVED"
	reviewChanges = "REQUEST_CHANGES"
	reviewComment = "COMMENT"
)
