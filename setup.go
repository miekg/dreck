package dreck

import (
	"fmt"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func setup(c *caddy.Controller) error {

	dr, err := parseDreck(c)
	if err != nil {
		return err
	}
	if dr.clientID == "" {
		return fmt.Errorf("Need a valid client_id")
	}
	if dr.key == "" {
		return fmt.Errorf("Need a private_key for a path to a private key file")
	}

	mid := func(next httpserver.Handler) httpserver.Handler {
		dr.Next = next
		return dr
	}
	httpserver.GetConfig(c).AddMiddleware(mid)

	return nil
}

func parseDreck(c *caddy.Controller) (Dreck, error) {
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
			}
		}
	}
	return d, nil
}
