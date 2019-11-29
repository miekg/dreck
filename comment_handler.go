package dreck

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/github"
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
	testConst      = "test"
	duplicateConst = "duplicate"
	mergeConst     = "merge"
	fortuneConst   = "fortune"
)

func (d Dreck) comment(req types.IssueCommentOuter, conf *types.DreckConfig) error {
	body := strings.ToLower(req.Comment.Body)
	c := parse(body, conf)

	if isCodeOwner(conf, req.Comment.User.Login) {
		log.Infof("user %s is a code owner", req.Comment.User.Login)
	} else {
		log.Infof("user %s is not a code owner", req.Comment.User.Login)
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	log.Infof("executing %d commands", len(c))
For:
	for _, command := range c {
		if err != nil {
			log.Error(err)
		}
		switch command.Type {
		case addLabelConst, removeLabelConst:
			if isMe(req.Comment.User.Login, command.Value) || isCodeOwner(conf, req.Comment.User.Login) {
				err = d.label(ctx, client, req, command.Type, command.Value)
				continue For
			}
			err = fmt.Errorf("user %s not permitted to use [un]label", req.Comment.User.Login)
		case assignConst, unassignConst:
			if isMe(req.Comment.User.Login, command.Value) || isCodeOwner(conf, req.Comment.User.Login) {
				err = d.assign(ctx, client, req, command.Type, command.Value)
				continue For
			}
			err = fmt.Errorf("user %s not permitted to use [un]assign", req.Comment.User.Login)
		case closeConst, reopenConst:
			err = d.state(ctx, client, req, command.Type)
		case titleConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				err = d.title(ctx, client, req, command.Type, command.Value)
				continue For
			}
			err = fmt.Errorf("user %s not permitted to use title", req.Comment.User.Login)
		case lockConst, unlockConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				err = d.lock(ctx, client, req, command.Type)
				continue For
			}
			err = fmt.Errorf("user %s not permitted to use [un]lock", req.Comment.User.Login)
		case approveConst, unapproveConst:
			fallthrough
		case lgtmConst, unlgtmConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				err = d.lgtm(ctx, client, req, command.Type)
				continue For
			}
			err = fmt.Errorf("user %s not permitted to use [un]lgtm", req.Comment.User.Login)
		case ccConst, unccConst:
			if isMe(req.Comment.User.Login, command.Value) || isCodeOwner(conf, req.Comment.User.Login) {
				err = d.cc(ctx, client, req, command.Type, command.Value)
				continue For
			}
			err = fmt.Errorf("user %s not permitted to use [un]cc", req.Comment.User.Login)
		case testConst:
			err = d.test(ctx, client, req, command.Type, command.Value)
		case duplicateConst:
			err = d.duplicate(ctx, client, req, command.Type, command.Value)
		case fortuneConst:
			err = d.fortune(ctx, client, req, command.Type)
		case execConst:
			if !aliasOK(conf) {
				err = fmt.Errorf("feature %s is not enabled, so %s can't work", Trigger+execConst, Aliases)
				continue For
			}
			if !isCodeOwner(conf, req.Comment.User.Login) {
				err = fmt.Errorf("user %s not permitted to use exec", req.Comment.User.Login)
				continue For
			}
			err = d.exec(ctx, client, req, conf, command.Type, command.Value)
		case mergeConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				err = d.merge(ctx, client, req)
				continue For
			}
			err = fmt.Errorf("user %s is not a code owner", req.Comment.User.Login)
		}
	}

	if len(c) == 0 {
		log.Info("No command found")
	}
	return err
}

func (d Dreck) label(ctx context.Context, client *github.Client, req types.IssueCommentOuter, cmdType, labelValue string) error {
	labels, err := d.allLabels(ctx, client, req)
	if err != nil {
		return err
	}
	if found := labelDuplicate(labels, labelValue); !found {
		return fmt.Errorf("label %s does not exist", labelValue)
	}

	if cmdType == addLabelConst {
		_, _, err = client.Issues.AddLabelsToIssue(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, []string{labelValue})
	} else {
		_, err = client.Issues.RemoveLabelForIssue(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, labelValue)
	}
	return err
}

func (d Dreck) title(ctx context.Context, client *github.Client, req types.IssueCommentOuter, cmdType, cmdValue string) error {
	newTitle := cmdValue
	if newTitle == req.Issue.Title || len(newTitle) == 0 {
		return fmt.Errorf("setting the title of #%d by %s was unsuccessful as the new title was empty or unchanged", req.Issue.Number, req.Comment.User.Login)
	}

	input := &github.IssueRequest{Title: &newTitle}
	_, _, err := client.Issues.Edit(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, input)
	return err
}

