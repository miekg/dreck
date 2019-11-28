package dreck

import (
	"io/ioutil"
	"testing"
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
