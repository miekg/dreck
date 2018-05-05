package dreck

import (
	"net/http"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("dreck", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

type Dreck struct {
	Next httpserver.Handler
	// more
}

func (d Dreck) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	return d.Next.ServeHTTP(w, r)
}
