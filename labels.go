package dreck

import (
	"context"
	"strings"

	"github.com/miekg/dreck/types"

	"github.com/google/go-github/v28/github"
)

func labelDuplicate(current []types.IssueLabel, label string) bool {
	for _, l := range current {
		if strings.EqualFold(l.Name, label) {
			return true
		}
	}
	return false
}

// allLabels returns the first 100 labels from the repository.
func (d Dreck) allLabels(ctx context.Context, client *github.Client, req types.IssueCommentOuter) ([]types.IssueLabel, error) {
	labels, _, err := client.Issues.ListLabels(ctx, req.Repository.Owner.Login, req.Repository.Name, &github.ListOptions{PerPage: 100, Page: 0})
	if err != nil {
		return nil, err
	}

	ret := make([]types.IssueLabel, len(labels))
	for i, l := range labels {
		ret[i].Name = l.GetName()
	}

	return ret, nil
}
