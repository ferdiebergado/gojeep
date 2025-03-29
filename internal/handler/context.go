package handler

import "context"

type ctxKey int

const paramsCtxKey ctxKey = 1

func NewParamsContext[T any](ctx context.Context, t T) context.Context {
	return context.WithValue(ctx, paramsCtxKey, t)
}

func FromParamsContext[T any](ctx context.Context) (any, T, bool) {
	ctxVal := ctx.Value(paramsCtxKey)
	t, ok := ctxVal.(T)
	return ctxVal, t, ok
}
