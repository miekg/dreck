package dreck

import (
	"context"
	"fmt"
	"strings"

	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/github"
)

// pullRequestDCO handles the DCO check. I.e. a PR must have commits for Signed-off-by.
func (d Dreck) pullRequestDCO(req types.PullRequestOuter) error {
	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

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

		if !hasNoDcoLabel(issue) {
			log.Info("Applying label")
			_, _, assignLabelErr := client.Issues.AddLabelsToIssue(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, []string{"no-dco"})
			if assignLabelErr != nil {
				return assignLabelErr
			}

			link := fmt.Sprintf("https://github.com/%s/%s/blob/master/CONTRIBUTING.md", req.Repository.Owner.Login, req.Repository.Name)
			body := thanks + `I've just checked and your commit doesn't appear to be signed-off.
That's something we need before your Pull Request can be merged. Please see our [contributing guide](` + link + `).`

			comment := githubIssueComment(body)

			comment, resp, err := client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, comment)
			if err != nil {
				return err
			}

			logRateLimit(resp)

		}

		return nil
	}

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

	listOpts := &github.ListOptions{Page: 0}

	commits, resp, err := client.PullRequests.ListCommits(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, listOpts)
	if err != nil {
		return hasUnsigned, fmt.Errorf("getting PR %d: %s", req.PullRequest.Number, err.Error())
	}

	logRateLimit(resp)

	for _, commit := range commits {
		if commit.Commit != nil && commit.Commit.Message != nil {
			if issigned(*commit.Commit.Message) == false {
				hasUnsigned = true
			}
		}
	}

	return hasUnsigned, nil
}

func issigned(msg string) bool { return strings.Contains(msg, "Signed-off-by:") }

// pullRequestReview will look at the (first 5) files of a PR, retrieve the nearest OWNERS files
// merge all the reviewers and randomly pick a reviewer that should be assigned for this PR.
func (d Dreck) pullRequestReviewers(req types.PullRequestOuter) error {
	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	listOpts := &github.ListOptions{PerPage: 100} // only the first page of files, but up to 100 files
	files, resp, err := client.PullRequests.ListFiles(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, listOpts)
	if err != nil {
		return fmt.Errorf("getting PR %d: %s", req.PullRequest.Number, err.Error())
	}

	pull, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number)
	if err != nil {
		return fmt.Errorf("getting PR %d: %s", req.PullRequest.Number, err.Error())
	}

	title := pull.GetTitle()
	log.Infof("Title for PR %d: %s", req.PullRequest.Number, title)
	if hasWIPPrefix(title) {
		// We used to add a comment in the PR that no reviewer was assigned, stop doing that to cut back on the spam.
		return nil
	}

	// ignore err here, we want to see if there are 0 reviewers
	reviewers, _, _ := client.PullRequests.ListReviewers(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, listOpts)
	if reviewers != nil {
		if len(reviewers.Users) != 0 || len(reviewers.Teams) != 0 {
			// We used to set an issue comment; skip this.
			return nil
		}
	}

	logRateLimit(resp)

	victim, file := d.findReviewers(files, *pull.User.Login, func(path string) ([]byte, error) {
		return githubFile(req.Repository.Owner.Login, req.Repository.Name, path)
	})

	if victim != "" {
		rev := github.ReviewersRequest{Reviewers: []string{victim}}
		if _, _, err := client.PullRequests.RequestReviewers(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, rev); err != nil {
			return err
		}
	}

	body := thanks + "I've just checked the *%s* files to find a suitable reviewer."
	if victim != "" {
		body += " This search was successful and I've asked **%s** (via `%s`) for a review."
		body = fmt.Sprintf(body, d.owners, victim, file)
	} else {
		body += " Alas, this search was *not* successful."
		body = fmt.Sprintf(body, d.owners)
	}

	comment := githubIssueComment(body + Details)
	comment, resp, err = client.Issues.CreateComment(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number, comment)

	logRateLimit(resp)

	return err
}

// PullRequestWIP check if the title changed from WIP to non WIP and assigns reviewers if needed. It
// returns true if we need to assign reviewers otherwise false.
func (d Dreck) pullRequestWIP(req types.PullRequestOuter) (bool, error) {
	title, ok := req.Changes["title"]
	if !ok {
		log.Info("No title changes, doing nothing")
		return false, nil
	}
	from, ok := title["from"]
	if !ok {
		log.Info("No title changes, doing nothing")
		return false, nil
	}

	cur, err := d.pullRequestTitle(req)
	if err != nil {
		return false, err
	}

	// If the previous PR title had WIP prefix and this one hasn't we assume we went from WIP -> no WIP.
	if hasWIPPrefix(from) && !hasWIPPrefix(cur) {
		return true, nil
	}

	return false, nil
}

const thanks = "Thank you for your contribution. " // leave space after the .
