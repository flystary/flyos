package converter

import (
	"fmt"
	"strings"

	"flyos/dsl"
	"flyos/routing"
)

type ACLConverter struct{}

func NewACLConverter() *ACLConverter {
	return &ACLConverter{}
}

func (c *ACLConverter) FromDSL(cmd *dsl.Command) (*routing.ACLRule, error) {
	if strings.ToLower(cmd.Kind) != "acl" {
		return nil, fmt.Errorf("invalid command kind: %s", cmd.Kind)
	}
	r := &routing.ACLRule{
		Subtype: cmd.Subtype,
		Attrs:   cmd.Attrs,
	}
	return r, nil
}

func (c *ACLConverter) FromDSLBatch(cmds []dsl.Command) ([]*routing.ACLRule, error) {
	var rules []*routing.ACLRule
	for _, cdm := range cmds {
		if strings.ToLower(cdm.Kind) != "acl" {
			continue
		}
		if cdm.Verb == "sync" {
			for _, b := range cdm.Blocks {
				r, err := c.FromDSL(&b)
				if err != nil {
					return nil, err
				}
				rules = append(rules, r)
			}
		} else {
			r, err := c.FromDSL(&cdm)
			if err != nil {
				return nil, err
			}
			rules = append(rules, r)
		}
	}
	return rules, nil
}
