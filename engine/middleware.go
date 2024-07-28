package engine

import (
	"context"
	"time"
)

type executorContextKey struct{}

// ExecutorContext SQL 执行器上下文
type ExecutorContext struct {
	Type      SQLType
	TableName string
	SQL       string
	Args      []any
	NameArgs  any
	Start     time.Time
}

// ExecutorResult 执行结果
type ExecutorResult struct {
	Err       error         // 执行异常异常
	Affected  int64         // 新增、删除、修改返回的影响行数
	QueryRows int64         // 查询记录行数
	Duration  time.Duration // 耗时
}

// ExecutorCtx 获取执行上下文
func ExecutorCtx(ctx context.Context) *ExecutorContext {
	if val, ok := ctx.Value(executorContextKey{}).(*ExecutorContext); ok {
		return val
	}
	return nil
}

// WithExecutorCtx 添加执行上下文
func WithExecutorCtx(ctx context.Context, ec *ExecutorContext) context.Context {
	return context.WithValue(ctx, executorContextKey{}, ec)
}

// Middleware sql 执行中间件
type Middleware interface {
	// Before sql 执行前
	Before(ctx context.Context, ec *ExecutorContext) error
	// After sql 执行后
	After(ctx context.Context, ec *ExecutorContext, er *ExecutorResult)
}

// Chain 中间件执行链
type Chain struct {
	middlewares []Middleware
}

// NewChain 创建 middleware Chain
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{middlewares: middlewares}
}

// Before 执行前
func (c *Chain) Before(ctx context.Context, ec *ExecutorContext) error {
	// 最后一个包装的先执行
	ln := len(c.middlewares)
	for i := ln - 1; i >= 0; i-- {
		err := c.middlewares[i].Before(ctx, ec)
		if err != nil {
			return err
		}
	}
	return nil
}

// After 执行后
func (c *Chain) After(ctx context.Context, ec *ExecutorContext, er *ExecutorResult) {
	// 第一个先执行
	for _, middleware := range c.middlewares {
		middleware.After(ctx, ec, er)
	}
}

// Print sql 打印
type Print func(ctx context.Context, ec *ExecutorContext, er *ExecutorResult)

// LogMiddleware 打印日志中间件
type LogMiddleware struct {
	Print Print
}

// NewLogMiddleware 创建打印日志中间件
func NewLogMiddleware(p Print) *LogMiddleware {
	return &LogMiddleware{Print: p}
}

// Before 执行前
func (l LogMiddleware) Before(ctx context.Context, ec *ExecutorContext) error {
	return nil
}

// After 执行后
func (l LogMiddleware) After(ctx context.Context, ec *ExecutorContext, er *ExecutorResult) {
	l.Print(ctx, ec, er)
}
