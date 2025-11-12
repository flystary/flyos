package routing

import "strconv"

// PBRRule represents a policy-based routing rule and associated route.
type PBRRule struct {
	BaseRoute

	// FwMark is the firewall mark used to match packets (from iptables).
	FwMark uint32

	// Priority is the rule priority in the ip rule chain.
	Priority int

	// From is the source CIDR to match (optional).
	From string

	// To is the destination CIDR to match (optional).
	To string

	// Iif is the input interface name (optional).
	Iif string
}

func init() {
	RegisterRoute("pbr", defaultEnabled, NewPBRRule)
}

func NewPBRRule() Route {
	return &PBRRule{
		BaseRoute: BaseRoute{Table: "main", Scope: "global"},
	}
}

func (r *PBRRule) Validate() error {
	r.Proto = "pbr"
	return r.ValidateBase()
}

func (r *PBRRule) GetPrefix() string { return r.Prefix }
func (r *PBRRule) GetProto() string  { return r.Proto }

func (r *PBRRule) ToArgs() ([]string, error) {
	args := []string{
		"--id", strconv.Itoa(r.Priority),
		"--table", r.Table,
	}
	if r.From != "" {
		args = append(args, "--src-cidr", r.From)
	}
	if r.To != "" {
		args = append(args, "--dst-cidr", r.To)
	}
	if r.Proto != "" {
		args = append(args, "--protocol", r.Proto)
	}
	if r.Priority > 0 {
		args = append(args, "--priority", strconv.Itoa(r.Priority))
	}
	return args, nil
}
