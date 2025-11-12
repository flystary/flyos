package routing

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type RouteManager interface {
	Name() string
	Add(r Route) error
	Set(r Route) error
	Remove(r Route) error
	List() ([]Route, error)
	Sync(routes []Route) error
}

type CLIManager struct{}

func (m *CLIManager) doOp(op string, r Route) error {
	if err := r.Validate(); err != nil {
		return fmt.Errorf("invalid route: %w", err)
	}
	args, err := r.ToArgs()
	if err != nil {
		return fmt.Errorf("build args: %w", err)
	}
	cmd := exec.Command(fmt.Sprintf("%s_ipv4_%s_route", op, r.GetProto()), args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exec failed: %w; output: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (m *CLIManager) Add(r Route) error    { return m.doOp("add", r) }
func (m *CLIManager) Set(r Route) error    { return m.doOp("set", r) }
func (m *CLIManager) Remove(r Route) error { return m.doOp("remove", r) }

func (m *CLIManager) List() ([]Route, error) {
	var all []Route
	for _, typ := range []string{"static", "ospf", "bgp", "pbr"} {
		cmd := exec.Command(fmt.Sprintf("list_ipv4_%s_route", typ))
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		var routes []json.RawMessage
		if err := json.Unmarshal(out, &routes); err != nil {
			continue
		}
		for _, raw := range routes {
			var r Route
			switch typ {
			case "static":
				r = &StaticRoute{}
			case "ospf":
				r = &OSPFRoute{}
			case "bgp":
				r = &BGPRoute{}
			case "pbr":
				r = &PBRRule{}
			}
			if err := json.Unmarshal(raw, r); err == nil && r.Validate() == nil {
				all = append(all, r)
			}
		}
	}
	return all, nil
}

func (m *CLIManager) Sync(routes []Route) error {
	grouped := make(map[string][]Route)
	for _, r := range routes {
		if err := r.Validate(); err != nil {
			return err
		}
		grouped[r.GetProto()] = append(grouped[r.GetProto()], r)
	}
	for typ, rs := range grouped {
		data, _ := json.Marshal(rs)
		cmd := exec.Command(fmt.Sprintf("sync_ipv4_%s_route", typ))
		cmd.Stdin = strings.NewReader(string(data))
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
