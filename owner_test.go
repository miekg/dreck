package dreck

import (
	"reflect"
	"testing"

	"github.com/google/go-github/github"
)

func TestOwnersSingle(t *testing.T) {
	files := []*github.CommitFile{
		&github.CommitFile{Filename: String("/home/example/test.txt")},
	}
	victims := findReviewers(files, "OWNERS", func(path string) ([]byte, error) {
		return []byte(`reviewers:
- ab
- ac
`), nil
	})

	expect := map[string]string{
		"ab": "/home/example/OWNERS",
		"ac": "/home/example/OWNERS",
	}

	if !reflect.DeepEqual(victims, expect) {
		t.Errorf("expected %v, got %v", expect, victims)
	}
}

func TestOwnersMultiple(t *testing.T) {
	files := []*github.CommitFile{
		&github.CommitFile{Filename: String("/home/example/a/test.txt")},
		&github.CommitFile{Filename: String("/home/example/b/test.txt")},
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

	expect := map[string]string{
		"ab": "/home/example/a/OWNERS",
		"ac": "/home/example/a/OWNERS",
		"xb": "/home/example/b/OWNERS",
		"xc": "/home/example/b/OWNERS",
	}

	if !reflect.DeepEqual(victims, expect) {
		t.Errorf("expected %v, got %v", expect, victims)
	}
}

func TestOwnersMostSpecific(t *testing.T) {
	files := []*github.CommitFile{
		&github.CommitFile{Filename: String("/home/plugin/reload/test.txt")},
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

	expect := map[string]string{
		"aa": "/home/plugin/reload/OWNERS",
	}

	if !reflect.DeepEqual(victims, expect) {
		t.Errorf("expected %v, got %v", expect, victims)
	}
}
