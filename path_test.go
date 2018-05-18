package dreck

import (
	"reflect"
	"testing"
)

func TestOwnersPath(t *testing.T) {
	ex := []string{"/home/miek/tmp/example/OWNERS",
		"/home/miek/tmp/OWNERS",
		"/home/miek/OWNERS",
		"/home/OWNERS",
		"/OWNERS"}

	p := "/home/miek/tmp/example/test.txt"
	s := ownersPaths(p, "OWNERS")
	for i := range s {
		if s[i] != ex[i] {
			t.Errorf("expected %v, got %v", ex[i], s[i])
		}
	}
}

func TestMostSpecific(t *testing.T) {
	ex := []string{"/home/miek/tmp/example/OWNERS",
		"/home/miek/tmp/OWNERS",
		"/home/miek/OWNERS",
		"/home/OWNERS",
		"/OWNERS"}

	m := mostSpecific(ex)
	expect := map[string]int{
		"/home/miek/tmp/example/OWNERS": 1,
		"/home/miek/tmp/OWNERS":         1,
		"/home/miek/OWNERS":             1,
		"/home/OWNERS":                  1,
		"/OWNERS":                       1,
	}

	if !reflect.DeepEqual(m, expect) {
		t.Errorf("expected %v, got %v", expect, m)
	}

	ex = []string{"/home/miek/tmp/example/OWNERS",
		"/home/miek/tmp/example/OWNERS",
		"/home/miek/tmp/OWNERS",
		"/home/miek/OWNERS"}

	m = mostSpecific(ex)
	expect = map[string]int{
		"/home/miek/tmp/example/OWNERS": 2,
		"/home/miek/tmp/OWNERS":         1,
		"/home/miek/OWNERS":             1,
	}

	if !reflect.DeepEqual(m, expect) {
		t.Errorf("expected %v, got %v", expect, m)
	}
}

func TestSortOnOccurence(t *testing.T) {
	ex := []string{"/home/miek/tmp/example/OWNERS",
		"/home/miek/tmp/example/OWNERS",
		"/home/miek/tmp/OWNERS",
		"/home/miek/OWNERS"}

	m := mostSpecific(ex)
	o := sortOnOccurence(m)
	expect := [][]string{
		[]string{},
		[]string{"/home/miek/tmp/OWNERS", "/home/miek/OWNERS"},
		[]string{"/home/miek/tmp/example/OWNERS"},
	}
	if len(o) != len(expect) {
		t.Errorf("expected %v, got %v", expect, o)
	}
	for i := range o {
		ex := expect[i]
		vo := o[i]
		for j := range vo {
			if vo[j] != ex[j] {
				t.Errorf("expected %s, got %s", vo[j], ex[j])
			}
		}
	}
}
