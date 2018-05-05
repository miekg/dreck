package dreck

import (
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func setup(c *caddy.Controller) error {

	dr, err := parseDreck(c)
	if err != nil {
		return err
	}

	mid := func(next httpserver.Handler) httpserver.Handler {
		dr.Next = next
		return dr
	}
	httpserver.GetConfig(c).AddMiddleware(mid)

	return nil
}

func parseDreck(c *caddy.Controller) (Dreck, error) {
	for c.Next() {
		// get configuration
	}
	return Dreck{}, nil
}
