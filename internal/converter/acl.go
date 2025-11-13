package converter

import (
	"flyos/pkg/module"
	"fmt"
)

func init() {
	Register("acl", &ACLConverter{})
}

// 默认 ACL Converter
type ACLConverter struct{}

func (c *ACLConverter) ConvertFromJSON(data map[string]interface{}) (module.ModuleObject, error) {
	aclObj := &acl.ACLExecutor{}
	if err := aclObj.LoadFromMap(data); err != nil {
		return nil, err
	}
	return aclObj, nil
}

func (c *ACLConverter) ConvertFromDSL(cmd interface{}) (module.ModuleObject, error) {
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
