package netserver

import (
	"fmt"
	"strings"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyfile"
	"github.com/mholt/caddy/caddytls"
)

const serverType = "net"

// directives for the net server type
// The ordering of this list is important, host need to be called before
// tls to get the relevant hostname needed
var directives = []string{"host", "tls"}

func init() {

	caddy.RegisterServerType(serverType, caddy.ServerType{
		Directives: func() []string { return directives },
		DefaultInput: func() caddy.Input {
			return caddy.CaddyfileInput{
				ServerTypeName: serverType,
			}
		},
		NewContext: newContext,
	})

	caddy.RegisterParsingCallback(serverType, "tls", activateTLS)
	caddytls.RegisterConfigGetter(serverType, func(c *caddy.Controller) *caddytls.Config { return GetConfig(c).TLS })
}

func newContext(inst *caddy.Instance) caddy.Context {
	return &netContext{instance: inst, keysToConfigs: make(map[string]*Config)}
}

type netContext struct {
	instance *caddy.Instance
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

type configTokens map[string][]string

// InspectServerBlocks make sure that everything checks out before
// executing directives and otherwise prepares the directives to
// be parsed and executed.
func (n *netContext) InspectServerBlocks(sourceFile string, serverBlocks []caddyfile.ServerBlock) ([]caddyfile.ServerBlock, error) {
	cfg := make(map[string]configTokens)

	// Example:
	// proxy :12017 :22017 {
	//	host localhost
	//	tls off
	// }
	// ServerBlock Keys will be proxy :12017 :22017 and Tokens will be host and tls

	// For each key in each server block, make a new config
	for _, sb := range serverBlocks {
		// build unique key from server block keys and join with '~' i.e echo~:12345
		key := ""
		for _, k := range sb.Keys {
			k = strings.ToLower(k)
			if key == "" {
				key = k
			} else {
				key += fmt.Sprintf("~%s", k)
			}
		}
		if _, dup := n.keysToConfigs[key]; dup {
			return serverBlocks, fmt.Errorf("duplicate key: %s", key)
		}

		tokens := make(map[string][]string)
		for k, v := range sb.Tokens {
			tokens[k] = []string{}
			for _, token := range v {
				tokens[k] = append(tokens[k], token.Text)
			}
		}

		cfg[key] = tokens
	}

	// build the actual Config from gathered data
	// key is the server block key joined by ~
	// value is the tokens (NOTE: tokens are not used at the moment)
	for k := range cfg {
		params := strings.Split(k, "~")
		listenType := params[0]
		params = params[1:]

		if len(params) == 0 {
			return serverBlocks, fmt.Errorf("invalid configuration: %s", k)
		}

		if listenType == "proxy" && len(params) < 2 {
			return serverBlocks, fmt.Errorf("invalid configuration: proxy server block expects a source and destination address")
		}

		// Make our caddytls.Config, which has a pointer to the
		// instance's certificate cache
		caddytlsConfig := caddytls.NewConfig(n.instance)

		// Save the config to our master list, and key it for lookups
		c := &Config{
			TLS:        caddytlsConfig,
			Type:       listenType,
			ListenPort: params[0], // first element should always be the port
			Parameters: params,
		}

		n.saveConfig(k, c)
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
			s, err := NewEchoServer(cfg.Parameters[0], cfg)
			if err != nil {
				return nil, err
			}
			servers = append(servers, s)
		case "proxy":
			s, err := NewProxyServer(cfg.Parameters[0], cfg.Parameters[1], cfg)
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

	//only check for config if the value is proxy or echo
	//we need to do this because we specify the ports in the server block
	//and those values need to be ignored as they are also sent from caddy main process.
	if key == "echo" || key == "proxy" {
		if cfg, ok := ctx.keysToConfigs[key]; ok {
			return cfg
		}
	}

	// we should only get here if value of key in server block
	// is not echo or proxy i.e port number :12017
	// we can't return a nil because caddytls.RegisterConfigGetter will panic
	// so we return a default (blank) config value
	caddytlsConfig := caddytls.NewConfig(ctx.instance)

	return &Config{TLS: caddytlsConfig}
}
