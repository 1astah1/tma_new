package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"tma-backend/internal/handler"
)

func Recover() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()
					slog.Error("panic recovered",
						slog.String("request_id", GetRequestID(r)),
						slog.Any("error", err),
						slog.String("stack", string(stack)),
					)
					handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
