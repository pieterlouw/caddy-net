package proxy

import (
	"fmt"

	"github.com/mholt/caddy"
	"github.com/pieterlouw/caddy-net/caddynet/netserver"
)

func init() {
	caddy.RegisterPlugin("proxy", caddy.Plugin{
		ServerType: "net",
		Action:     setupProxy,
	})
}

func setupProxy(c *caddy.Controller) error {
	config := netserver.GetConfig(c)
	//get destination address of proxy
	for c.Next() {
		//if !c.Args(&config.Addr) {
		//	return c.ArgErr()
		//}

		fmt.Printf("[INFO] setupProxy config: %+v\n", config)
	}
	return nil
}
