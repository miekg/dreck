package dreck

import (
	"io/ioutil"
	"testing"

	"github.com/miekg/dreck/types"
)

func TestParseComment(t *testing.T) {
	f := "testdata/issue_comment.json"
	event := "issue_comment"
	body, err := ioutil.ReadFile(f)
	if err != nil {
		t.Error(err)
	}

	req, err := parseEvent(event, body)
	if err != nil {
		t.Error(err)
	}
	if l := req.Comment.User.Login; l != "miekg" {
		t.Errorf("expected login to be %s, got %s", "miekg", l)
	}
}

func TestParseIssueOpen(t *testing.T) {
	f := "testdata/issues-open.json"
	event := "issues"
	body, err := ioutil.ReadFile(f)
	if err != nil {
		t.Error(err)
	}

	req, err := parseEvent(event, body)
	if err != nil {
		t.Error(err)
	}
	if l := req.Comment.User.Login; l != "miekg" {
		t.Errorf("expected login to be %s, got %s", "miekg", l)
	}
}

func TestParsePullRequest(t *testing.T) {
	f := "testdata/pull_request.json"
	event := "pull_request"
	body, err := ioutil.ReadFile(f)
	if err != nil {
		t.Error(err)
	}

	req, err := parseEvent(event, body)
	if err != nil {
		t.Error(err)
	}
	if l := req.Comment.User.Login; l != "miekg" {
		t.Errorf("expected login to be %s, got %s", "miekg", l)
	}
	if l := req.Comment.Body; l != "/label moo" {
		t.Errorf("expected body to be %s, got %s", "/label moo", l)
	}
}

func TestParsePullRequesNumbert(t *testing.T) {
	f := "testdata/pull_request.json"
	event := "pull_request"
	body, err := ioutil.ReadFile(f)
	if err != nil {
		t.Error(err)
	}

	req, err := parseEvent(event, body)
	if err != nil {
		t.Error(err)
	}
	if l := req.PullRequest.Number; l != 128 {
		t.Errorf("expected number to be %d, got %d", 128, l)
	}
}

func TestParseMultipleCommands(t *testing.T) {
	f := "testdata/issue_comment-multiple.json"
	event := "issue_comment"
	body, err := ioutil.ReadFile(f)
	if err != nil {
		t.Error(err)
	}

	req, err := parseEvent(event, body)
	if err != nil {
		t.Error(err)
	}
	conf := &types.DreckConfig{}
	c := parse(req.Comment.Body, conf)
	if len(c) != 2 {
		t.Errorf("expected 2 commands, got %d", len(c))
	}
	for i := range c {
		if c[i].Type == "lgtm" {
			if c[i].Value != "" {
				t.Errorf("expected not value, got %s", c[i].Value)
			}
		}
	}
}

func TestParseDulicate(t *testing.T) {
	f := "testdata/issue_comment-duplicate.json"
	event := "issue_comment"
	body, err := ioutil.ReadFile(f)
	if err != nil {
		t.Error(err)
	}

	req, err := parseEvent(event, body)
	if err != nil {
		t.Error(err)
	}
	conf := &types.DreckConfig{}
	c := parse(req.Comment.Body, conf)
	if c[0].Type != "duplicate" {
		t.Errorf("expected not duplicate, got %s", c[0].Type)
	}
}
