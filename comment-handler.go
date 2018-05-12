package dreck

import (
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
	setTitleConst    = "SetTitle"
	assignConst      = "Assign"
	unassignConst    = "Unassign"
	removeLabelConst = "RemoveLabel"
	addLabelConst    = "AddLabel"
	lgtmConst        = "lgtm"
)

func (d Dreck) comment(req types.IssueCommentOuter, conf *types.DreckConfig) error {
	command := parse(req.Comment.Body, conf)

	switch command.Type {

	case addLabelConst, removeLabelConst:
		return d.label(req, command.Type, command.Value)
	case assignConst, unassignConst:
		return d.assign(req, command.Type, command.Value)
	case closeConst, reopenConst:
		return d.state(req, command.Type)
	case setTitleConst:
		return d.title(req, command.Type, command.Value)
	case lockConst, unlockConst:
		return d.lock(req, command.Type)
	case lgtmConst:
		return d.lgtm(req, command.Type)
	}

	if len(req.Comment.Body) > 25 {
		log.Warningf("Unable to work with comment: %s", req.Comment.Body[:25])
	} else {
		log.Warningf("Unable to work with comment: %s", req.Comment.Body)
	}
	return nil
}

func findLabel(currentLabels []types.IssueLabel, cmdLabel string) bool {

	for _, label := range currentLabels {
		if strings.EqualFold(label.Name, cmdLabel) {
			return true
		}
	}
	return false
}

func (d Dreck) label(req types.IssueCommentOuter, cmdType string, labelValue string) error {

	labelAction := strings.Replace(strings.ToLower(cmdType), "label", "", 1)

	log.Infof("%s wants to %s label of '%s' on issue #%d \n", req.Comment.User.Login, labelAction, labelValue, req.Issue.Number)

	found := findLabel(req.Issue.Labels, labelValue)

	if !validAction(found, cmdType, addLabelConst, removeLabelConst) {
		log.Errorf("Request to %s label of '%s' on issue #%d was unnecessary.", labelAction, labelValue, req.Issue.Number)
		return nil
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	if cmdType == addLabelConst {
		_, _, err = client.Issues.AddLabelsToIssue(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, []string{labelValue})
	} else {
		_, err = client.Issues.RemoveLabelForIssue(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, labelValue)
	}

	if err != nil {
		return err
	}

	log.Infof("Request to %s label of '%s' on issue #%d was successfully completed.", labelAction, labelValue, req.Issue.Number)

	return nil
}

func (d Dreck) title(req types.IssueCommentOuter, cmdType string, cmdValue string) error {

	log.Infof("%s wants to set the title of issue #%d\n", req.Comment.User.Login, req.Issue.Number)

	newTitle := cmdValue

	if newTitle == req.Issue.Title || len(newTitle) == 0 {
		log.Errorf("Setting the title of #%d by %s was unsuccessful as the new title was empty or unchanged.\n", req.Issue.Number, req.Comment.User.Login)
		return nil
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	input := &github.IssueRequest{Title: &newTitle}

	if _, _, err := client.Issues.Edit(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, input); err != nil {
		return err
	}

	log.Infof("Request to set the title of issue #%d by %s was successful.\n", req.Issue.Number, req.Comment.User.Login)
	return nil
}

func (d Dreck) assign(req types.IssueCommentOuter, cmdType string, cmdValue string) error {

	log.Infof("%s wants to %s user '%s' from issue #%d\n", req.Comment.User.Login, strings.ToLower(cmdType), cmdValue, req.Issue.Number)

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	if cmdValue == "me" {
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

	log.Infof("%s %sed successfully or already %sed.\n", cmdValue, strings.ToLower(cmdType), strings.ToLower(cmdType))

	return nil
}

func (d Dreck) state(req types.IssueCommentOuter, cmdType string) error {

	log.Infof("%s wants to %s issue #%d\n", req.Comment.User.Login, cmdType, req.Issue.Number)

	newState, validTransition := checkTransition(cmdType, req.Issue.State)

	if !validTransition {
		log.Errorf("Request to %s issue #%d by %s was invalid.\n", cmdType, req.Issue.Number, req.Comment.User.Login)
		return nil
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}
	input := &github.IssueRequest{State: &newState}

	if _, _, err := client.Issues.Edit(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, input); err != nil {
		return err
	}

	log.Infof("Request to %s issue #%d by %s was successful.\n", cmdType, req.Issue.Number, req.Comment.User.Login)

	return nil

}

func (d Dreck) lock(req types.IssueCommentOuter, cmdType string) error {

	log.Infof("%s wants to %s issue #%d\n", req.Comment.User.Login, strings.ToLower(cmdType), req.Issue.Number)

	if !validAction(req.Issue.Locked, cmdType, lockConst, unlockConst) {
		log.Errorf("Issue #%d is already %sed\n", req.Issue.Number, strings.ToLower(cmdType))
		return nil
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	if cmdType == lockConst {
		_, err = client.Issues.Lock(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
	} else {
		_, err = client.Issues.Unlock(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
	}

	if err != nil {
		return err
	}

	log.Infof("Request to %s issue #%d by %s was successful.\n", strings.ToLower(cmdType), req.Issue.Number, req.Comment.User.Login)
	return nil
}

func (d Dreck) lgtm(req types.IssueCommentOuter, cmdType string) error {
	log.Infof("%s wants to %s pull request #%d\n", req.Comment.User.Login, strings.ToLower(cmdType), req.Issue.Number)

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
		Body:  String("LGTM by **" + req.Comment.User.Login + "**"),
		Event: String("APPROVE"),
	}

	_, _, err = client.PullRequests.CreateReview(ctx, req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, input)

	return err
}

func parse(body string, conf *types.DreckConfig) *types.CommentAction {
	for trigger, commandType := range IssueCommands {
		if ok, val := isValidCommand(body, trigger, conf); ok {
			return &types.CommentAction{Type: commandType, Value: val}
		}
	}

	return &types.CommentAction{}
}

func isValidCommand(body string, trigger string, conf *types.DreckConfig) (bool, string) {
	for _, a := range conf.Aliases {
		r, err := NewAlias(a)
		if err != nil {
			log.Warningf("Failed to parse alias: %s, %v", a, err)
			continue
		}
		body = r.Expand(body) // either noop or replaces something
	}

	val := ""
	ok := (len(body) > len(trigger) && body[0:len(trigger)] == trigger) ||
		(body == trigger && !strings.HasSuffix(trigger, ": "))
	if ok {
		val = body[len(trigger):]
		val = strings.Trim(val, " \t.,\n\r")
	}
	return ok, val
}

func validAction(running bool, requestedAction string, start string, stop string) bool {
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
	Trigger + "label: ":        addLabelConst,
	Trigger + "label add: ":    addLabelConst,
	Trigger + "label remove: ": removeLabelConst,
	Trigger + "label rm: ":     removeLabelConst,
	Trigger + "assign: ":       assignConst,
	Trigger + "unassign: ":     unassignConst,
	Trigger + "close":          closeConst,
	Trigger + "reopen":         reopenConst,
	Trigger + "title: ":        setTitleConst,
	Trigger + "title set: ":    setTitleConst,
	Trigger + "title edit: ":   setTitleConst,
	Trigger + "lock":           lockConst,
	Trigger + "unlock":         unlockConst,
	Trigger + "lgtm":           lgtmConst, // Only works on Pull Requests comments.
}
