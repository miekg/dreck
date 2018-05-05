package dreck

import "github.com/mholt/caddy/caddyhttp/httpserver"

const (
	// Trigger is the prefix that triggers action from this bot.
	Trigger = "/"
	// Owners is the main file for the permissions
	Owners = "OWNERS"
)

type Dreck struct {
	Next httpserver.Handler

	path string // when should dreck trigger, default to '/dreck'
}
