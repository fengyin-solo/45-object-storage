package handler

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"objectstore/pkg/httpx"
)

// apiGuardMiddleware 对 /api/ 路径统一做鉴权 + 限流，其余路径（静态页面）直接放行。
func (s *Server) apiGuardMiddleware(next http.Handler) http.Handler {
	expected := ""
	if s.cfg != nil {
		expected = s.cfg.APIKey
	}
	limit := 600
	if s.cfg != nil && s.cfg.RateLimitPerMinute > 0 {
		limit = s.cfg.RateLimitPerMinute
	}
	rl := &rateLimiter{limit: limit}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		if expected == "" || r.Header.Get("X-API-Key") != expected {
			httpx.Unauthorized(w, "无效的 API Key")
			return
		}
		if !rl.allow() {
			httpx.TooManyRequests(w, "请求过于频繁，请稍后再试")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimiter 简单固定窗口限流器。
type rateLimiter struct {
	mu     sync.Mutex
	window int64
	count  int
	limit  int
}

// allow 判断当前请求是否放行，并推进窗口计数。
func (r *rateLimiter) allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().Unix()
	if now > r.window {
		r.window = now
		r.count = 0
	}
	if r.count >= r.limit {
		return false
	}
	r.count++
	return true
}
