package dns

type DNS interface {
	// Validate checks that all required fields are present and valid.
	Validate() error
}