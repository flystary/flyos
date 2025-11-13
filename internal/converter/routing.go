package converter

import (
	"flyos/modules/routing"
	"flyos/pkg/module"
	"fmt"
)

func init() {
	Register("route", &RouteConverter{})
}

// 默认 RouteConverter
type RouteConverter struct{}

func (c *RouteConverter) ConvertFromJSON(data map[string]interface{}) (module.ModuleObject, error) {
	subtype, ok := data["subtype"].(string)
	if !ok {
		return nil, fmt.Errorf("missing subtype in route data")
	}

	var route routing.Route
	switch subtype {
	case "static":
		route = &routing.StaticRoute{}
	case "bgp":
		route = &routing.BGPRoute{}
	case "ospf":
		route = &routing.OSPFRoute{}
	case "pbr":
		route = &routing.PBRRule{}
	default:
		return nil, fmt.Errorf("unsupported route subtype: %s", subtype)
	}

	if err := route.LoadFromMap(data); err != nil {
		return nil, err
	}

	return &RouteExecutor{route: route}, nil
}

func (c *RouteConverter) ConvertFromDSL(cmd interface{}) (module.ModuleObject, error) {
	// 这里 cmd 可以是 flyos/pkg/dsl.Command
	dslCmd, ok := cmd.(interface {
		Subtype() string
		Attrs() map[string]interface{}
	})
	if !ok {
		return nil, fmt.Errorf("invalid DSL command type")
	}

	data := dslCmd.Attrs()
	data["subtype"] = dslCmd.Subtype()
	return c.ConvertFromJSON(data)
}
