package converter

import (
	"fmt"
	"strings"

	"flyos/pkg/dsl"
	"flyos/modules/routing"
)

type RouteConverter struct{}

func NewRouteConverter() *RouteConverter {
	return &RouteConverter{}
}

func (c *RouteConverter) FromDSL(cmd *dsl.Command) (routing.Route, error) {
	var r routing.Route
	switch strings.ToLower(cmd.Subtype) {
	case "static":
		r = &routing.StaticRoute{}
	case "ospf":
		r = &routing.OSPFRoute{}
	case "bgp":
		r = &routing.BGPRoute{}
	case "pbr":
		r = &routing.PBRRule{}
	default:
		return nil, fmt.Errorf("unknown route subtype: %s", cmd.Subtype)
	}

	base := r.(routing.BaseRoute)
	if prefix, ok := cmd.Attrs["prefix"].(string); ok {
		base.SetPrefix(prefix)
	}
	if via, ok := cmd.Attrs["via"].(string); ok {
		base.SetVia(via)
	}
	if dev, ok := cmd.Attrs["dev"].(string); ok {
		base.SetDev(dev)
	}
	if table, ok := cmd.Attrs["table"].(string); ok {
		base.SetTable(table)
	}
	if metric, ok := cmd.Attrs["metric"].(int); ok {
		base.SetMetric(metric)
	}

	// BGP 扩展属性
	if bgp, ok := r.(*routing.BGPRoute); ok {
		if lp, ok := cmd.Attrs["local_pref"].(int); ok {
			bgp.LocalPref = uint32(lp)
		}
		if comms, ok := cmd.Attrs["community"].([]string); ok {
			for _, s := range comms {
				if c, err := routing.ParseCommunity(s); err == nil {
					bgp.Communities = append(bgp.Communities, c)
				}
			}
		}
	}

	return r, nil
}

func (c *RouteConverter) FromDSLBatch(cmds []dsl.Command) ([]routing.Route, error) {
	var routes []routing.Route
	for _, cdm := range cmds {
		if strings.ToLower(cdm.Kind) != "route" {
			continue
		}
		if cdm.Verb == "sync" {
			for _, b := range cdm.Blocks {
				r, err := c.FromDSL(&b)
				if err != nil {
					return nil, err
				}
				routes = append(routes, r)
			}
		} else {
			r, err := c.FromDSL(&cdm)
			if err != nil {
				return nil, err
			}
			routes = append(routes, r)
		}
	}
	return routes, nil
}
