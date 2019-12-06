package dreck

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/v28/github"
)

const (
	openConst        = "open"
	openPRConst      = "opened"
	closedConst      = "closed"
	closeConst       = "close"
	reopenConst      = "reopen"
	lockConst        = "Lock"
	unlockConst      = "Unlock"
	titleConst       = "SetTitle"
	assignConst      = "Assign"
	unassignConst    = "Unassign"
	removeLabelConst = "RemoveLabel"
	addLabelConst    = "AddLabel"

	ccConst        = "cc"
	unccConst      = "uncc"
	lgtmConst      = "lgtm"
	unlgtmConst    = "unlgtm"
	approveConst   = "approve"
	unapproveConst = "unapprove"
	execConst      = "exec"
	retestConst    = "retest"
	duplicateConst = "duplicate"
	mergeConst     = "merge"
	fortuneConst   = "fortune"
	blockConst     = "block"
	unblockConst   = "unblock"
)

func (d Dreck) comment(req types.IssueCommentOuter, conf *types.DreckConfig) error {
	body := strings.ToLower(req.Comment.Body)
	cs := parse(body, conf)

	if isCodeOwner(conf, req.Comment.User.Login) {
		log.Infof("user %s is a code owner", req.Comment.User.Login)
	} else {
		log.Infof("user %s is not a code owner", req.Comment.User.Login)
	}
	if len(cs) == 0 {
		return nil
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	log.Infof("executing %d commands", len(cs))
For:
	for _, c := range cs {
		if err != nil {
			log.Error(err)
		}
		switch c.Type {
		case addLabelConst, removeLabelConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				err = d.label(ctx, client, req, c)
				continue For
			}
			err = fmt.Errorf("user %s not permitted to use [un]label", req.Comment.User.Login)
		case assignConst, unassignConst:
			if isMe(req.Comment.User.Login, c.Value) || isCodeOwner(conf, req.Comment.User.Login) {
				err = d.assign(ctx, client, req, c)
				continue For
			}
			err = fmt.Errorf("user %s not permitted to use [un]assign", req.Comment.User.Login)
		case closeConst, reopenConst:
			err = d.state(ctx, client, req, c)
		case titleConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				err = d.title(ctx, client, req, c)
				continue For
			}
			err = fmt.Errorf("user %s not permitted to use title", req.Comment.User.Login)
		case lockConst, unlockConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				err = d.lock(ctx, client, req, c)
				continue For
			}
			err = fmt.Errorf("user %s not permitted to use [un]lock", req.Comment.User.Login)
		case approveConst, unapproveConst:
			fallthrough
		case lgtmConst, unlgtmConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				err = d.lgtm(ctx, client, req, c)
				continue For
			}
			err = fmt.Errorf("user %s not permitted to use [un]lgtm", req.Comment.User.Login)
		case ccConst, unccConst:
			if isMe(req.Comment.User.Login, c.Value) || isCodeOwner(conf, req.Comment.User.Login) {
				err = d.cc(ctx, client, req, c)
				continue For
			}
			err = fmt.Errorf("user %s not permitted to use [un]cc", req.Comment.User.Login)
		case retestConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				err = d.retest(ctx, client, req, c)
				continue For
			}
			err = fmt.Errorf("user %s not permitted to use retest", req.Comment.User.Login)
		case duplicateConst:
			err = d.duplicate(ctx, client, req, c)
		case fortuneConst:
			err = d.fortune(ctx, client, req, c)
		case execConst:
			if !aliasOK(conf) {
				err = fmt.Errorf("feature %s is not enabled, so %s can't work", Trigger+execConst, Aliases)
				continue For
			}
			if !isCodeOwner(conf, req.Comment.User.Login) {
				err = fmt.Errorf("user %s not permitted to use exec", req.Comment.User.Login)
				continue For
			}
			err = d.exec(ctx, client, req, conf, c)
		case mergeConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				err = d.merge(ctx, client, req)
				continue For
			}
			err = fmt.Errorf("user %s is not a code owner", req.Comment.User.Login)
		case blockConst, unblockConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				err = d.block(ctx, client, req, c)
				continue For
			}
			err = fmt.Errorf("user %s is not a code owner", req.Comment.User.Login)
		}
	}

	return err
}

