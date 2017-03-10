package netserver

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyfile"
)

const serverType = "net"

//tcpecho don't have directives
var directives = []string{"echo", "proxy"}

func init() {
	flag.StringVar(&LocalTCPAddr, serverType+".localtcp", DefaultLocalTCPAddr, "Default local TCP Address")

	caddy.RegisterServerType(serverType, caddy.ServerType{
		Directives: func() []string { return directives },
		DefaultInput: func() caddy.Input {
			return caddy.CaddyfileInput{
				Contents:       []byte(fmt.Sprintf("%s\n", LocalTCPAddr)),
				ServerTypeName: serverType,
			}
		},
		NewContext: newContext,
	})
}

func newContext() caddy.Context {
	return &netContext{keysToConfigs: make(map[string]*Config)}
}

type netContext struct {
	// keysToConfigs maps an address at the top of a
	// server block (a "key") to its Config. Not all
	// Configs will be represented here, only ones
	// that appeared in the Caddyfile.
	keysToConfigs map[string]*Config

	// configs is the master list of all site configs.
	configs []*Config
}

func (n *netContext) saveConfig(key string, cfg *Config) {
	n.configs = append(n.configs, cfg)
	n.keysToConfigs[key] = cfg
}

// InspectServerBlocks make sure that everything checks out before
// executing directives and otherwise prepares the directives to
// be parsed and executed.
func (n *netContext) InspectServerBlocks(sourceFile string, serverBlocks []caddyfile.ServerBlock) ([]caddyfile.ServerBlock, error) {

	fmt.Printf("[INFO] InspectServerBlocks [%s]\n", sourceFile)

	// For each address in each server block, make a new config
	for _, sb := range serverBlocks {
		for _, key := range sb.Keys {
			fmt.Printf("[INFO] range serverBlocks key [%s]\n", key)
			fmt.Printf("[INFO] range serverBlock tokens [%+v]\n", sb.Tokens)

			key = strings.ToLower(key)
			if _, dup := n.keysToConfigs[key]; dup {
				return serverBlocks, fmt.Errorf("duplicate address: %s", key)
			}

			_, isEchoServer := sb.Tokens["echo"]
			_, isProxyServer := sb.Tokens["proxy"]

			netType := ""
			if isEchoServer {
				netType = "echo"
			} else if isProxyServer {
				netType = "proxy"
			}

			if netType == "" {
				return serverBlocks, fmt.Errorf("invalid server type: %s", key)
			}

			tokens := make(map[string][]string)
			for k, v := range sb.Tokens {
				tokens[k] = []string{}
				for _, token := range v {
					tokens[k] = append(tokens[k], token.Text)
				}
			}

			// Save the config to our master list, and key it for lookups
			cfg := &Config{
				Addr:   key,
				Type:   netType,
				Tokens: tokens,
			}
			n.saveConfig(key, cfg)
		}
	}
	return serverBlocks, nil
}

// MakeServers uses the newly-created configs to create and return a list of server instances.
func (n *netContext) MakeServers() ([]caddy.Server, error) {
	//  create servers based on config type
	var servers []caddy.Server
	for _, cfg := range n.configs {
		switch cfg.Type {
		case "echo":
			s, err := NewEchoServer(cfg.Addr)
			if err != nil {
				return nil, err
			}
			servers = append(servers, s)
		case "proxy":
			s, err := NewProxyServer(cfg.Addr, "localhost:22017")
			if err != nil {
				return nil, err
			}
			servers = append(servers, s)

		}
	}

	return servers, nil
}

// GetConfig gets the Config that corresponds to c.
// If none exist (should only happen in tests), then a
// new, empty one will be created.
func GetConfig(c *caddy.Controller) *Config {
	ctx := c.Context().(*netContext)
	key := strings.ToLower(c.Key)
	if cfg, ok := ctx.keysToConfigs[key]; ok {
		return cfg
	}
	// we should only get here during tests because directive
	// actions typically skip the server blocks where we make
	// the configs
	cfg := &Config{Addr: key}
	ctx.saveConfig(key, cfg)
	return cfg
}

const (
	// DefaultLocalTCPAddr is the default local TCP Address.
	DefaultLocalTCPAddr = ":12017"
)

// These "soft defaults" are configurable by
// command line flags, etc.
var (
	// LocalTCPAddr is the local TCP Address.
	LocalTCPAddr = DefaultLocalTCPAddr
)
