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
	setTitleConst    = "SetTitle"
	assignConst      = "Assign"
	unassignConst    = "Unassign"
	removeLabelConst = "RemoveLabel"
	addLabelConst    = "AddLabel"
	lgtmConst        = "lgtm"
	autosubmitConst  = "autosubmit"
	execConst        = "exec"
	testConst        = "test"
)

func (d Dreck) comment(req types.IssueCommentOuter, conf *types.DreckConfig) error {
	body := strings.ToLower(req.Comment.Body)
	c := parse(body, conf)

	for _, command := range c {

		switch command.Type {

		case addLabelConst, removeLabelConst:
			if err := d.label(req, command.Type, command.Value); err != nil {
				return err
			}
		case assignConst, unassignConst:
			if err := d.assign(req, command.Type, command.Value); err != nil {
				return err
			}
		case closeConst, reopenConst:
			if err := d.state(req, command.Type); err != nil {
				return err
			}
		case setTitleConst:
			if err := d.title(req, command.Type, command.Value); err != nil {
				return err
			}
		case lockConst, unlockConst:
			if err := d.lock(req, command.Type); err != nil {
				return err
			}
		case lgtmConst:
			if err := d.lgtm(req, command.Type); err != nil {
				return err
			}
		case testConst:
			if err := d.test(req, command.Type, command.Value); err != nil {
				return err
			}
		case autosubmitConst:
			if permittedUserFeature(featureAutosubmit, conf, req.Comment.User.Login) {
				if err := d.autosubmit(req); err != nil {
					return err
				}
			}
			return fmt.Errorf("user %s not permitted to use %s or this feature is disabled", req.Comment.User.Login, autosubmitConst)
		case execConst:
			if !enabledFeature(featureAliases, conf) {
				return fmt.Errorf("feature %s is not enabled, so %s can't work", Trigger+execConst, featureAliases)
			}
			if !permittedUserFeature(featureExec, conf, req.Comment.User.Login) {
				return fmt.Errorf("user %s not permitted to use %s or this feature is disabled", req.Comment.User.Login, execConst)
			}

			if err := d.exec(req, conf, command.Type, command.Value); err != nil {
				return err
			}
		}
	}

	if len(c) == 0 {
		log.Warningf("No command found in comment %d", req.Issue.Number)
	}
	return nil
}

func (d Dreck) label(req types.IssueCommentOuter, cmdType, labelValue string) error {

	labelAction := strings.Replace(cmdType, "label", "", 1)

	log.Infof("%s wants to %s label of '%s' on issue #%d \n", req.Comment.User.Login, labelAction, labelValue, req.Issue.Number)

	found := labelDuplicate(req.Issue.Labels, labelValue)
	if !validAction(found, cmdType, addLabelConst, removeLabelConst) {
		return fmt.Errorf("request to %s label of '%s' on issue #%d was unnecessary", labelAction, labelValue, req.Issue.Number)
	}

	client, ctx, err := d.newClient(req.Installation.ID)
	if err != nil {
		return err
	}

	labels, err := d.allLabels(ctx, client, req)
	if err != nil {
		return err
	}
	found = labelDuplicate(labels, labelValue)
	if !found {
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

	log.Infof("Request to %s label of '%s' on issue #%d was successfully completed.", labelAction, labelValue, req.Issue.Number)

	return nil
}

func (d Dreck) title(req types.IssueCommentOuter, cmdType, cmdValue string) error {

	log.Infof("%s wants to set the title of issue #%d\n", req.Comment.User.Login, req.Issue.Number)

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

	log.Infof("Request to set the title of issue #%d by %s was successful.\n", req.Issue.Number, req.Comment.User.Login)
	return nil
}

func (d Dreck) assign(req types.IssueCommentOuter, cmdType, cmdValue string) error {

	log.Infof("%s wants to %s user '%s' from issue #%d\n", req.Comment.User.Login, cmdType, cmdValue, req.Issue.Number)

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

	log.Infof("%s %sed successfully or already %sed.\n", cmdValue, cmdType, cmdType)

	return nil
}

func (d Dreck) state(req types.IssueCommentOuter, cmdType string) error {

	log.Infof("%s wants to %s issue #%d\n", req.Comment.User.Login, cmdType, req.Issue.Number)

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

	log.Infof("Request to %s issue #%d by %s was successful.\n", cmdType, req.Issue.Number, req.Comment.User.Login)

	return nil

}

func (d Dreck) lock(req types.IssueCommentOuter, cmdType string) error {

	log.Infof("%s wants to %s issue #%d\n", req.Comment.User.Login, cmdType, req.Issue.Number)

	if !validAction(req.Issue.Locked, cmdType, lockConst, unlockConst) {
		return fmt.Errorf("issue #%d is already %sed", req.Issue.Number, cmdType)
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

	log.Infof("Request to %s issue #%d by %s was successful.\n", cmdType, req.Issue.Number, req.Comment.User.Login)
	return nil
}

func (d Dreck) lgtm(req types.IssueCommentOuter, cmdType string) error {
	log.Infof("%s wants to %s pull request #%d\n", req.Comment.User.Login, cmdType, req.Issue.Number)

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

func (d Dreck) test(req types.IssueCommentOuter, cmdType, cmdValue string) error {
	log.Infof("%s wants to %s %s issue #%d\n", req.Comment.User.Login, cmdType, cmdValue, req.Issue.Number)
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
	if ok := enabledFeature(featureAliases, conf); ok {
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
		if !strings.HasPrefix(line, trigger) {
			continue
		}
		// rest of the line is the value.
		v := line[len(trigger):]
		v = strings.Trim(v, " \t.,\n\r")
		val = append(val, v)
	}

	return val
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
	Trigger + "exec":           execConst,
	Trigger + "lgtm":           lgtmConst,       // Only works on Pull Requests comments.
	Trigger + "autosubmit":     autosubmitConst, // Only works on Pull Request comments.
	Trigger + "test: ":         testConst,
}
