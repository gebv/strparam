package httprouter

import "context"

type ctxKey uint

const (
	_ ctxKey = iota
	parsedParamsCtxKey
)

func setParsedParamsToCtx(ctx context.Context, params map[string]string) context.Context {
	return context.WithValue(ctx, parsedParamsCtxKey, params)
}

// ParsedParamsFromCtx returns map[string]string from context if sets or empty map.
func ParsedParamsFromCtx(ctx context.Context) map[string]string {
	vali := ctx.Value(parsedParamsCtxKey)
	if vali == nil {
		return map[string]string{}
	}
	val, ok := vali.(map[string]string)
	if !ok {
		return map[string]string{}
	}
	return val
}
