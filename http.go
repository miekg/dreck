package dreck

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/miekg/dreck/auth"
	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/caddyserver/caddy"
)

func init() {
	caddy.RegisterPlugin("dreck", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func (d Dreck) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	// Give up if we can't find this header in the event.
	event := r.Header.Get("X-GitHub-Event")
	if event == "" {
		return d.Next.ServeHTTP(w, r)
	}

	// Not the correct path.
	if !strings.HasPrefix(r.URL.Path, d.path) {
		return d.Next.ServeHTTP(w, r)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}
	r.Body.Close()

	hubSignature := r.Header.Get("X-Hub-Signature")
	if d.hmac {
		if hubSignature == "" {
			return 0, fmt.Errorf("must provide X-Hub-Signature")
		}
		err := auth.ValidateHMAC(body, hubSignature)
		if err != nil {
			return 0, err
		}
	}

	err = d.handleEvent(event, body)
	return 0, err
}

func (d Dreck) handleEvent(event string, body []byte) error {
	switch event {

	}

	switch event {
	case "pull_request":
		req := types.PullRequestOuter{}
		if err := json.Unmarshal(body, &req); err != nil {
			if e, ok := err.(*json.SyntaxError); ok {
				log.Errorf("Syntax error at byte offset %d", e.Offset)
			}
			return fmt.Errorf("parse error %s: %s", string(body), err.Error())
		}

		log.Infof("Pull request action %s", req.Action)

		conf, err := d.getConfig(req.Repository.Owner.Login, req.Repository.Name)
		if err != nil {
			return fmt.Errorf("Unable to access maintainers file at %s/%s: %s", req.Repository.Owner.Login, req.Repository.Name, err)
		}

		// Branch deletion handling. Only done when req.Action is closed (happens after merge).
		if req.Action == closedConst {
			if enabledFeature(featureBranches, conf) {
				d.pullRequestBranches(req)
			}

			// delete pending reviews

			client, ctx, err := d.newClient(req.Installation.ID)
			if err != nil {
				return err
			}

			pull, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.PullRequest.Number)
			if err != nil {
				// Pr does not exist, noop.
				return err
			}

			return d.pullRequestDeletePendingReviews(client, types.PullRequestToIssueComment(req), pull)
		}

		// Reviewers, title change WIP, none WIP.
		if req.Action == "edited" {
			ok, err := d.pullRequestWIP(req)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			if enabledFeature(featureReviewers, conf) {
				if err := d.pullRequestReviewers(req); err != nil {
					return err
				}
			}
		}

		autosubmit, _ := d.isAutosubmit(req, conf)

		// Reviewers, only on PR opens.
		if req.Action == openPRConst {
			if enabledFeature(featureReviewers, conf) && !autosubmit {
				if err := d.pullRequestReviewers(req); err != nil {
					return err
				}
			}
		}

		if req.Action == openPRConst && autosubmit {
			if err := d.pullRequestAutosubmit(req); err != nil {
				return err
			}
		}

	case "issue_comment", "pull_request_review":
		req := types.IssueCommentOuter{}
		if err := json.Unmarshal(body, &req); err != nil {
			if e, ok := err.(*json.SyntaxError); ok {
				log.Errorf("Syntax error at byte offset %d", e.Offset)
			}
			return fmt.Errorf("parse error %s: %s", string(body), err.Error())
		}

		log.Infof("Issue comment action %s", req.Action)

		// Do nothing when the comment is deleted.
		if req.Action == "deleted" {
			return nil
		}

		conf, err := d.getConfig(req.Repository.Owner.Login, req.Repository.Name)
		if err != nil {
			return fmt.Errorf("Unable to access maintainers file at %s/%s: %s", req.Repository.Owner.Login, req.Repository.Name, err)
		}

		if permittedUserFeature(featureComments, conf, req.Comment.User.Login) {
			err := d.comment(req, conf)
			if err != nil {
				return err
			}
		}

	case "ping":
		fallthrough

	case "status":
		log.Infof("Seen %s", eventType)
		return nil

	default:
		return fmt.Errorf("unsupported event: %s", eventType)
	}

	return nil
}
