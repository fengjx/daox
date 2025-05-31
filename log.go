package daox

import (
	"context"

	"github.com/fengjx/daox/v2/engine"
)

// LogHook 打印日志中间件
type LogHook struct {
	Print engine.AfterHandler
}

// NewLogHook 创建打印日志中间件
func NewLogHook(p engine.AfterHandler) *LogHook {
	return &LogHook{Print: p}
}

// Before 执行前
func (l LogHook) Before(ctx context.Context, ec *engine.ExecutorContext) error {
	return nil
}

// After 执行后
func (l LogHook) After(ctx context.Context, ec *engine.ExecutorContext, er *engine.ExecutorResult) {
	l.Print(ctx, ec, er)
}
