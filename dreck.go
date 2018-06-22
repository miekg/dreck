package dreck

import "github.com/mholt/caddy/caddyhttp/httpserver"

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
	d.user = "nobody"

	return d
}

const (
	// featureDCO enables the "Signed-off-by" checking of PRs.
	featureDCO = "dco"
	// featureComments allows commands to be given in comments.
	featureComments = "comments"
	// featureReviewers enables automatically assigning reviewers based on OWNERS.
	featureReviewers = "reviewers"
	// featureAliases enables alias expansion.
	featureAliases = "aliases"
	// featureBranches enables branch deletion after a merge.
	featureBranches = "branches"
	// featureAutosubmit enables the auto submitting or pull requests when the tests are green.
	featureAutosubmit = "autosubmit"
	// featureExec enables the exec command.
	featureExec = "exec"
)

// Trigger is the prefix that triggers action from this bot.
const Trigger = "/"
