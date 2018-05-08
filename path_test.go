package dreck

import "testing"

func TestOwnersPath(t *testing.T) {
	ex := []string{"OWNERS",
		"/home/miek/tmp/example/OWNERS",
		"/home/miek/tmp/OWNERS",
		"/home/miek/OWNERS",
		"/home/OWNERS",
		"/OWNERS"}

	p := "/home/miek/tmp/example/test"
	s := ownersPaths(p, "OWNERS")
	for i := range s {
		if s[i] != ex[i] {
			t.Errorf("expected %v, got %v", ex[i], s[i])
		}
	}
}
