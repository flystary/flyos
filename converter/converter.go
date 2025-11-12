package converter

import "flyos/dsl"

// Converter 定义 DSL Command → 业务对象的转换接口
type Converter[T any] interface {
	// FromDSL 将单条 DSL Command 转为业务对象
	FromDSL(cmd *dsl.Command) (T, error)

	// FromDSLBatch 批量转换
	FromDSLBatch(cmds []dsl.Command) ([]T, error)
}
