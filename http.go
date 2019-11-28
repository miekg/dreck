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
	case "issue_comment", "pull_request_review":
		req := types.IssueCommentOuter{}
		if err := json.Unmarshal(body, &req); err != nil {
			if e, ok := err.(*json.SyntaxError); ok {
				log.Errorf("Syntax error at byte offset %d", e.Offset)
			}
			return fmt.Errorf("parse error %s: %s", string(body), err.Error())
		}

		log.Infof("Comment action %s", req.Action)

		// Do nothing on deletion
		if req.Action == "deleted" {
			return nil
		}

		conf, err := d.getConfig(req.Repository.Owner.Login, req.Repository.Name)
		if err != nil {
			return fmt.Errorf("unable to access maintainers file at %s/%s: %s", req.Repository.Owner.Login, req.Repository.Name, err)
		}
		if err := d.comment(req, conf); err != nil {
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
