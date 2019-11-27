package dreck

import (
	"fmt"

	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/github"
)

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
		log.Infof("No searching for owners because of Work-in-Progress status for PR %d: %s", req.PullRequest.Number, title)
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
		body += "\nNote this is not an exclusive request. Anyone is free to provide a review of this pull request."
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
