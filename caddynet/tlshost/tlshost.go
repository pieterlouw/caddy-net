package tlshost

import (
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddytls"
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

	// Ignore call to setupTLSHost if the key is not echo or proxy
	if c.Key != "echo" && c.Key != "proxy" {
		return nil
	}

	for c.Next() {
		if !c.NextArg() {
			return c.ArgErr()
		}
		config.TLSHost = c.Val()

		if config.TLS == nil {
			config.TLS = &caddytls.Config{Hostname: c.Val()}
		} else {
			config.TLS.Hostname = c.Val()
		}

		if c.NextArg() {
			// only one argument allowed
			return c.ArgErr()
		}
	}

	return nil
}
