// Package plugin is a Caddy plugin that tunnels UDP/TCP data through HTTP.
package plugin

import (
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("sstp", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	server := &Server{}

	for c.Next() { // skip the directive name
		for c.NextBlock() {
			directive := c.Val()
			args := c.RemainingArgs()
			_ = args
			server.requestPath = directive
		}
	}

	cfg := httpserver.GetConfig(c)
	mid := func(next httpserver.Handler) httpserver.Handler {
		server.NextHandler = next
		return server
	}
	cfg.AddMiddleware(mid)

	return nil
}
