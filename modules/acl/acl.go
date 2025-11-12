package acl

type ACLRule interface {
	// Validate checks that all required fields are present and valid.
	Validate() error
	// GetProto returns the route protocol name (e.g., "static", "bgp").
	GetProto() string
	// ToArgs return the route CLI Args
	ToArgs() ([]string, error)
}
