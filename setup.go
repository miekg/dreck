package dreck

import (
	"fmt"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
)

func setup(c *caddy.Controller) error {
	dr, err := parseConfig(c)
	if err != nil {
		return err
	}
	if dr.clientID == "" {
		return fmt.Errorf("need a valid client_id")
	}
	if dr.key == "" {
		return fmt.Errorf("need a path to a private key file")
	}
	if dr.secret == "" {
		return fmt.Errorf("need a webhook secret")
	}
	if !dr.hmac {
		return fmt.Errorf("need a to validate the webhook secret")
	}

	dreck := func(next httpserver.Handler) httpserver.Handler {
		dr.Next = next
		return dr
	}
	httpserver.GetConfig(c).AddMiddleware(dreck)

	return nil
}

func parseConfig(c *caddy.Controller) (Dreck, error) {
	d := New()
	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "client_id":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return d, c.ArgErr()
				}
				d.clientID = args[0]
			case "private_key":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return d, c.ArgErr()
				}
				d.key = args[0]
			case "owners":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return d, c.ArgErr()
				}
				d.owners = args[0]
			case "secret":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return d, c.ArgErr()
				}
				d.secret = args[0]
			case "path":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return d, c.ArgErr()
				}
				d.path = args[0]
			case "validate":
				args := c.RemainingArgs()
				if len(args) != 0 {
					return d, c.ArgErr()
				}
				d.hmac = true
			case "merge":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return d, c.ArgErr()
				}
				if args[0] != mergeMerge && args[0] != mergeSquash && args[0] != mergeRebase {
					return d, c.Errf("invalid strategy: %s", args[0])
				}
				d.strategy = args[0]
			case "user":
				args := c.RemainingArgs()
				if len(args) != 1 {
					return d, c.ArgErr()
				}
				d.user = args[0]
			case "env":
				args := c.RemainingArgs()
				if len(args) != 2 {
					return d, c.ArgErr()
				}
				d.env[args[0]] = args[1]
			}
		}
	}
	return d, nil
}

const (
	mergeMerge  = "merge"
	mergeSquash = "squash"
	mergeRebase = "rebase"
)