func (d Dreck) assign(ctx context.Context, client *github.Client, req types.IssueCommentOuter, cmdType, cmdValue string) error {
	if len(cmdValue) > 1 && cmdValue[0] == '@' {
		cmdValue = cmdValue[1:]
	}

	if cmdValue == "me" || cmdValue == "" {
		cmdValue = req.Comment.User.Login
	}

	if cmdType == unassignConst {
		_, _, err := client.Issues.RemoveAssignees(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, []string{cmdValue})
		return err
	} else {
		_, _, err := client.Issues.AddAssignees(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, []string{cmdValue})
		return err
	}
	return nil
}

func (d Dreck) cc(ctx context.Context, client *github.Client, req types.IssueCommentOuter, cmdType, cmdValue string) error {
	if len(cmdValue) > 1 && cmdValue[0] == '@' {
		cmdValue = cmdValue[1:]
	}

	if cmdValue == "me" || cmdValue == "" {
		cmdValue = req.Comment.User.Login
	}

	number := req.PullRequest.Number
	if number == 0 {
		number = req.Issue.Number
	}

	// check if this a pull request.
	_, _, err := client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, number)
	if err != nil {
		return fmt.Errorf("not a pull request: %d", number)
	}

	rev := github.ReviewersRequest{Reviewers: []string{cmdValue}}
	if cmdType == ccConst {
		_, _, err := client.PullRequests.RequestReviewers(ctx, req.Repository.Owner.Login, req.Repository.Name, number, rev)
		return err
	} else {
		_, err := client.PullRequests.RemoveReviewers(ctx, req.Repository.Owner.Login, req.Repository.Name, number, rev)
		return err
	}
	return nil
}

func (d Dreck) state(ctx context.Context, client *github.Client, req types.IssueCommentOuter, cmdType string) error {
	newState, validTransition := checkTransition(cmdType, req.Issue.State)
	if !validTransition {
		return fmt.Errorf("request to %s issue #%d by %s was invalid", cmdType, req.Issue.Number, req.Comment.User.Login)
	}

	input := &github.IssueRequest{State: &newState}
	_, _, err := client.Issues.Edit(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, input)
	return err
}

func (d Dreck) lock(ctx context.Context, client *github.Client, req types.IssueCommentOuter, cmdType string) error {
	if !isAction(req.Issue.Locked, cmdType, lockConst, unlockConst) {
		return fmt.Errorf("issue #%d is already %sed", req.Issue.Number, cmdType)
	}

	if cmdType == lockConst {
		_, err := client.Issues.Lock(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, &github.LockIssueOptions{})
		return err
	} else {
		_, err := client.Issues.Unlock(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
		return err
	}
	return nil
}

func (d Dreck) lgtm(ctx context.Context, client *github.Client, req types.IssueCommentOuter, cmdType string) error {
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
	if cmdType == lgtmConst {
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

func (d Dreck) test(ctx context.Context, _ *github.Client, _ types.IssueCommentOuter, _, _ string) error {
	return nil
}

func (d Dreck) duplicate(ctx context.Context, client *github.Client, req types.IssueCommentOuter, cmdType, cmdValue string) error {
	if err := d.label(ctx, client, req, addLabelConst, "duplicate"); err != nil {
		return err
	}
	return d.state(ctx, client, req, closeConst)
}

// Body must be downcased already.
func parse(body string, conf *types.DreckConfig) []*types.CommentAction {
	actions := []*types.CommentAction{}

	for trigger, commandType := range IssueCommands {
		if val := isValidCommand(body, trigger, conf); len(val) > 0 {
			for _, v := range val {
				actions = append(actions, &types.CommentAction{Type: commandType, Value: v})
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
	Trigger + "label":      addLabelConst,
	Trigger + "unlabel":    removeLabelConst,
	Trigger + "cc":         ccConst,
	Trigger + "uncc":       unccConst,
	Trigger + "assign":     assignConst,
	Trigger + "unassign":   unassignConst,
	Trigger + "close":      closeConst,
	Trigger + "reopen":     reopenConst,
	Trigger + "title":      titleConst,
	Trigger + "lock":       lockConst,
	Trigger + "unlock":     unlockConst,
	Trigger + "exec":       execConst,
	Trigger + "lgtm":       lgtmConst,   // Only works on Pull Request comments.
	Trigger + "unlgtm":     unlgtmConst, // Only works on Pull Request comments.
	Trigger + "merge":      mergeConst,  // Only works on Pull Request comments.
	Trigger + "fortune":    fortuneConst,
	Trigger + "test":       testConst,
	Trigger + "duplicate ": duplicateConst,
}
