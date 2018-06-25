package dreck

import (
	"strings"
	"testing"

	"github.com/miekg/dreck/types"
)

var actionOptions = []struct {
	title          string
	body           string
	expectedAction string
}{
	{
		title:          "Correct reopen command",
		body:           Trigger + "reopen",
		expectedAction: reopenConst,
	},
	{
		title:          "Correct close command",
		body:           Trigger + "close",
		expectedAction: closeConst,
	},
	{
		title:          "Uppercase close command",
		body:           Trigger + "cLOse",
		expectedAction: closeConst,
	},
	{
		title:          "invalid command",
		body:           Trigger + "dance",
		expectedAction: "",
	},
	{
		title:          "Longer reopen command",
		body:           Trigger + "reopen: ",
		expectedAction: reopenConst,
	},
	{
		title:          "Longer close command",
		body:           Trigger + "close: ",
		expectedAction: closeConst,
	},
}

func TestParsingOpenClose(t *testing.T) {

	for _, test := range actionOptions {
		t.Run(test.title, func(t *testing.T) {
			test.body = strings.ToLower(test.body)
			actions := parse(test.body, &types.DreckConfig{})
			if len(actions) != 1 {
				t.Errorf("Action - not parsed correctly")
				return
			}
			action := actions[0]

			if action.Type != test.expectedAction {
				t.Errorf("Action - want: %s, got %s", test.expectedAction, action.Type)
			}

		})
	}
}

func TestParsingLabels(t *testing.T) {

	var labelOptions = []struct {
		title        string
		body         string
		expectedType string
		expectedVal  string
	}{
		{
			title:        "Add label of demo",
			body:         Trigger + "label add: demo",
			expectedType: "AddLabel",
			expectedVal:  "demo",
		},
		{
			title:        "Remove label of demo",
			body:         Trigger + "label remove: demo",
			expectedType: "RemoveLabel",
			expectedVal:  "demo",
		},
		{
			title:        "Invalid label action",
			body:         Trigger + "label peel: demo",
			expectedType: "",
			expectedVal:  "",
		},
	}

	for _, test := range labelOptions {
		t.Run(test.title, func(t *testing.T) {

			actions := parse(test.body, &types.DreckConfig{})
			if len(actions) != 1 {
				t.Errorf("Action - not parsed correctly")
				return
			}
			action := actions[0]
			if action.Type != test.expectedType || action.Value != test.expectedVal {
				t.Errorf("Action - wanted: %s, got %s\nLabel - wanted: %s, got %s", test.expectedType, action.Type, test.expectedVal, action.Value)
			}
		})
	}
}

func TestParsingAssignments(t *testing.T) {

	var assignmentOptions = []struct {
		title        string
		body         string
		expectedType string
		expectedVal  string
	}{
		{
			title:        "Assign to burt",
			body:         Trigger + "assign: burt",
			expectedType: assignConst,
			expectedVal:  "burt",
		},
		{
			title:        "Unassign burt",
			body:         Trigger + "unassign: burt",
			expectedType: unassignConst,
			expectedVal:  "burt",
		},
		{
			title:        "Assign to me",
			body:         Trigger + "assign: me",
			expectedType: assignConst,
			expectedVal:  "me",
		},
		{
			title:        "Unassign me",
			body:         Trigger + "unassign: me",
			expectedType: unassignConst,
			expectedVal:  "me",
		},
		{
			title:        "Invalid assignment action",
			body:         Trigger + "consign: burt",
			expectedType: "",
			expectedVal:  "",
		},
		{
			title:        "Unassign blank",
			body:         Trigger + "unassign: ",
			expectedType: unassignConst,
			expectedVal:  "",
		},
	}

	for _, test := range assignmentOptions {
		t.Run(test.title, func(t *testing.T) {

			actions := parse(test.body, &types.DreckConfig{})
			if len(actions) == 0 && test.expectedType == "" { // Ugly hack to should be cleaned up (miek)
				// correct, we didn't parse anything
				return
			}
			if len(actions) != 1 {
				t.Errorf("Action - not parsed correctly")
				return
			}
			action := actions[0]
			if action.Type != test.expectedType || action.Value != test.expectedVal {
				t.Errorf("Action - wanted: %s, got %s\nMaintainer - wanted: %s, got %s", test.expectedType, action.Type, test.expectedVal, action.Value)
			}
		})
	}
}

func TestParsingTitles(t *testing.T) {

	var titleOptions = []struct {
		title        string
		body         string
		expectedType string
		expectedVal  string
	}{
		{
			title:        "Set Title",
			body:         Trigger + "title set: This is a really great Title!",
			expectedType: setTitleConst,
			expectedVal:  "This is a really great Title!",
		},
		{
			title:        "Mis-spelling of title",
			body:         Trigger + "titel set: This is a really great Title!",
			expectedType: "",
			expectedVal:  "",
		},
		{
			title:        "Empty Title",
			body:         Trigger + "title set: ",
			expectedType: setTitleConst,
			expectedVal:  "",
		},
		{
			title:        "Empty Title (Double Space)",
			body:         Trigger + "title set:  ",
			expectedType: setTitleConst,
			expectedVal:  "",
		},
	}

	for _, test := range titleOptions {
		t.Run(test.title, func(t *testing.T) {

			actions := parse(test.body, &types.DreckConfig{})
			if len(actions) == 0 && test.expectedType == "" { // Ugly hack to should be cleaned up (miek)
				// correct, we didn't parse anything
				return
			}

			if len(actions) != 1 {
				t.Errorf("Action - not parsed correctly")
				return
			}
			action := actions[0]
			if action.Type != test.expectedType || action.Value != test.expectedVal {
				t.Errorf("\nAction - wanted: %q, got %q\nValue - wanted: %q, got %q", test.expectedType, action.Type, test.expectedVal, action.Value)
			}
		})
	}
}

