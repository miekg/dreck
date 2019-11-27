package dreck

import (
	"strings"

	"github.com/miekg/dreck/types"
)

func aliasOK(c *types.DreckConfig) bool {
	for _, f := range config.Features {
		if strings.EqualFold(Aliases, availableFeature) {
			return true
		}
	}
	return false
}

func execOK(c *types.DreckConfig) bool {
	for _, f := range config.Features {
		if strings.EqualFold(Exec, availableFeature) {
			return true
		}
	}
	return false
}

func isCodeOwner(c *types.DreckConfig, user string) bool {
	for _, o := range c.CodeOwners {
		if strings.EqualFold(user, o) {
			return true
		}
	}
	return false
}
