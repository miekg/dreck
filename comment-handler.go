package dreck

import (
	"strings"

	"github.com/miekg/dreck/log"
	"github.com/miekg/dreck/types"

	"github.com/google/go-github/github"
)

const (
	openConst        string = "open"
	openPRConst      string = "opened"
	closedConst      string = "closed"
	closeConst       string = "close"
	reopenConst      string = "reopen"
	lockConst        string = "Lock"
	unlockConst      string = "Unlock"
	setTitleConst    string = "SetTitle"
	assignConst      string = "Assign"
	unassignConst    string = "Unassign"
	removeLabelConst string = "RemoveLabel"
	addLabelConst    string = "AddLabel"
)

func (d Dreck) handleComment(req types.IssueCommentOuter) (err error) {
	command := parse(req.Comment.Body)

	switch command.Type {

	case addLabelConst, removeLabelConst:

		err = d.manageLabel(req, command.Type, command.Value)

	case assignConst, unassignConst:

		err = d.manageAssignment(req, command.Type, command.Value)

	case closeConst, reopenConst:

		err = d.manageState(req, command.Type)

	case setTitleConst:

		err = d.manageTitle(req, command.Type, command.Value)

	case lockConst, unlockConst:

		err = d.manageLocking(req, command.Type)

	default:
		log.Warningf("Unable to work with comment: %s" + req.Comment.Body)
		return nil
	}

	return err
}

func findLabel(currentLabels []types.IssueLabel, cmdLabel string) bool {

	for _, label := range currentLabels {
		if strings.EqualFold(label.Name, cmdLabel) {
			return true
		}
	}
	return false
}

func (d Dreck) manageLabel(req types.IssueCommentOuter, cmdType string, labelValue string) error {

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

func (d Dreck) manageTitle(req types.IssueCommentOuter, cmdType string, cmdValue string) error {

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

func (d Dreck) manageAssignment(req types.IssueCommentOuter, cmdType string, cmdValue string) error {

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

func (d Dreck) manageState(req types.IssueCommentOuter, cmdType string) error {

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

func (d Dreck) manageLocking(req types.IssueCommentOuter, cmdType string) error {

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

func parse(body string) *types.CommentAction {
	for trigger, commandType := range IssueCommands {

		if isValidCommand(body, trigger) {
			val := body[len(trigger):]
			val = strings.Trim(val, " \t.,\n\r")

			return &types.CommentAction{Type: commandType, Value: val}
		}
	}

	return &types.CommentAction{}
}

func isValidCommand(body string, trigger string) bool {
	return (len(body) > len(trigger) && body[0:len(trigger)] == trigger) ||
		(body == trigger && !strings.HasSuffix(trigger, ": "))
}

func validAction(running bool, requestedAction string, start string, stop string) bool {
	return !running && requestedAction == start || running && requestedAction == stop
}

func checkTransition(requestedAction string, currentState string) (string, bool) {
	if requestedAction == closeConst && currentState != closedConst {
		return closedConst, true
	} else if requestedAction == reopenConst && currentState != openConst {
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
}
