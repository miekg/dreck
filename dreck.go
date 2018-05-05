package dreck

import "github.com/mholt/caddy/caddyhttp/httpserver"

type Dreck struct {
	Next httpserver.Handler

	path string // when should dreck trigger, default to '/dreck'
}
