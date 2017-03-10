package proxyserver

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyfile"
	"github.com/mholt/caddy/caddytls"
)

const serverType = "proxy"

//tcpecho don't have directives
var directives = []string{"echo", "proxy"}

func init() {
	flag.StringVar(&LocalTCPAddr, serverType+".localtcp", DefaultLocalTCPAddr, "Default local TCP Address")
	flag.StringVar(&DestTCPAddr, serverType+".desttcp", DefaultDestTCPAddr, "Default destination TCP Address")

	caddy.RegisterServerType(serverType, caddy.ServerType{
		Directives: func() []string { return directives },
		DefaultInput: func() caddy.Input {
			return caddy.CaddyfileInput{
				Contents:       []byte(fmt.Sprintf("%s %s\n", LocalTCPAddr, DestTCPAddr)),
				ServerTypeName: serverType,
			}
		},
		NewContext: newContext,
	})
}

func newContext() caddy.Context {
	return &echoContext{}
}

type echoContext struct{}

// InspectServerBlocks for echo is a no-op
func (t *echoContext) InspectServerBlocks(sourceFile string, serverBlocks []caddyfile.ServerBlock) ([]caddyfile.ServerBlock, error) {

	fmt.Printf("[INFO] InspectServerBlocks [%s]\n", sourceFile)

	// For each address in each server block, make a new config
	for _, sb := range serverBlocks {

		for _, key := range sb.Keys {
			fmt.Printf("[INFO] range serverBlocks key [%s]\n", key)
		}
	}
	return serverBlocks, nil
}

// MakeServers uses the newly-created configs to create and return a list of server instances.
func (t *echoContext) MakeServers() ([]caddy.Server, error) {
	// create a server
	var servers []caddy.Server

	s, err := NewServer(LocalTCPAddr, DestTCPAddr)
	if err != nil {
		return nil, err
	}
	servers = append(servers, s)

	return servers, nil
}

// GetConfig gets the SiteConfig that corresponds to c.
// If none exist (should only happen in tests), then a
// new, empty one will be created.
func GetConfig(c *caddy.Controller) *SiteConfig {
	ctx := c.Context().(*httpContext)
	key := strings.ToLower(c.Key)
	if cfg, ok := ctx.keysToSiteConfigs[key]; ok {
		return cfg
	}
	// we should only get here during tests because directive
	// actions typically skip the server blocks where we make
	// the configs
	cfg := &SiteConfig{Root: Root, TLS: new(caddytls.Config)}
	ctx.saveConfig(key, cfg)
	return cfg
}

const (
	// DefaultLocalTCPAddr is the default local TCP Address.
	DefaultLocalTCPAddr = ":12017"

	// DefaultDestTCPAddr is the default destination TCP Address.
	DefaultDestTCPAddr = "localhost:22017"
)

// These "soft defaults" are configurable by
// command line flags, etc.
var (
	// LocalTCPAddr is the local TCP Address.
	LocalTCPAddr = DefaultLocalTCPAddr

	// LocalTCPAddr is the destination TCP Address.
	DestTCPAddr = DefaultDestTCPAddr
)
