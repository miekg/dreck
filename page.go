package dreck

import (
	"context"

	"github.com/google/go-github/v28/github"
)

func ListLabels(ctx context.Context, client *github.Client, req IssueCommentOuter) ([]*github.Label, error) {
	opt := &github.ListOptions{PerPage: 100}
	allLabels := []*github.Label{}
	for {
		labels, resp, err := client.Issues.ListLabels(ctx, req.Repository.Owner.Login, req.Repository.Name, opt)
		if err != nil {
			return nil, err
		}
		allLabels = append(allLabels, labels...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allLabels, nil
}

func ListReviews(ctx context.Context, client *github.Client, req IssueCommentOuter, pull *github.PullRequest) ([]*github.PullRequestReview, error) {
	opt := &github.ListOptions{PerPage: 100}
	allReviews := []*github.PullRequestReview{}
	for {
		reviews, resp, err := client.PullRequests.ListReviews(ctx, req.Repository.Owner.Login, req.Repository.Name, pull.GetNumber(), opt)
		if err != nil {
			return nil, err
		}
		allReviews = append(allReviews, reviews...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allReviews, nil
}
