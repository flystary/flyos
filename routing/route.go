package routing

type Route interface {
	// Validate checks that all required fields are present and valid.
	Validate() error
	// GetPrefix returns the destination network in CIDR notation.
	GetPrefix() string
	// GetProto returns the route protocol name (e.g., "static", "bgp").
	GetProto() string
	// ToArgs return the route CLI Args
	ToArgs() ([]string, error)
}
