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

// GetExecutorContext 获取执行上下文
func GetExecutorContext(ctx context.Context) *ExecutorContext {
	if val, ok := ctx.Value(executorContextKey{}).(*ExecutorContext); ok {
		return val
	}
	return nil
}

// SetExecutorContext 添加执行上下文
func SetExecutorContext(ctx context.Context, ec *ExecutorContext) context.Context {
	return context.WithValue(ctx, executorContextKey{}, ec)
}

// Hook sql 执行中间件
type Hook interface {
	// Before sql 执行前
	Before(ctx context.Context, ec *ExecutorContext) error
	// After sql 执行后
	After(ctx context.Context, ec *ExecutorContext, er *ExecutorResult)
}

// AfterHandler sql 执行后回调
type AfterHandler func(ctx context.Context, ec *ExecutorContext, er *ExecutorResult)

// Chain 中间件执行链
type Chain struct {
	hooks []Hook
}

// NewHookChain 创建 hooks Chain
func NewHookChain(hooks ...Hook) *Chain {
	return &Chain{hooks: hooks}
}

// Before 执行前
func (c *Chain) Before(ctx context.Context, ec *ExecutorContext) error {
	// 第一个先执行
	for i := range c.hooks {
		err := c.hooks[i].Before(ctx, ec)
		if err != nil {
			return err
		}
	}
	return nil
}

// After 执行后
func (c *Chain) After(ctx context.Context, ec *ExecutorContext, er *ExecutorResult) {
	// 最后一个包装的先执行
	ln := len(c.hooks)
	for i := ln - 1; i >= 0; i-- {
		c.hooks[i].After(ctx, ec, er)
	}
}
