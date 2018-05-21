package dreck

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

	log.Infof("Pull %+v", pull)

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

		println(strings.Replace("heads/"+*pull.Head.Ref, "#", "%23", -1))
		return nil

		_, err := client.Git.DeleteRef(ctx, req.Repository.Owner.Login, *pull.Head.Repo.Name, strings.Replace("heads/"+*pull.Head.Ref, "#", "%23", -1))
		// 422 is the error code for when the branch does not exist.
		if err != nil && !strings.Contains(err.Error(), " 422 ") {
			return err
		}
		log.Infof("Branch %s on %s/%s no longer exists.", branch, req.Repository.Owner.Login, *pull.Head.Repo.Name)
	}

	return nil
}
