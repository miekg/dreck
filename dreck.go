package dreck

import "github.com/mholt/caddy/caddyhttp/httpserver"

const (
	// Trigger is the prefix that triggers action from this bot.
	Trigger = "/"
	// The App's private key to access Github.
	//PrivateKeyPath = "/home/miek/dreck.2018-05-05.private-key.pem"
)

type Dreck struct {
	Next httpserver.Handler

	clientID string
	key      string

	owners string
	secret string
	path   string // when should dreck trigger, default to '/dreck'
	hmac   bool   // validate HMAC on the webhook
}

func New() Dreck {
	d := Dreck{}
	d.owners = "OWNERS"
	d.path = "/dreck"

	return d
}

const (
	// featureDCO enables the "Signed-off-by" checking of PRs.
	featureDCO = "dco"
	// featureComments allows commands to be given in comments.
	featureComments = "comments"
)
