package dreck

import "testing"

func TestRunSanitize(t *testing.T) {
	tests := []struct {
		in string
		ok bool
	}{
		{"a", true},
		{"a b", true},
		{"1 2", true},
		{"1", true},
		//{".", true},
		//{"..", false},
		//{"...", false},
	}

	for _, test := range tests {
		ok := sanitize(test.in)
		if ok != test.ok {
			t.Errorf("Expected %t, got %t, for %s", test.ok, ok, test.in)
		}
	}
}
