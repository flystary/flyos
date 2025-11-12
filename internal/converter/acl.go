package converter

import (
	"fmt"
	"strings"

	"flyos/pkg/dsl"
	"flyos/modules/acl"
)

type ACLConverter struct{}

func NewACLConverter() *ACLConverter {
	return &ACLConverter{}
}

func (c *ACLConverter) FromDSL(cmd *dsl.Command) (*acl.ACLRule, error) {
	if strings.ToLower(cmd.Kind) != "acl" {
		return nil, fmt.Errorf("invalid command kind: %s", cmd.Kind)
	}
	r := &acl.ACLRule{
		Subtype: cmd.Subtype,
		Attrs:   cmd.Attrs,
	}
	return r, nil
}

func (c *ACLConverter) FromDSLBatch(cmds []dsl.Command) ([]*acl.ACLRule, error) {
	var rules []*acl.ACLRule
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
