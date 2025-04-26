package handler

import "context"

type ctxKey int

const (
	paramsCtxKey ctxKey = iota
	userCtxKey
)

func NewParamsContext[T any](ctx context.Context, t T) context.Context {
	return context.WithValue(ctx, paramsCtxKey, t)
}

func FromParamsContext[T any](ctx context.Context) (any, T, bool) {
	ctxVal := ctx.Value(paramsCtxKey)
	t, ok := ctxVal.(T)
	return ctxVal, t, ok
}

func NewUserContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userCtxKey, userID)
}

func FromUserContext(ctx context.Context) (string, bool) {
	ctxVal := ctx.Value(userCtxKey)
	userID, ok := ctxVal.(string)
	return userID, ok
}
