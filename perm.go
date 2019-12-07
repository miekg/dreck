package dreck

import (
	"strings"
)

func aliasOK(c *DreckConfig) bool {
	for _, f := range c.Features {
		if strings.EqualFold(Aliases, f) {
			return true
		}
	}
	return false
}

func execOK(c *DreckConfig) bool {
	for _, f := range c.Features {
		if strings.EqualFold(Exec, f) {
			return true
		}
	}
	return false
}

func isCodeOwner(c *DreckConfig, user string) bool {
	for _, o := range c.CodeOwners {
		if strings.EqualFold(user, o) {
			return true
		}
	}
	return false
}
