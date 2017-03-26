package tlshost

import (
	"github.com/mholt/caddy"
	"github.com/pieterlouw/caddy-net/caddynet/netserver"
)

func init() {
	caddy.RegisterPlugin("tlshost", caddy.Plugin{
		ServerType: "net",
		Action:     setupTLSHost,
	})
}

func setupTLSHost(c *caddy.Controller) error {
	config := netserver.GetConfig(c)

	for c.Next() {
		if !c.NextArg() {
			return c.ArgErr()
		}
		config.TLSHost = c.Val()
		if c.NextArg() {
			// only one argument allowed
			return c.ArgErr()
		}
	}

	return nil
}
