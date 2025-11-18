package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/cursed-ninja/internal-transfers-system/internal/utils"
	"go.uber.org/zap"
)

// loggingMiddleware attaches a request ID and logger to each incoming HTTP request's context.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := newRequestID()
		logger := utils.GetLogger(s.cfg.Env)
		ctx := context.WithValue(r.Context(), utils.LoggerContextKey, logger)
		ctx, _ = utils.LoggerWithKey(ctx, zap.String("request_id", reqID))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// newRequestID generates a new unique request ID based on the current timestamp.
func newRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}
