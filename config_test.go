package dreck

import (
	"testing"

	"github.com/miekg/dreck/types"
	yaml "gopkg.in/yaml.v2"
)

func TestOwnersConfigParse(t *testing.T) {
	owners, err := parseOwners([]byte(`# Order is important; the last matching pattern takes the most
*.js    @js-owner
*       @miekg @blaa
`))
	if err != nil {
		t.Error(err)
	}
	if actual := len(owners); actual != 3 {
		t.Errorf("want: %d approvers, got: %d", 3, actual)
	}
	found := false
	for i := range owners {
		if owners[i] == "miekg" {
			found = true
		}
	}
	if !found {
		t.Errorf("want: %q in owners", "miekg")
	}
}

func TestAliasConfigParse(t *testing.T) {
	config := types.DreckConfig{}
	buf := []byte(`aliases:
- |
  /plugin (.*) - /label plugin/$1
`)
	err := yaml.Unmarshal(buf, &config)
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
	buf := []byte(`aliases:
- >
  /plugin (.*) - /label plugin/$1
`)
	err := yaml.Unmarshal(buf, &config)
	if err != nil {
		t.Errorf("failed to parse config: %s", err)
	}
}
