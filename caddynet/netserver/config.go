package netserver

// Config contains information about a netserver type
type Config struct {
	Type   string //echo, proxy
	Addr   string
	Tokens map[string][]string
}
