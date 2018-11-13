package dreck

import (
	"testing"

	"github.com/miekg/dreck/types"
)

func TestApproversConfigParse(t *testing.T) {
	config := types.DreckConfig{}
	parseConfig([]byte(`approvers:
- aa
- ac
`), &config)
	actual := len(config.Approvers)
	if actual != 2 {
		t.Errorf("want: %d approvers, got: %d", 2, actual)
	}
}

func TestReviewerConfigParse(t *testing.T) {
	config := types.DreckConfig{}
	parseConfig([]byte(`reviewers:
- aa
- ac 
`), &config)
	actual := len(config.Reviewers)
	if actual != 2 {
		t.Errorf("want: %d reviewers, got: %d", 2, actual)
	}
}

func TestAliasConfigParse(t *testing.T) {
	config := types.DreckConfig{}
	err := parseConfig([]byte(`aliases:
- |
  /plugin: (.*) - /label add: plugin/$1
`), &config)
	if err != nil {
		t.Errorf("failed to parse config: %s", err)
	}
	actual := len(config.Aliases)
	if actual != 1 {
		t.Errorf("want: %d aliases, got: %d", 1, actual)
	}
}

func TestConfigParse(t *testing.T) {
	config := types.DreckConfig{}
	err := parseConfig([]byte(`reviewers:
- aa
- ac

aliases:
- >
  /plugin: (.*) - /label add: plugin/$1
`), &config)
	if err != nil {
		t.Errorf("failed to parse config: %s", err)
	}
}

func TestReviewerConfigParseComment(t *testing.T) {
	config := types.DreckConfig{}
	parseConfig([]byte(`reviewers:
- aa with any comment following the github handle
- ab #with a real comment
- ac
`), &config)
	actual := len(config.Reviewers)
	if actual != 3 {
		t.Errorf("want: %d reviewers, got: %d", 3, actual)
	}
	expected := []string{"aa", "ab", "ac"}
	for i, r := range config.Reviewers {
		if r != expected[i] {
			t.Errorf("expected reviewer to be : %s, got: %s", expected[i], r)
		}
	}
}
