package dreck

import (
	"context"

	"github.com/miekg/dreck/auth"

	"github.com/google/go-github/github"
)

func githubIssueComment(body string) *github.IssueComment {
	return &github.IssueComment{
		Body: &body,
	}
}

func (d Dreck) newClient(installation int) (*github.Client, context.Context, error) {
	ctx := context.Background()

	token, err := auth.MakeAccessTokenForInstallation(d.clientID, d.key, installation)
	if err != nil {
		return nil, ctx, err
	}

	client := auth.MakeClient(ctx, token)

	return client, ctx, nil
}
