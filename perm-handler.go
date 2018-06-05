package dreck

import (
	"strings"

	"github.com/miekg/dreck/types"

	yaml "gopkg.in/yaml.v2"
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

func permittedUserFeatureRun(config *types.DreckConfig, user string) bool {
	if enabledFeature(featureRun, config) {
		for _, runner := range config.Runners {
			if strings.EqualFold(user, runner) {
				return true
			}
		}

	}
	return false
}

func (d Dreck) getConfig(owner string, repository string) (*types.DreckConfig, error) {

	var config types.DreckConfig

	buf, err := githubFile(owner, repository, d.owners)
	if err != nil {
		return nil, err
	}

	if err := parseConfig(buf, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func parseConfig(bytesOut []byte, config *types.DreckConfig) error {
	err := yaml.Unmarshal(bytesOut, &config)

	if len(config.Reviewers) == 0 && len(config.Approvers) > 0 {
		config.Reviewers = config.Approvers
	}

	return err
}
