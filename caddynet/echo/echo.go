package echo

import (
	"fmt"

	"github.com/mholt/caddy"
	"github.com/pieterlouw/caddy-net/caddynet/netserver"
)

func init() {
	caddy.RegisterPlugin("echo", caddy.Plugin{
		ServerType: "net",
		Action:     setupEcho,
	})
}

func setupEcho(c *caddy.Controller) error {
	config := netserver.GetConfig(c)
	for c.Next() {
		//if !c.Args(&config.Addr) {
		//	return c.ArgErr()
		//}

		fmt.Printf("[INFO] setupEcho config: %+v\n", config)
	}
	return nil
}
