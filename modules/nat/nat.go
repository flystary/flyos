package nat

type NAT interface {
	// Validate checks that all required fields are present and valid.
	Validate() error
}