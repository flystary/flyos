package routing

import (
	"errors"
	"net"
)

var (
	// ErrMissingField indicates a required field is missing.
	ErrMissingField = errors.New("missing required field")

	// ErrInvalidCIDR indicates the provided prefix is not a valid CIDR.
	ErrInvalidCIDR = errors.New("invalid CIDR")

	// ErrInvalidIPv4 indicates the provided IP address is not a valid IPv4.
	ErrInvalidIPv4 = errors.New("invalid IPv4 address")

	// ErrMissingViaOrDev indicates neither 'via' nor 'dev' was specified.
	ErrMissingViaOrDev = errors.New("either 'via' or 'dev' must be specified")
)

// BaseRoute holds common fields shared by all route types.
type BaseRoute struct {
	// Prefix is the destination network in CIDR format (e.g., "192.168.1.0/24").
	Prefix string

	// Via is the next-hop IPv4 address.
	Via string

	// Dev is the outgoing network interface name (e.g., "eth0").
	Dev string

	// Table specifies the routing table name (default: "main").
	Table string

	// Scope defines the route scope (default: "global").
	Scope string

	// Metric is the route metric (lower is preferred).
	Metric int

	// Proto is the route protocol type (set automatically by concrete types).
	Proto string
}

func (b *BaseRoute) ValidateBase() error {
	if b.Prefix == "" {
		return errors.Join(ErrMissingField, errors.New("prefix"))
	}
	if _, _, err := net.ParseCIDR(b.Prefix); err != nil {
		if ip := net.ParseIP(b.Prefix); ip != nil {
			b.Prefix += "/32"
		} else {
			return errors.Join(ErrInvalidCIDR, errors.New(b.Prefix))
		}
	}

	if b.Via != "" {
		viaIP := net.ParseIP(b.Via)
		if viaIP == nil || viaIP.To4() == nil {
			return errors.Join(ErrInvalidIPv4, errors.New("via: "+b.Via))
		}
	}

	if b.Table == "" {
		b.Table = "main"
	}

	if b.Scope == "" {
		b.Scope = "global"
	}

	if b.Via == "" && b.Dev == "" {
		return ErrMissingViaOrDev
	}
	return nil
}
