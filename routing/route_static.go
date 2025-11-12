package routing

import (
	"net"
	"strconv"
)

// StaticRoute represents a static kernel route.
type StaticRoute struct {
	BaseRoute

	// Track enables health monitoring for this route (FlyOS extension).
	Track bool
}

func init() {
	RegisterRoute("static", defaultEnabled, NewStaticRoute)
}

func NewStaticRoute() Route {
	return &StaticRoute{
		BaseRoute: BaseRoute{
			Table: "main",
			Scope: "global",
		},
	}
}
func (r *StaticRoute) Validate() error {
	r.Proto = "static"
	return r.ValidateBase()
}

func (r *StaticRoute) GetPrefix() string { return r.Prefix }
func (r *StaticRoute) GetProto() string  { return r.Proto }

func (r *StaticRoute) ToArgs() ([]string, error) {
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
	if r.Via != "" {
		args = append(args, "--nexthop", r.Via)
	}
	if r.Dev != "" {
		args = append(args, "--interface", r.Dev)
	}
	args = append(args, "--track", strconv.FormatBool(r.Track))
	return args, nil
}
