package dreck

// Copied from https://github.com/genuinetools/ghb0t/blob/master/main.go

import (
	"fmt"
	"strings"

	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"
)

// pullRequestBranches will ...
func (d Dreck) pullRequestBranches(req types.PullRequestOuter) error {
	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	pull, resp, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number)
	if err != nil {
		return fmt.Errorf("getting PR %d\n%s", req.PullRequest.Number, err.Error())
	}

	log.Warningf("Rate limiting: %s", resp.Rate)

	// Double check again.
	if *pull.State == "closed" && *pull.Merged {

		// Never delete the master branch.
		branch := *pull.Head.Ref
		if branch == "master" {
			log.Info("Not touching master branch")
			return nil
		}
		if pull.Head.Repo == nil {
			return fmt.Errorf("no head found")
		}
		if pull.Head.Repo.Owner == nil {
			return fmt.Errorf("no owner found")
		}

		log.Infof("Deleting branch %s on %s/%s", branch, req.Repository.Owner.Login, *pull.Head.Repo.Name)

		resp, err := client.Git.DeleteRef(ctx, req.Repository.Owner.Login, *pull.Head.Repo.Name, strings.Replace("heads/"+*pull.Head.Ref, "#", "%23", -1))
		// 422 is the error code for when the branch does not exist.
		if err != nil && resp.Response.StatusCode != 422 {
			return err
		}
		log.Infof("Branch %s on %s/%s no longer exists.", branch, req.Repository.Owner.Login, *pull.Head.Repo.Name)
	}

	return nil
}
