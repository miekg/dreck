// +build fortune

package dreck

import "testing"

func TestFortune(t *testing.T) {
	out, err := runFortune()
	if err != nil {
		t.Errorf("failed to run fortune: %s", err)
	}
	if out[0] != '>' {
		t.Errorf("first character of output must be '%s'", ">")
	}
}