func TestAssessState(t *testing.T) {

	var stateOptions = []struct {
		title            string
		requestedAction  string
		currentState     string
		expectedNewState string
		expectedBool     bool
	}{
		{
			title:            "Currently Closed and trying to close",
			requestedAction:  closeConst,
			currentState:     closedConst,
			expectedNewState: "",
			expectedBool:     false,
		},
		{
			title:            "Currently Open and trying to reopen",
			requestedAction:  reopenConst,
			currentState:     openConst,
			expectedNewState: "",
			expectedBool:     false,
		},
		{
			title:            "Currently Closed and trying to open",
			requestedAction:  reopenConst,
			currentState:     closedConst,
			expectedNewState: openConst,
			expectedBool:     true,
		},
		{
			title:            "Currently Open and trying to close",
			requestedAction:  closeConst,
			currentState:     openConst,
			expectedNewState: closedConst,
			expectedBool:     true,
		},
	}

	for _, test := range stateOptions {
		t.Run(test.title, func(t *testing.T) {

			newState, validTransition := checkTransition(test.requestedAction, test.currentState)

			if newState != test.expectedNewState || validTransition != test.expectedBool {
				t.Errorf("\nStates - wanted: %s, got %s\nValidity - wanted: %t, got %t\n", test.expectedNewState, newState, test.expectedBool, validTransition)
			}
		})
	}
}

func TestValidAction(t *testing.T) {

	var stateOptions = []struct {
		title           string
		running         bool
		requestedAction string
		start           string
		stop            string
		expectedBool    bool
	}{
		{
			title:           "Currently unlocked and trying to lock",
			running:         false,
			requestedAction: lockConst,
			start:           lockConst,
			stop:            unlockConst,
			expectedBool:    true,
		},
		{
			title:           "Currently unlocked and trying to unlock",
			running:         false,
			requestedAction: unlockConst,
			start:           lockConst,
			stop:            unlockConst,
			expectedBool:    false,
		},
		{
			title:           "Currently locked and trying to lock",
			running:         true,
			requestedAction: lockConst,
			start:           lockConst,
			stop:            unlockConst,
			expectedBool:    false,
		},
		{
			title:           "Currently locked and trying to unlock",
			running:         true,
			requestedAction: unlockConst,
			start:           lockConst,
			stop:            unlockConst,
			expectedBool:    true,
		},
	}

	for _, test := range stateOptions {
		t.Run(test.title, func(t *testing.T) {

			isValid := validAction(test.running, test.requestedAction, test.start, test.stop)

			if isValid != test.expectedBool {
				t.Errorf("\nActions - wanted: %t, got %t\n", test.expectedBool, isValid)
			}
		})
	}
}

func TestLabelDuplicate(t *testing.T) {

	var stateOptions = []struct {
		title         string
		currentLabels []types.IssueLabel
		cmdLabel      string
		expectedFound bool
	}{
		{
			title: "Label exists lowercase",
			currentLabels: []types.IssueLabel{
				types.IssueLabel{
					Name: "rod",
				},
				types.IssueLabel{
					Name: "jane",
				},
				types.IssueLabel{
					Name: "freddie",
				},
			},
			cmdLabel:      "rod",
			expectedFound: true,
		},
		{
			title: "Label exists case insensitive",
			currentLabels: []types.IssueLabel{
				types.IssueLabel{
					Name: "rod",
				},
				types.IssueLabel{
					Name: "jane",
				},
				types.IssueLabel{
					Name: "freddie",
				},
			},
			cmdLabel:      "Rod",
			expectedFound: true,
		},
		{
			title: "Label doesnt exist lowercase",
			currentLabels: []types.IssueLabel{
				types.IssueLabel{
					Name: "rod",
				},
				types.IssueLabel{
					Name: "jane",
				},
				types.IssueLabel{
					Name: "freddie",
				},
			},
			cmdLabel:      "derek",
			expectedFound: false,
		},
		{
			title: "Label doesnt exist case insensitive",
			currentLabels: []types.IssueLabel{
				types.IssueLabel{
					Name: "rod",
				},
				types.IssueLabel{
					Name: "jane",
				},
				types.IssueLabel{
					Name: "freddie",
				},
			},
			cmdLabel:      "Derek",
			expectedFound: false,
		},
		{
			title:         "no existing labels lowercase",
			currentLabels: nil,
			cmdLabel:      "derek",
			expectedFound: false,
		},
		{title: "Label doesnt exist case insensitive",
			currentLabels: nil,
			cmdLabel:      "Derek",
			expectedFound: false,
		},
	}

	for _, test := range stateOptions {
		t.Run(test.title, func(t *testing.T) {

			labelFound := labelDuplicate(test.currentLabels, test.cmdLabel)

			if labelFound != test.expectedFound {
				t.Errorf("Find Labels(%s) - wanted: %t, got %t\n", test.title, test.expectedFound, labelFound)
			}
		})
	}
}
