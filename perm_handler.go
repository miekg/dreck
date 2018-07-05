package dreck

import (
	"strings"

	"github.com/miekg/dreck/types"
)

func enabledFeature(attemptedFeature string, config *types.DreckConfig) bool {
	for _, availableFeature := range config.Features {
		if strings.EqualFold(attemptedFeature, availableFeature) {
			return true
		}
	}
	return false
}

func permittedUserFeature(attemptedFeature string, config *types.DreckConfig, user string) bool {

	if enabledFeature(attemptedFeature, config) {
		for _, reviewer := range config.Reviewers {
			if strings.EqualFold(user, reviewer) {
				return true
			}
		}
	}
	if enabledFeature(attemptedFeature, config) {
		for _, approver := range config.Approvers {
			if strings.EqualFold(user, approver) {
				return true
			}
		}
	}

	return false
}
