package dreck

import (
	"context"
	"fmt"
	"time"

	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/v28/github"
)

func (d Dreck) pullRequestMerge(ctx context.Context, client *github.Client, req types.IssueCommentOuter, pull *github.PullRequest) error {
	opt := &github.PullRequestOptions{MergeMethod: d.strategy}
	msg := "Automatically submitted."
	_, _, err := client.PullRequests.Merge(ctx, req.Repository.Owner.Login, req.Repository.Name, *pull.Number, msg, opt)
	return err
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
	return true, nil
}

func (d Dreck) pullRequestReviewed(ctx context.Context, client *github.Client, req types.IssueCommentOuter, pull *github.PullRequest) (bool, error) {
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
		return false, fmt.Errorf("PR %d is no reviewers or has a %s", pull.GetNumber(), reviewChanges)
	}
	return true, nil
}

func (d Dreck) merge(ctx context.Context, client *github.Client, req types.IssueCommentOuter) error {
	pull, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
	if err != nil {
		// Pr does not exist, noop.
		return err
	}
	if pull.ClosedAt != nil {
		// Pr has been closed or deleted. Don't merge!
		return fmt.Errorf("PR %d has been deleted at %s", req.Issue.Number, pull.GetClosedAt())
	}

	// wait a tiny bit for checking these; we see \lgtm and \merge together a lot, give /merge a small head start
	time.Sleep(2 * time.Second)

	statusOK, _ := d.pullRequestStatus(ctx, client, req, pull)
	reviewOK, _ := d.pullRequestReviewed(ctx, client, req, pull)
	if statusOK && reviewOK && pull.Mergeable != nil {
		return d.pullRequestMerge(ctx, client, req, pull)
	}
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
