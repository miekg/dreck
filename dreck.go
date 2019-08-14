package dreck

import (
	"github.com/miekg/dreck/types"

	"github.com/caddyserver/caddy/caddyhttp/httpserver"
	yaml "gopkg.in/yaml.v2"
)

// Dreck is a plugin that handles Github Issues and Pull Requests for you.
type Dreck struct {
	Next httpserver.Handler

	clientID string
	key      string

	owners   string
	secret   string
	path     string            // when should dreck trigger, default to '/dreck'
	hmac     bool              // validate HMAC on the webhook
	strategy string            // how to merge when we merge
	user     string            // user to use to exec commands
	env      map[string]string // environment to give to commands
}

// New returns a new, initialized Dreck.
func New() Dreck {
	d := Dreck{}
	d.owners = "OWNERS"
	d.path = "/dreck"
	d.strategy = mergeSquash
	d.env = make(map[string]string)

	return d
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

const (
	featureDCO        = "dco"        // featureDCO enables the "Signed-off-by" checking of PRs.
	featureComments   = "comments"   // featureComments allows commands to be given in comments.
	featureReviewers  = "reviewers"  // featureReviewers enables automatically assigning reviewers based on OWNERS.
	featureAliases    = "aliases"    // featureAliases enables alias expansion.
	featureBranches   = "branches"   // featureBranches enables branch deletion after a merge.
	featureAutosubmit = "autosubmit" // featureAutosubmit enables the auto submitting or pull requests when the tests are green.
	featureExec       = "exec"       // featureExec enables the exec command.
)

// Trigger is the prefix that triggers action from this bot.
const Trigger = "/"
