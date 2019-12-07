package dreck

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/miekg/dreck/auth"
	"github.com/miekg/dreck/log"

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
		if d.secret == "" {
			return 0, fmt.Errorf("must provide a secret")
		}
		err := auth.ValidateHMAC(d.secret, body, hubSignature)
		if err != nil {
			return 0, err
		}
	}

	err = d.handleEvent(event, body)
	return 0, err
}

func parseEvent(event string, body []byte) (IssueCommentOuter, error) {
	req := IssueCommentOuter{}
	if err := json.Unmarshal(body, &req); err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			return req, fmt.Errorf("Syntax error at byte offset %d", e.Offset)
		}
		return req, fmt.Errorf("parse error %s: %s", string(body), err.Error())
	}
	// If event is issues or pull_request; we copy some elements, so we can treat it as an issue comment.
	// But only when the issue/pr is created.
	switch event {
	case "issues":
		if req.Action == "opened" {
			req.Comment.Body = req.Issue.Body
		}
		req.Comment.User.Login = req.Issue.User.Login
	case "pull_request":
		if req.Action == "opened" {
			req.Comment.Body = req.PullRequest.Body
		}
		req.Issue.Number = req.PullRequest.Number
		req.Comment.User.Login = req.PullRequest.User.Login
	}

	return req, nil
}

func (d Dreck) handleEvent(event string, body []byte) error {
	switch event {
	case "issue_comment", "issues", "pull_request_review", "pull_request":
		req, err := parseEvent(event, body)
		if err != nil {
			return err
		}
		log.Infof("Action: %q for: %s", req.Action, req.Comment.User.Login)
		if strings.HasSuffix(req.Comment.User.Login, "[bot]") {
			return nil
		}

		// Do nothing on these actions
		switch req.Action {
		case "deleted":
			fallthrough
		case "synchronize":
			fallthrough
		case "locked":
			fallthrough
		case "labeled":
			return nil
		}

		conf, err := d.getConfig(req.Repository.Owner.Login, req.Repository.Name)
		if err != nil {
			return fmt.Errorf("unable to access maintainers file at %s/%s: %s", req.Repository.Owner.Login, req.Repository.Name, err)
		}
		if _, err := d.comment(req, conf); err != nil {
			return err
		}

	case "ping":
		fallthrough

	case "status":
		log.Infof("Seen %s", event)
		return nil

	default:
		return fmt.Errorf("unsupported event: %s", event)
	}

	return nil
}
