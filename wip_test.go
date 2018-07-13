package dreck

import "testing"

func TestWIP(t *testing.T) {
	wiptests := []struct {
		title string
		wip   bool
	}{
		{"Correct reopen command", false},
		{"WIP Correct reopen command", true},
		{"wip Correct reopen command", true},
	}

	for _, w := range wiptests {
		if x := hasWIPPrefix(w.title); x != w.wip {
			t.Errorf("expected %s to be %t, bug got %t", w.title, w.wip, x)
		}
	}
}
