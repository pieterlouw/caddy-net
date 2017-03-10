package echo

import (
	"github.com/pieterlouw/caddy-proxy/proxy/proxyserver""
)

func init() {
	caddy.RegisterPlugin("echo", caddy.Plugin{
		ServerType: "proxy",
		Action:     setupEcho,
	})
}

func setupBind(c *caddy.Controller) error {
	config := echoserver.GetConfig(c)
	for c.Next() {
		if !c.Args(&config.ListenHost) {
			return c.ArgErr()
		}		
	}
	return nil
}