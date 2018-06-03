package dreck

import "github.com/mholt/caddy/caddyhttp/httpserver"

const (
	// Trigger is the prefix that triggers action from this bot.
	Trigger = "/"
	// The App's private key to access Github.
	//PrivateKeyPath = "/home/miek/dreck.2018-05-05.private-key.pem"
)

// Dreck is a plugin that handles Github Issues and Pull Requests for you.
type Dreck struct {
	Next httpserver.Handler

	clientID string
	key      string

	owners   string
	secret   string
	path     string // when should dreck trigger, default to '/dreck'
	hmac     bool   // validate HMAC on the webhook
	strategy string
}

// New returns a new, initialized Dreck.
func New() Dreck {
	d := Dreck{}
	d.owners = "OWNERS"
	d.path = "/dreck"
	d.strategy = mergeSquash

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
)
