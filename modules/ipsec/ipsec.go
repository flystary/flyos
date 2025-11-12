package ipsec

type IPSEC interface {
	// Validate checks that all required fields are present and valid.
	Validate() error
}