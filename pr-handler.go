package dreck

import (
	"context"
	"fmt"
	"strings"

	"github.com/miekg/dreck/auth"
	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/github"
)

func (d Dreck) handlePullRequest(req types.PullRequestOuter) error {
	ctx := context.Background()

	token, err := auth.MakeAccessTokenForInstallation(d.clientID, d.key, req.Installation.ID)
	if err != nil {
		return err
	}

	client := auth.MakeClient(ctx, token)

	hasUnsignedCommits, err := hasUnsigned(req, client)

	if err != nil {
		return err
	}

	if hasUnsignedCommits {
		issue, _, labelErr := client.Issues.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number)

		if labelErr != nil {
			return labelErr
		}
		log.Infof("Current labels: %s", issue.Labels)

		if hasNoDcoLabel(issue) == false {
			log.Info("Applying label")
			_, _, assignLabelErr := client.Issues.AddLabelsToIssue(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, []string{"no-dco"})
			if assignLabelErr != nil {
				return assignLabelErr
			}

			link := fmt.Sprintf("https://github.com/%s/%s/blob/master/CONTRIBUTING.md", req.Repository.Owner.Login, req.Repository.Name)
			body := `Thank you for your contribution. I've just checked and your commit doesn't appear to be signed-off.
That's something we need before your Pull Request can be merged. Please see our [contributing guide](` + link + `).`

			comment := githubIssueComment(body)

			comment, resp, err := client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, comment)
			if err != nil {
				return err
			}
			log.Infof("%s %s", comment, resp.Rate)
		}
	} else {
		issue, _, labelErr := client.Issues.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number)

		if labelErr != nil {
			return labelErr
		}

		if hasNoDcoLabel(issue) {
			log.Info("Removing label")
			_, removeLabelErr := client.Issues.RemoveLabelForIssue(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, "no-dco")
			if removeLabelErr != nil {
				return removeLabelErr
			}
		}
	}
	return nil
}

func hasNoDcoLabel(issue *github.Issue) bool {
	if issue != nil {
		for _, label := range issue.Labels {
			if label.GetName() == "no-dco" {
				return true
			}
		}
	}
	return false
}

func hasUnsigned(req types.PullRequestOuter, client *github.Client) (bool, error) {
	hasUnsigned := false
	ctx := context.Background()

	var err error
	listOpts := &github.ListOptions{
		Page: 0,
	}

	commits, resp, err := client.PullRequests.ListCommits(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, listOpts)
	if err != nil {
		return hasUnsigned, fmt.Errorf("getting PR %d\n%s", req.PullRequest.Number, err.Error())
	}

	log.Warningf("Rate limiting: %s", resp.Rate)

	for _, commit := range commits {
		if commit.Commit != nil && commit.Commit.Message != nil {
			if isSigned(*commit.Commit.Message) == false {
				hasUnsigned = true
			}
		}
	}

	return hasUnsigned, err
}

func isSigned(msg string) bool {
	return strings.Contains(msg, "Signed-off-by:")
}
