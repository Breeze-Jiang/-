package httpx

import "context"

type contextKey string

const requestIDKey contextKey = "request-id"
const userIDKey contextKey = "user-id"
const sessionIDKey contextKey = "session-id"
const clientIPKey contextKey = "client-ip"

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}
func RequestID(ctx context.Context) string { v, _ := ctx.Value(requestIDKey).(string); return v }
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}
func UserID(ctx context.Context) (string, bool) { v, ok := ctx.Value(userIDKey).(string); return v, ok }
func WithSessionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, sessionIDKey, id)
}
func SessionID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(sessionIDKey).(string)
	return v, ok
}
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey, ip)
}
func ClientIP(ctx context.Context) string { v, _ := ctx.Value(clientIPKey).(string); return v }
