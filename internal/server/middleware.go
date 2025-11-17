package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/cursed-ninja/internal-transfers-system/internal/utils"
	"go.uber.org/zap"
)

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := newRequestID()
		logger := utils.GetLogger(s.cfg.Env)
		ctx := context.WithValue(r.Context(), utils.LoggerContextKey, logger)
		ctx, _ = utils.LoggerWithKey(ctx, zap.String("request_id", reqID))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(buf)
}
