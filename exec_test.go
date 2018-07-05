package dreck

import (
	"strings"
	"testing"

	"github.com/miekg/dreck/types"
)

var owner = []byte(`
approvers:
  - miek
features:
  - aliases
  - exec
aliases:
  - |
    /echo: (.*) -> /exec: /bin/echo $1
`)

var execOptions = []struct {
	title     string
	body      string
	shouldErr bool
	expected  string
}{
	{
		"Correct exec command",
		Trigger + "echo: boe",
		false,
		"boe\n",
	},
}

func TestExec(t *testing.T) {
	d := New()
	config := &types.DreckConfig{}
	parseConfig(owner, config)

	for _, test := range execOptions {
		execs := parse(test.body, config)
		if len(execs) == 0 && test.shouldErr {
			continue
		}
		if len(execs) == 0 && !test.shouldErr {
			t.Errorf("Exec - not parsed correctly")
			continue
		}
		exec := execs[0]

		run, err := stripValue(exec.Value)
		if err != nil {
			t.Errorf("Exec illegal command %s", run)
		}
		parts := strings.Fields(run) // simple split
		if len(parts) == 0 {
			t.Errorf("Exec illegal command %s", run)
		}
		ok := isValidExec(config, parts, run)
		if !ok {
			t.Errorf("Exec illegal command %s", run)
		}
		cmd, err := d.execCmd(parts, "42")
		if err != nil {
			t.Errorf("Exec could not be set up: %s", err)
		}
		out, err := cmd.Output()
		if err != nil {
			t.Errorf("Cmd not be executed: %s", err)
		}

		if x := string(out); x != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, x)
		}
	}
}