func (d Dreck) label(ctx context.Context, client *github.Client, req types.IssueCommentOuter, c *types.Action) error {
	labels, err := d.allLabels(ctx, client, req)
	if err != nil {
		return err
	}
	if found := labelDuplicate(labels, c.Value); !found {
		return fmt.Errorf("label %s does not exist", c.Value)
	}

	if c.Type == addLabelConst {
		_, _, err = client.Issues.AddLabelsToIssue(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, []string{c.Value})
	} else {
		_, err = client.Issues.RemoveLabelForIssue(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, c.Value)
	}
	return err
}

func (d Dreck) title(ctx context.Context, client *github.Client, req types.IssueCommentOuter, c *types.Action) error {
	newTitle := c.Value
	if newTitle == req.Issue.Title || len(newTitle) == 0 {
		return fmt.Errorf("setting the title of #%d by %s was unsuccessful as the new title was empty or unchanged", req.Issue.Number, req.Comment.User.Login)
	}

	input := &github.IssueRequest{Title: &newTitle}
	_, _, err := client.Issues.Edit(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, input)
	return err
}

func (d Dreck) assign(ctx context.Context, client *github.Client, req types.IssueCommentOuter, c *types.Action) error {
	if len(c.Value) > 1 && c.Value[0] == '@' {
		c.Value = c.Value[1:]
	}

	if c.Value == "me" || c.Value == "" {
		c.Value = req.Comment.User.Login
	}

	if c.Type == unassignConst {
		_, _, err := client.Issues.RemoveAssignees(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, []string{c.Value})
		return err
	} else {
		_, _, err := client.Issues.AddAssignees(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, []string{c.Value})
		return err
	}
	return nil
}

func (d Dreck) cc(ctx context.Context, client *github.Client, req types.IssueCommentOuter, c *types.Action) error {
	if len(c.Value) > 1 && c.Value[0] == '@' {
		c.Value = c.Value[1:]
	}

	if c.Value == "me" || c.Value == "" {
		c.Value = req.Comment.User.Login
	}

	// check if this a pull request, if not call assign
	_, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
	if err != nil {
		log.Infof("not a pull request: %d", req.Issue.Number)
		return d.assign(ctx, client, req, c)
	}

	rev := github.ReviewersRequest{Reviewers: []string{c.Value}}
	if c.Type == ccConst {
		_, _, err := client.PullRequests.RequestReviewers(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, rev)
		return err
	} else {
		_, err := client.PullRequests.RemoveReviewers(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, rev)
		return err
	}
	return nil
}

func (d Dreck) state(ctx context.Context, client *github.Client, req types.IssueCommentOuter, c *types.Action) error {
	newState, validTransition := checkTransition(c.Type, req.Issue.State)
	if !validTransition {
		return fmt.Errorf("request to %s issue #%d by %s was invalid", c.Type, req.Issue.Number, req.Comment.User.Login)
	}

	input := &github.IssueRequest{State: &newState}
	_, _, err := client.Issues.Edit(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, input)
	return err
}

