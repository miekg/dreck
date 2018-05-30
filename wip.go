package dreck

import "strings"

func hasWIPPrefix(s string) bool {
	for _, w := range wip {
		if strings.HasPrefix(s, w) {
			return true
		}
	}
	return false
}

var wip = []string{"WIP", "[WIP]", "WIP:", "[WIP]:"}
