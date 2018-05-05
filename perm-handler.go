package dreck

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/miekg/dreck/types"

	yaml "gopkg.in/yaml.v2"
)

const configFile = ".dreck.yml"

func enabledFeature(attemptedFeature string, config *types.DerekConfig) bool {

	featureEnabled := false

	for _, availableFeature := range config.Features {
		if strings.EqualFold(attemptedFeature, availableFeature) {
			featureEnabled = true
			break
		}
	}
	return featureEnabled
}

func permittedUserFeature(attemptedFeature string, config *types.DerekConfig, user string) bool {

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

func getConfig(owner string, repository string) (*types.DerekConfig, error) {

	var config types.DerekConfig

	maintainersFile := fmt.Sprintf("https://github.com/%s/%s/raw/master/%s", owner, repository, Owners)

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	req, _ := http.NewRequest(http.MethodGet, maintainersFile, nil)

	res, resErr := client.Do(req)
	if resErr != nil {
		return nil, resErr
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Status code: %d while fetching maintainers list (%s)", res.StatusCode, maintainersFile)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, _ := ioutil.ReadAll(res.Body)

	err := parseConfig(bytesOut, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func parseConfig(bytesOut []byte, config *types.DerekConfig) error {
	err := yaml.Unmarshal(bytesOut, &config)

	if len(config.Reviewers) == 0 && len(config.Approvers) > 0 {
		config.Reviewers = config.Approvers
	}

	return err
}
