package caddynet

import (
	// plug in the server
	_ "github.com/pieterlouw/caddy-net/caddynet/netserver"

	// plug in the standard directives
	_ "github.com/pieterlouw/caddy-net/caddynet/echo"
	_ "github.com/pieterlouw/caddy-net/caddynet/proxy"
)
