package bond

type BOND interface {
	// Validate checks that all required fields are present and valid.
	Validate() error
}