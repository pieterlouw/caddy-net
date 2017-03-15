package netserver

// Config contains configuration details about a net server type
type Config struct {
	Type       string
	Parameters []string
	Tokens     map[string][]string
}
