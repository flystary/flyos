package routing

import (
	"errors"
	"net"
	"strconv"
)

// OSPFRoute represents an OSPF-learned or injected route.
type OSPFRoute struct {
	BaseRoute

	// Area is the OSPF area ID (e.g., "0.0.0.0").
	Area string

	// Type specifies the OSPF route type.
	// Valid values: "intra-area", "inter-area", "external-1", "external-2".
	Type string

	// Tag is an optional tag for external routes.
	Tag uint32
}

func init() {
	RegisterRoute("ospf", defaultDisabled, NewOSPFRoute)
}

func NewOSPFRoute() Route {
	return &OSPFRoute{
		BaseRoute: BaseRoute{
			Table: "main",
			Scope: "global",
		},
	}
}

func (r *OSPFRoute) Validate() error {
	r.Proto = "ospf"
	if err := r.ValidateBase(); err != nil {
		return err
	}
	if r.Type != "" {
		valib := map[string]bool{
			"intra-area": true,
			"inter-area": true,
			"external-1": true,
			"external-2": true,
		}
		if !valib[r.Type] {
			return errors.New("invalib OSPF type: " + r.Type)
		}
	}
	return nil
}

func (r *OSPFRoute) GetPrefix() string { return r.Prefix }
func (r *OSPFRoute) GetProto() string  { return r.Proto }

func (r *OSPFRoute) ToArgs() ([]string, error) {
	_, ipNet, err := net.ParseCIDR(r.Prefix)
	if err != nil {
		return nil, err
	}
	ip := ipNet.IP.To4()
	mask := net.IP(ipNet.Mask).To4()

	args := []string{
		"--ip", ip.String(),
		"--netmask", mask.String(),
		"--area", r.Area,
	}
	if r.Metric > 0 {
		args = append(args, "--metric", strconv.FormatUint(uint64(r.Metric), 10))
	}
	if r.Dev != "" {
		args = append(args, "--interface", r.Dev)
	}
	if r.Table != "" {
		args = append(args, "--table", r.Table)
	}
	return args, nil
}
