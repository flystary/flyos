package converter

import (
	"flyos/pkg/module"
	"fmt"
)

// Converter 接口：将 JSON 或 DSL Command 转换成模块对象
type Converter interface {
	// ConvertFromJSON 将 REST/MCP JSON 数据转换为模块对象
	ConvertFromJSON(data map[string]interface{}) (module.ModuleObject, error)
	// ConvertFromDSL 将 DSL Command 转换为模块对象
	ConvertFromDSL(cmd interface{}) (module.ModuleObject, error)
}

// 全局注册表
var converters = map[string]Converter{}

// Register 注册 Converter
func Register(kind string, c Converter) {
	converters[kind] = c
}

// Get 返回已注册的 Converter
func Get(kind string) (Converter, error) {
	c, ok := converters[kind]
	if !ok {
		return nil, fmt.Errorf("converter not found for kind '%s'", kind)
	}
	return c, nil
}
