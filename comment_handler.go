package dreck

import (
	"bufio"
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
	execConst      = "exec"
	testConst      = "test"
	duplicateConst = "duplicate"
	mergeConst     = "merge"
	fortuneConst   = "fortune"
)

func (d Dreck) comment(req types.IssueCommentOuter, conf *types.DreckConfig) error {
	body := strings.ToLower(req.Comment.Body)
	c := parse(body, conf)

	for _, command := range c {
		log.Infof("Incoming request from %s, %s: %s", req.Comment.User.Login, command.Type, command.Value)
		switch command.Type {
		case addLabelConst, removeLabelConst:
			if isMe(req.Comment.User.Login, command.Value) || isCodeOwner(conf, req.Comment.User.Login) {
				return d.label(req, command.Type, command.Value)
			}
			return fmt.Errorf("user %s not permitted to use [un]label", req.Comment.User.Login)
		case assignConst, unassignConst:
			if isMe(req.Comment.User.Login, command.Value) || isCodeOwner(conf, req.Comment.User.Login) {
				return d.assign(req, command.Type, command.Value)
			}
			return fmt.Errorf("user %s not permitted to use [un]assign", req.Comment.User.Login)
		case closeConst, reopenConst:
			return d.state(req, command.Type)
		case titleConst:
			return d.title(req, command.Type, command.Value)
		case lockConst, unlockConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				return d.lock(req, command.Type)
			}
			return fmt.Errorf("user %s not permitted to use [un]lock", req.Comment.User.Login)
		case lgtmConst, unlgtmConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				return d.lgtm(req, command.Type)
			}
			return fmt.Errorf("user %s not permitted to use [un]lgtm", req.Comment.User.Login)
		case ccConst, unccConst:
			if isMe(req.Comment.User.Login, command.Value) || isCodeOwner(conf, req.Comment.User.Login) {
				//return d.cc(req, command.Type, command.Value)
				return nil
			}
			return fmt.Errorf("user %s not permitted to use [un]cc", req.Comment.User.Login)
		case testConst:
			if err := d.test(req, command.Type, command.Value); err != nil {
				return err
			}
		case duplicateConst:
			return d.duplicate(req, command.Type, command.Value)
		case fortuneConst:
			return d.fortune(req, command.Type)
		case execConst:
			if !aliasOK(conf) {
				return fmt.Errorf("feature %s is not enabled, so %s can't work", Trigger+execConst, Aliases)
			}
			if !isCodeOwner(conf, req.Comment.User.Login) {
				return fmt.Errorf("user %s not permitted to use exec", req.Comment.User.Login)
			}
			return d.exec(req, conf, command.Type, command.Value)
		case mergeConst:
			if isCodeOwner(conf, req.Comment.User.Login) {
				return d.merge(req)
			}
			return fmt.Errorf("user %s is not a code owner", req.Comment.User.Login)
		}
	}

	if len(c) == 0 {
		log.Infof("No command found in comment %d", req.Issue.Number)
	}
	return nil
}

func (d Dreck) label(req types.IssueCommentOuter, cmdType, labelValue string) error {
	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

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

	if err != nil {
		return err
	}

	return nil
}

func (d Dreck) title(req types.IssueCommentOuter, cmdType, cmdValue string) error {
	newTitle := cmdValue
	if newTitle == req.Issue.Title || len(newTitle) == 0 {
		return fmt.Errorf("setting the title of #%d by %s was unsuccessful as the new title was empty or unchanged", req.Issue.Number, req.Comment.User.Login)
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	input := &github.IssueRequest{Title: &newTitle}

	if _, _, err := client.Issues.Edit(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, input); err != nil {
		return err
	}
	return nil
}

func (d Dreck) assign(req types.IssueCommentOuter, cmdType, cmdValue string) error {
	if len(cmdValue) > 1 && cmdValue[0] == '@' {
		cmdValue = cmdValue[1:]
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	if cmdValue == "me" || cmdValue == "" {
		cmdValue = req.Comment.User.Login
	}

	if cmdType == unassignConst {
		_, _, err = client.Issues.RemoveAssignees(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, []string{cmdValue})
	} else {
		_, _, err = client.Issues.AddAssignees(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, []string{cmdValue})
	}

	if err != nil {
		return err
	}
	return nil
}

func (d Dreck) state(req types.IssueCommentOuter, cmdType string) error {
	newState, validTransition := checkTransition(cmdType, req.Issue.State)
	if !validTransition {
		return fmt.Errorf("request to %s issue #%d by %s was invalidn", cmdType, req.Issue.Number, req.Comment.User.Login)
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}
	input := &github.IssueRequest{State: &newState}

	if _, _, err := client.Issues.Edit(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, input); err != nil {
		return err
	}
	return nil
}

func (d Dreck) lock(req types.IssueCommentOuter, cmdType string) error {
	if !isAction(req.Issue.Locked, cmdType, lockConst, unlockConst) {
		return fmt.Errorf("issue #%d is already %sed", req.Issue.Number, cmdType)
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	if cmdType == lockConst {
		_, err = client.Issues.Lock(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, &github.LockIssueOptions{})
	} else {
		_, err = client.Issues.Unlock(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
	}

	if err != nil {
		return err
	}

	return nil
}

func (d Dreck) lgtm(req types.IssueCommentOuter, cmdType string) error {
	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	_, _, err = client.PullRequests.Get(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
	// will be 404 not found if this isn't a PR.
	if err != nil {
		return err
	}

	input := &github.PullRequestReviewRequest{
		Body:  github.String("LGTM by **" + req.Comment.User.Login + "**"),
		Event: github.String("APPROVE"),
	}

	_, _, err = client.PullRequests.CreateReview(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, input)

	return err
}

func (d Dreck) test(req types.IssueCommentOuter, cmdType, cmdValue string) error {
	return nil
}

func (d Dreck) duplicate(req types.IssueCommentOuter, cmdType, cmdValue string) error {
	if err := d.label(req, addLabelConst, "duplicate"); err != nil {
		return err
	}
	if err := d.state(req, closeConst); err != nil {
		return err
	}
	return nil
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
// Body myst be lowercased.
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
		if line == trigger+"\n" {
			val = append(val, "")
			continue
		}
		if !strings.HasPrefix(line, trigger+" ") {
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
