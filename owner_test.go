package dreck

import (
	"reflect"
	"testing"

	"github.com/google/go-github/github"
)

func TestOwnersSingle(t *testing.T) {
	t1 := "/home/example/test.txt"
	files := []*github.CommitFile{
		&github.CommitFile{Filename: &t1},
	}
	victims := findReviewers(files, "OWNERS", func(path string) ([]byte, error) {
		return []byte(`reviewers:
- ab
- ac
`), nil
	})

	expect := map[string]bool{
		"ab": true,
		"ac": true,
	}

	if !reflect.DeepEqual(victims, expect) {
		t.Errorf("expected %v, got %v", expect, victims)
	}
}

func TestOwnersMultiple(t *testing.T) {
	t1 := "/home/example/a/test.txt"
	t2 := "/home/example/b/test.txt"
	files := []*github.CommitFile{
		&github.CommitFile{Filename: &t1},
		&github.CommitFile{Filename: &t2},
	}
	victims := findReviewers(files, "OWNERS", func(path string) ([]byte, error) {
		switch path {
		case "/home/example/a/OWNERS":
			return []byte(`reviewers:
- ab
- ac
`), nil
		case "/home/example/b/OWNERS":
			return []byte(`reviewers:
- xb
- xc
`), nil
		}
		return nil, nil
	})

	expect := map[string]bool{
		"ab": true,
		"ac": true,
		"xb": true,
		"xc": true,
	}

	if !reflect.DeepEqual(victims, expect) {
		t.Errorf("expected %v, got %v", expect, victims)
	}
}

func TestOwnersMostSpecific(t *testing.T) {
	t1 := "/home/plugin/reload/test.txt"
	files := []*github.CommitFile{
		&github.CommitFile{Filename: &t1},
	}
	victims := findReviewers(files, "OWNERS", func(path string) ([]byte, error) {
		switch path {
		case "/home/plugin/reload/OWNERS":
			return []byte(`reviewers:
- aa
`), nil
		case "/home/plugin/OWNERS":
			return []byte(`reviewers:
- bb
`), nil
		}
		return nil, nil
	})

	expect := map[string]bool{
		"aa": true,
	}

	if !reflect.DeepEqual(victims, expect) {
		t.Errorf("expected %v, got %v", expect, victims)
	}
}
