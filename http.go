package dreck

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/miekg/dreck/auth"
	"github.com/miekg/dreck/types"

	"github.com/mholt/caddy"
)

func init() {
	caddy.RegisterPlugin("dreck", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func (d Dreck) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}
	r.Body.Close()

	// Give up if we can't find this event
	event := r.Header.Get("X-GitHub-Event")
	if event == "" {
		return d.Next.ServeHTTP(w, r)
	}

	// HMAC Validated or not turned on.
	xHubSignature := os.Getenv("Http_X_Hub_Signature")

	if hmacValidation() && len(xHubSignature) == 0 {
		return 0, fmt.Errorf("must provide X_Hub_Signature")
	}

	if len(xHubSignature) > 0 {

		err := auth.ValidateHMAC(body, xHubSignature)
		if err != nil {
			return 0, err
		}
	}

	err = handleEvent(event, body)
	return 0, err
}

func handleEvent(eventType string, body []byte) error {

	switch eventType {
	case "pull_request":
		req := types.PullRequestOuter{}
		if err := json.Unmarshal(body, &req); err != nil {
			if e, ok := err.(*json.SyntaxError); ok {
				log.Printf("syntax error at byte offset %d", e.Offset)
			}
			log.Printf("sakura response: %q", body)
			return fmt.Errorf("Parse error %s: %s", string(body), err.Error())
		}

		derekConfig, err := getConfig(req.Repository.Owner.Login, req.Repository.Name)
		if err != nil {
			return fmt.Errorf("Unable to access maintainers file at %s/%s: %s", req.Repository.Owner.Login, req.Repository.Name, err)
		}
		if req.Action != closedConst {
			if enabledFeature(dcoCheck, derekConfig) {
				handlePullRequest(req)
			}
		}
		break

	case "issue_comment":
		req := types.IssueCommentOuter{}
		if err := json.Unmarshal(body, &req); err != nil {
			if e, ok := err.(*json.SyntaxError); ok {
				log.Printf("syntax error at byte offset %d", e.Offset)
			}
			log.Printf("sakura response: %q", body)
			return fmt.Errorf("Parse error %s: %s", string(body), err.Error())
		}

		derekConfig, err := getConfig(req.Repository.Owner.Login, req.Repository.Name)
		if err != nil {
			return fmt.Errorf("Unable to access maintainers file at %s/%s: %s", req.Repository.Owner.Login, req.Repository.Name, err)
		}

		if req.Action != deleted {
			if permittedUserFeature(comments, derekConfig, req.Comment.User.Login) {
				handleComment(req)
			}
		}
		break

	case "ping":
		fallthrough

	case "status":
		log.Printf("[INFO] Seen %s", eventType)
		return nil

	default:
		return fmt.Errorf("unsupported event: %s", eventType)
	}

	return nil
}

func hmacValidation() bool {
	val := os.Getenv("validate_hmac")
	return len(val) > 0 && (val == "1" || val == "true")
}

const (
	dcoCheck = "dco_check"
	comments = "comments"
	deleted  = "deleted"
)
