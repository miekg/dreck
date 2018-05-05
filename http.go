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
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("dreck", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

type Dreck struct {
	Next httpserver.Handler
	// more
}

func (d Dreck) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}
	r.Body.Close()

	// HMAC Validated or not turned on.
	// eventType := os.Getenv("Http_X_Github_Event")
	// Check if this header exists and if not, call next plugin.

	xHubSignature := os.Getenv("Http_X_Hub_Signature")

	if hmacValidation() && len(xHubSignature) == 0 {
		log.Fatal("must provide X_Hub_Signature")
		return 0, nil
	}

	if len(xHubSignature) > 0 {

		err := auth.ValidateHMAC(body, xHubSignature)
		if err != nil {
			log.Fatal(err.Error())
			return 0, nil
		}
	}

	// HMAC Validated or not turned on.
	eventType := os.Getenv("Http_X_Github_Event")

	if err := handleEvent(eventType, body); err != nil {
		log.Fatal(err)
	}
	return d.Next.ServeHTTP(w, r)
}

func handleEvent(eventType string, body []byte) error {

	switch eventType {
	case "pull_request":
		req := types.PullRequestOuter{}
		if err := json.Unmarshal(body, &req); err != nil {
			return fmt.Errorf("Cannot parse input %s", err.Error())
		}

		customer, err := auth.IsCustomer(req.Repository)
		if err != nil {
			return fmt.Errorf("Unable to verify customer: %s/%s", req.Repository.Owner.Login, req.Repository.Name)
		} else if customer == false {
			return fmt.Errorf("No customer found for: %s/%s", req.Repository.Owner.Login, req.Repository.Name)
		}

		derekConfig, err := getConfig(req.Repository.Owner.Login, req.Repository.Name)
		if err != nil {
			return fmt.Errorf("Unable to access maintainers file at: %s/%s", req.Repository.Owner.Login, req.Repository.Name)
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
			return fmt.Errorf("Cannot parse input %s", err.Error())
		}

		customer, err := auth.IsCustomer(req.Repository)
		if err != nil {
			return fmt.Errorf("Unable to verify customer: %s/%s", req.Repository.Owner.Login, req.Repository.Name)
		} else if customer == false {
			return fmt.Errorf("No customer found for: %s/%s", req.Repository.Owner.Login, req.Repository.Name)
		}

		derekConfig, err := getConfig(req.Repository.Owner.Login, req.Repository.Name)
		if err != nil {
			return fmt.Errorf("Unable to access maintainers file at: %s/%s", req.Repository.Owner.Login, req.Repository.Name)
		}

		if req.Action != deleted {
			if permittedUserFeature(comments, derekConfig, req.Comment.User.Login) {
				handleComment(req)
			}
		}
		break
	default:
		return fmt.Errorf("X_Github_Event want: ['pull_request', 'issue_comment'], got: " + eventType)
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
