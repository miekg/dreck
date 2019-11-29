package dreck

import (
	"fmt"
	"testing"

	"github.com/miekg/dreck/types"
)

func TestAlias(t *testing.T) {
	alias := "/plugin (.*) -> /label plugin/$1"

	r, err := NewAlias(alias)
	if err != nil {
		t.Errorf("Failed to parse %s: %v", alias, err)
	}
	input := "/plugin example"
	exp := r.Expand(input)
	if exp == input {
		t.Errorf("Failed to expand %s", input)
	}

	t.Logf("Got %s\n", exp)
}

func TestAliasSingle(t *testing.T) {
	alias := "/approve -> /lgtm"

	r, err := NewAlias(alias)
	if err != nil {
		t.Errorf("Failed to parse %s: %v", alias, err)
	}
	input := "/approve"
	exp := r.Expand(input)
	if exp == input {
		t.Errorf("Failed to expand %s", input)
	}

	t.Logf("Got %s\n", exp)
}

func TestAliasParse(t *testing.T) {
	alias := "/plugin (.*) - /label plugin/$1"
	if _, err := NewAlias(alias); err == nil {
		t.Errorf("Expected to not parse %s", alias)
	}

	alias = "/plugin (*) - /label plugin/$1"
	if _, err := NewAlias(alias); err == nil {
		t.Errorf("Expected to not parse %s", alias)
	}
}

func TestParsingAlias(t *testing.T) {
	conf := &types.DreckConfig{
		Features: []string{Aliases},
		Aliases: []string{
			fmt.Sprintf("%splugin (.*) -> %slabel plugin/$1", Trigger, Trigger),
			fmt.Sprintf("%splugin2 (.*) -> %slabel plugin/$2", Trigger, Trigger),
			fmt.Sprintf("%slooksOK -> %slgtm", Trigger, Trigger),
		},
	}

	var options = []struct {
		title        string
		body         string
		expectedType string
		expectedVal  string
	}{
		{
			title:        "Alias Add label of demo",
			body:         Trigger + "plugin demo",
			expectedType: "AddLabel",
			expectedVal:  "plugin/demo",
		},
		{
			title:        "Alias2 Add label of demo",
			body:         Trigger + "plugin2 demo",
			expectedType: "AddLabel",
			expectedVal:  "plugin/",
		},
		{
			title:        "Alias Add label of demo",
			body:         Trigger + "plugin demo",
			expectedType: "AddLabel",
			expectedVal:  "plugin/demo",
		},
		{
			title:        "Alias Add label of demo case",
			body:         Trigger + "plUGin demo",
			expectedType: "AddLabel",
			expectedVal:  "plugin/demo",
		},
		{
			title:        "Non alias label of demo",
			body:         Trigger + "pluginner demo",
			expectedType: "",
			expectedVal:  "",
		},
		{
			title:        "Lgtm",
			body:         Trigger + "looksOK",
			expectedType: "lgtm",
			expectedVal:  "",
		},
	}

	for _, test := range options {
		t.Run(test.title, func(t *testing.T) {

			actions := parse(test.body, conf)
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
				t.Errorf("Action - wanted: %s, got %s\nLabel - wanted: %s, got %s", test.expectedType, action.Type, test.expectedVal, action.Value)
			}
		})
	}
}
