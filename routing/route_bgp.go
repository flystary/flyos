package routing

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

// BGPRoute represents a BGP-advertised route with extended attributes.
type BGPRoute struct {
	BaseRoute

	// LocalPref is the BGP local preference value.
	LocalPref uint32

	// MED is the Multi-Exit Discriminator.
	MED uint32

	// ASPath is the list of AS numbers in the path.
	ASPath []uint32

	// Communities is a list of BGP community values (as 32-bit integers).
	Communities []uint32

	// NoExport, if true, appends the NO_EXPORT community (0xFFFFFF01).
	NoExport bool

	// NoAdv is reserved for future use (not implemented).
	NoAdv bool
}

func init() {
	RegisterRoute("bgp", defaultEnabled, NewBGPRoute)
}

func NewBGPRoute() Route {
	return &BGPRoute{
		BaseRoute: BaseRoute{Table: "main", Scope: "global"},
	}
}

func (r *BGPRoute) Validate() error {
	r.Proto = "bgp"
	return r.ValidateBase()
}

func (r *BGPRoute) GetPrefix() string { return r.Prefix }
func (r *BGPRoute) GetProto() string  { return r.Proto }

func ParseCommunity(s string) (uint32, error) {
	if strings.Contains(s, ":") {
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			return 0, errors.New("community must be A:B")
		}
		a, err1 := strconv.ParseUint(parts[0], 10, 16)
		b, err2 := strconv.ParseUint(parts[1], 10, 16)
		if err1 != nil || err2 != nil {
			return 0, errors.New("A and B must be <= 65535")
		}
		return uint32(a<<16 | b), nil
	}
	v, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		return 0, errors.New("invalid community format")
	}
	return uint32(v), nil
}

func (r *BGPRoute) ToArgs() ([]string, error) {
	_, ipNet, err := net.ParseCIDR(r.Prefix)
	if err != nil {
		return nil, err
	}
	ip := ipNet.IP.To4()
	mask := net.IP(ipNet.Mask).To4()

	args := []string{
		"--ip", ip.String(),
		"--netmask", mask.String(),
	}
	if r.LocalPref > 0 {
		args = append(args, "--local-pref", strconv.FormatUint(uint64(r.LocalPref), 10))
	}
	if len(r.ASPath) != 0 {
		asStrs := make([]string, len(r.ASPath))
		for i, as := range r.ASPath {
			asStrs[i] = strconv.FormatUint(uint64(as), 10)
		}
		args = append(args, "--as-path", strings.Join(asStrs, " "))
	}
	if r.Table != "" {
		args = append(args, "--table", r.Table)
	}
	return args, nil
}
