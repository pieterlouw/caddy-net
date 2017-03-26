package netserver

import "github.com/mholt/caddy/caddytls"

// Config contains configuration details about a net server type
type Config struct {
	Type string

	// The hostname to be used for TLS configurations
	TLSHost string

	// The port the server binds to and listens on
	ListenPort string

	// TLS configuration
	TLS *caddytls.Config

	Parameters []string
	Tokens     map[string][]string
}

// TLSConfig returns c.TLS.
func (c Config) TLSConfig() *caddytls.Config {
	return c.TLS
}

// Host returns c.TLSHost
func (c Config) Host() string {
	return c.TLSHost
}

// Port returns c.ListenPort
func (c Config) Port() string {
	return c.ListenPort
}
