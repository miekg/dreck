package dreck

import (
	"strings"

	"github.com/miekg/dreck/types"
)

func labelDuplicate(current []types.IssueLabel, label string) bool {

	for _, l := range current {
		if strings.EqualFold(l.Name, label) {
			return true
		}
	}
	return false
}
