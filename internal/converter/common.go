package converter

import (
	"flyos/pkg/dsl"
	"fmt"
)

// MapConverter 把 DSL Command 转成通用 map[string]interface{}
// 适用于快速启动和单元测试，后续可替换为具体类型的 Converter 实现。
type MapConverter struct{}

func NewMapConverter() *MapConverter { return &MapConverter{} }

// FromDSL 将一条命令转成 map（包含 Attrs 与 Subtype）
func (c *MapConverter) FromDSL(cmd *dsl.Command) (map[string]interface{}, error) {
	if cmd == nil {
		return nil, fmt.Errorf("nil command")
	}
	out := map[string]interface{}{}
	out["kind"] = cmd.Kind
	out["verb"] = cmd.Verb
	out["subtype"] = cmd.Subtype

	// shallow copy attrs
	m := map[string]interface{}{}
	for k, v := range cmd.Attrs {
		m[k] = v
	}
	out["attrs"] = m

	// blocks (for sync) -> list of maps
	if len(cmd.Blocks) > 0 {
		var bls []map[string]interface{}
		for _, b := range cmd.Blocks {
			sub, _ := c.FromDSL(&b)
			bls = append(bls, sub)
		}
		out["blocks"] = bls
	}
	return out, nil
}

// FromDSLBatch 批量转换（包含 sync 自动展开或按需要返回 blocks）
func (c *MapConverter) FromDSLBatch(cmds []dsl.Command) ([]map[string]interface{}, error) {
	var res []map[string]interface{}
	for _, cm := range cmds {
		if cm.Verb == "sync" && len(cm.Blocks) > 0 {
			for _, b := range cm.Blocks {
				m, err := c.FromDSL(&b)
				if err != nil {
					return nil, err
				}
				res = append(res, m)
			}
		} else {
			m, err := c.FromDSL(&cm)
			if err != nil {
				return nil, err
			}
			res = append(res, m)
		}
	}
	return res, nil
}
