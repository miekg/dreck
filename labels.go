package dreck

import (
	"strings"

	"github.com/google/go-github/github"
	"github.com/miekg/dreck/types"
)

func labelDuplicate(currentLabels []types.IssueLabel, cmdLabel string) bool {

	for _, label := range currentLabels {
		if strings.EqualFold(label.Name, cmdLabel) {
			return true
		}
	}
	return false
}

func labelExists(all []*github.Label, label string) bool {

	for _, l := range all {
		if strings.EqualFold(*l.Name, label) {
			return true
		}
	}
	return false
}
