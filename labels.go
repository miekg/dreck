package dreck

import (
	"context"
	"strings"

	"github.com/google/go-github/v28/github"
)

func labelDuplicate(current []IssueLabel, label string) bool {
	for _, l := range current {
		if strings.EqualFold(l.Name, label) {
			return true
		}
	}
	return false
}

// allLabels returns the first 100 labels from the repository.
func (d Dreck) allLabels(ctx context.Context, client *github.Client, req IssueCommentOuter) ([]IssueLabel, error) {
	labels, err := ListLabels(ctx, client, req)
	if err != nil {
		return nil, err
	}

	ret := make([]IssueLabel, len(labels))
	for i, l := range labels {
		ret[i].Name = l.GetName()
	}

	return ret, nil
}