func (d Dreck) lock(ctx context.Context, client *github.Client, req types.IssueCommentOuter, c *types.Action) error {
	if !isAction(req.Issue.Locked, c.Type, lockConst, unlockConst) {
		return fmt.Errorf("issue #%d is already %sed", req.Issue.Number, c.Type)
	}

	if c.Type == lockConst {
		_, err := client.Issues.Lock(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, &github.LockIssueOptions{})
		return err
	} else {
		_, err := client.Issues.Unlock(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
		return err
	}
	return nil
}

func (d Dreck) block(ctx context.Context, client *github.Client, req types.IssueCommentOuter, c *types.Action) error {
	if c.Type == blockConst {
		_, err := client.Organizations.BlockUser(ctx, req.Repository.Owner.Login, c.Value)
		return err
	} else {
		_, err := client.Organizations.UnblockUser(ctx, req.Repository.Owner.Login, c.Value)
		return err
	}
	return nil
}

func (d Dreck) lgtm(ctx context.Context, client *github.Client, req types.IssueCommentOuter, c *types.Action) error {
	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	_, _, err = client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
	// will be 404 not found if this isn't a PR.
	if err != nil {
		return err
	}

	input := &github.PullRequestReviewRequest{}
	if c.Type == lgtmConst {
		input = &github.PullRequestReviewRequest{
			Body:  github.String("Approved by **" + req.Comment.User.Login + "**"),
			Event: github.String(reviewOK),
		}
	} else {
		input = &github.PullRequestReviewRequest{
			Body:  github.String("Unapproved by **" + req.Comment.User.Login + "**"),
			Event: github.String(reviewChanges),
		}
	}

	_, _, err = client.PullRequests.CreateReview(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, input)
	return err
}

func (d Dreck) retest(ctx context.Context, client *github.Client, req types.IssueCommentOuter, c *types.Action) error {
	return nil
}

func (d Dreck) duplicate(ctx context.Context, client *github.Client, req types.IssueCommentOuter, c *types.Action) error {
	if err := d.label(ctx, client, req, &types.Action{Type: addLabelConst, Value: "duplicate"}); err != nil {
		return err
	}
	return d.state(ctx, client, req, &types.Action{Type: closeConst, Value: ""})
}

// Body must be downcased already.
func parse(body string, conf *types.DreckConfig) []*types.Action {
	actions := []*types.Action{}

	for trigger, commandType := range IssueCommands {
		if val := isValidCommand(body, trigger, conf); len(val) > 0 {
			for _, v := range val {
				actions = append(actions, &types.Action{Type: commandType, Value: v})
				// limit the amount of actions we allow.
				if len(actions) == 10 {
					return actions
				}
			}
		}
	}

	return actions
}

// isValidCommand checks the body of the comment to see if trigger is present. Commands
// are recognized if the are on a line by them selves and are placed at the beginning.
// Body must be lowercased.
func isValidCommand(body string, trigger string, conf *types.DreckConfig) []string {
	if aliasOK(conf) {
		for _, a := range conf.Aliases {
			r, err := NewAlias(a)
			if err != nil {
				log.Warningf("Failed to parse alias: %s, %v", a, err)
				continue
			}
			body = r.Expand(body) // either noop or replaces something
		}
	}

	if len(body) < len(trigger) {
		return nil
	}

	val := []string{}
	bodyr := bufio.NewReader(strings.NewReader(body + "\n"))
	for {
		line, err := bodyr.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.Trim(line, " \n\t\r")
		if line == trigger {
			val = append(val, "")
			continue
		}

		if !strings.HasPrefix(line, trigger+" ") && !strings.HasPrefix(line, trigger+"\t") {
			continue
		}

		v := line[len(trigger):]
		v = strings.Trim(v, " \t.,\n\r")
		val = append(val, v)
	}

	return val
}

// isMe returns true if login equals value or value is empty or "me"
func isMe(login, value string) bool {
	if value == "me" || value == "" {
		return true
	}
	if login == value {
		return true
	}
	// check if value starts with @
	if len(value) > 1 && value[1] == '@' && login == value[1:] {
		return true
	}
	return false
}

func isAction(running bool, requestedAction string, start string, stop string) bool {
	return !running && requestedAction == start || running && requestedAction == stop
}

func checkTransition(requestedAction string, currentState string) (string, bool) {
	if requestedAction == closeConst && currentState != closedConst {
		return closedConst, true
	}
	if requestedAction == reopenConst && currentState != openConst {
		return openConst, true
	}

	return "", false
}

// IssueCommands are all commands we support in issues.
var IssueCommands = map[string]string{
	Trigger + "label":     addLabelConst,
	Trigger + "unlabel":   removeLabelConst,
	Trigger + "cc":        ccConst,
	Trigger + "uncc":      unccConst,
	Trigger + "assign":    assignConst,
	Trigger + "unassign":  unassignConst,
	Trigger + "close":     closeConst,
	Trigger + "reopen":    reopenConst,
	Trigger + "title":     titleConst,
	Trigger + "lock":      lockConst,
	Trigger + "unlock":    unlockConst,
	Trigger + "exec":      execConst,
	Trigger + "fortune":   fortuneConst,
	Trigger + "duplicate": duplicateConst,
	Trigger + "retest":    retestConst, // Only works on Pull Request comments.
	Trigger + "lgtm":      lgtmConst,   // Only works on Pull Request comments.
	Trigger + "unlgtm":    unlgtmConst, // Only works on Pull Request comments.
	Trigger + "merge":     mergeConst,  // Only works on Pull Request comments.
}
