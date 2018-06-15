// Package plugin is a Caddy plugin that tunnels UDP/TCP data through HTTP.
package plugin

import (
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("tunnel", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	server := &Server{}

	for c.Next() { // skip the directive name
		if !c.Args(&server.RequestPath, &server.Upstream) {
			return c.ArgErr()
		}

		for c.NextBlock() {
			directive := c.Val()

			switch directive {
			case "protocol":
				if !c.Args(&server.UpstreamProto) {
					return c.ArgErr()
				}
			default:
				return c.ArgErr()
			}
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
