// Package handler 实现 HTTP 处理器层。
package handler

import (
	"errors"
	"net/http"
	"runtime/debug"
	"time"

	"objectstore/internal/config"
	"objectstore/internal/model"
	"objectstore/internal/service"
	"objectstore/internal/store"
	"objectstore/pkg/httpx"
	"objectstore/pkg/logger"
)

// Server HTTP 处理器服务。
type Server struct {
	svc *service.Service
	log *logger.Logger
	cfg *config.Config
}

// NewServer 构造处理器服务。
func NewServer(svc *service.Service, log *logger.Logger, cfg *config.Config) *Server {
	return &Server{svc: svc, log: log, cfg: cfg}
}

// Routes 注册全部路由并包裹中间件。
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// API 路由直接注册到顶层 mux，鉴权 + 限流由 apiGuardMiddleware 统一处理。
	s.registerBucketRoutes(mux)
	s.registerObjectRoutes(mux)
	s.registerObjectVersionRoutes(mux)
	s.registerLifecycleRuleRoutes(mux)
	s.registerMultipartUploadRoutes(mux)
	s.registerUploadPartRoutes(mux)
	s.registerBucketPolicyRoutes(mux)
	s.registerAccessLogRoutes(mux)
	s.registerBucketQuotaRoutes(mux)
	s.registerLifecycleExecutorRoutes(mux)
	s.registerObjectRestoreRoutes(mux)
	s.registerSeedRoutes(mux)
	s.registerStatsRoutes(mux)

	// 前端静态页面。
	s.registerStaticRoutes(mux)

	return s.loggingMiddleware(s.recoveryMiddleware(s.apiGuardMiddleware(mux)))
}

// maxPageSize 返回最大分页大小。
func (s *Server) maxPageSize() int {
	if s.cfg != nil && s.cfg.MaxPageSize > 0 {
		return s.cfg.MaxPageSize
	}
	return 100
}

// loggingMiddleware 请求日志中间件。
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.log.Infof("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// recoveryMiddleware panic 恢复中间件。
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.log.Errorf("panic: %v\n%s", rec, debug.Stack())
				httpx.InternalError(w, "服务器内部错误")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// writeServiceError 将业务错误统一映射为 HTTP 响应。
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case model.IsValidationError(err):
		httpx.BadRequest(w, err.Error())
	case errors.Is(err, store.ErrNotFound):
		httpx.NotFound(w, err.Error())
	case errors.Is(err, store.ErrStateTransition):
		httpx.Conflict(w, err.Error())
	case errors.Is(err, store.ErrConflict):
		httpx.Conflict(w, err.Error())
	default:
		httpx.InternalError(w, err.Error())
	}
}

// pathID 从请求路径提取 {id} 参数。
func pathID(r *http.Request) string {
	return r.PathValue("id")
}
