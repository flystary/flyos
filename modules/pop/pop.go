package pop

type PoP interface {
	// Validate checks that all required fields are present and valid.
	Validate() error
}