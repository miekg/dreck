package dreck

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/miekg/dreck/auth"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/github"
)

func handlePullRequest(req types.PullRequestOuter) error {
	ctx := context.Background()

	token := os.Getenv("access_token")
	if len(token) == 0 {

		newToken, tokenErr := auth.MakeAccessTokenForInstallation(os.Getenv("application"), req.Installation.ID)

		if tokenErr != nil {
			return tokenErr
		}

		token = newToken
	}

	client := auth.MakeClient(ctx, token)

	hasUnsignedCommits, err := hasUnsigned(req, client)

	if err != nil {
		return err
	}

	if hasUnsignedCommits {
		fmt.Println("May need to apply labels on item.")

		issue, _, labelErr := client.Issues.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number)

		if labelErr != nil {
			return labelErr
		}
		fmt.Println("Current labels ", issue.Labels)

		if hasNoDcoLabel(issue) == false {
			fmt.Println("Applying label")
			_, _, assignLabelErr := client.Issues.AddLabelsToIssue(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, []string{"no-dco"})
			if assignLabelErr != nil {
				return assignLabelErr
			}

			link := fmt.Sprintf("https://github.com/%s/%s/blob/master/CONTRIBUTING.md", req.Repository.Owner.Login, req.Repository.Name)
			body := `Thank you for your contribution. I've just checked and your commit doesn't appear to be signed-off.
That's something we need before your Pull Request can be merged. Please see our [contributing guide](` + link + `).`

			comment := &github.IssueComment{
				Body: &body,
			}

			comment, resp, err := client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, comment)
			if err != nil {
				return err
			}
			fmt.Println(comment, resp.Rate)
		}
	} else {
		fmt.Println("Things look OK right now.")
		issue, _, labelErr := client.Issues.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number)

		if labelErr != nil {
			return labelErr
		}

		if hasNoDcoLabel(issue) {
			fmt.Println("Removing label")
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

	fmt.Println("Rate limiting", resp.Rate)

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
