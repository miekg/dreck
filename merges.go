package dreck

import (
	"context"
	"fmt"

	"github.com/miekg/dreck/log"

	"github.com/google/go-github/v28/github"
)

func (d Dreck) pullRequestMerge(ctx context.Context, client *github.Client, req IssueCommentOuter, pull *github.PullRequest) (*github.Response, error) {
	opt := &github.PullRequestOptions{MergeMethod: d.strategy}
	msg := "Automatically submitted."
	_, resp, err := client.PullRequests.Merge(ctx, req.Repository.Owner.Login, req.Repository.Name, *pull.Number, msg, opt)
	return resp, err
}

func (d Dreck) pullRequestStatus(ctx context.Context, client *github.Client, req IssueCommentOuter, pull *github.PullRequest) (bool, error) {
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

func (d Dreck) pullRequestReviewed(ctx context.Context, client *github.Client, req IssueCommentOuter, pull *github.PullRequest) (bool, error) {
	listOpts := &github.ListOptions{PerPage: 100}
	reviews, _, err := client.PullRequests.ListReviews(ctx, req.Repository.Owner.Login, req.Repository.Name, pull.GetNumber(), listOpts)

	if err != nil {
		return false, err
	}

	ok := false
	for _, review := range reviews {
		if review.GetState() == "REQUEST_CHANGES" {
			ok = false
			break
		}
		if review.GetState() == "APPROVED" {
			ok = true
		}
	}
	if !ok {
		return false, fmt.Errorf("PR %d is no reviewers or has a %s", pull.GetNumber(), "REQUEST_CHANGES")
	}
	return true, nil
}

func (d Dreck) merge(ctx context.Context, client *github.Client, req IssueCommentOuter) (*github.Response, error) {
	pull, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
	if err != nil {
		// Pr does not exist, noop.
		return nil, err
	}
	if pull.ClosedAt != nil {
		// Pr has been closed or deleted. Don't merge!
		return nil, fmt.Errorf("PR %d has been deleted at %s", req.Issue.Number, pull.GetClosedAt())
	}

	ok1, _ := d.pullRequestStatus(ctx, client, req, pull)
	ok2, _ := d.pullRequestReviewed(ctx, client, req, pull)
	if ok1 && ok2 && pull.Mergeable != nil {
		return d.pullRequestMerge(ctx, client, req, pull)
	}
	return nil, nil
}

const (
	statusOK      = "success"
	statusPending = "pending"
	statusFail    = "failure"
	statusError   = "error"
)
