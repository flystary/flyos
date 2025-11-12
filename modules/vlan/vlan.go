package vlan

type VLAN interface {
	// Validate checks that all required fields are present and valid.
	Validate() error
}