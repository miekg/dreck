package dreck

import "github.com/mholt/caddy/caddyhttp/httpserver"

const (
	// Trigger is the prefix that triggers action from this bot.
	Trigger = "/"
	// Owners is the main file for the permissions.
	Owners = "OWNERS"
	// The App's private key to access Github.
	PrivateKeyPath = "/home/miek/dreck.2018-05-05.private-key.pem"
	// Our application ID.
	ApplicationID = "11824"
)

type Dreck struct {
	Next httpserver.Handler

	clientID string
	key      string

	owners string
	secret string
	path   string // when should dreck trigger, default to '/dreck'
}

func New() Dreck {
	d := Dreck{}
	d.owners = "OWNERS"
	d.path = "/dreck"

	return d
}
