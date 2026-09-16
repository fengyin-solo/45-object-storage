package handler

import (
	"net/http"
	"strconv"

	"objectstore/pkg/httpx"
)

// registerLifecycleExecutorRoutes 注册生命周期执行相关路由。
func (s *Server) registerLifecycleExecutorRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/lifecycle-rules/execute", s.executeLifecycle)
	mux.HandleFunc("POST /api/lifecycle-rules/execute-all", s.executeAllLifecycles)
	mux.HandleFunc("GET /api/lifecycle-rules/expire-candidates", s.expireCandidates)
}

type executeLifecycleRequest struct {
	BucketID string `json:"bucket_id"`
}

func (s *Server) executeLifecycle(w http.ResponseWriter, r *http.Request) {
	var req executeLifecycleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	if req.BucketID == "" {
		httpx.BadRequest(w, "bucket_id 不能为空")
		return
	}
	res, err := s.svc.ExecuteLifecycle(req.BucketID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, res)
}

func (s *Server) executeAllLifecycles(w http.ResponseWriter, r *http.Request) {
	res, err := s.svc.ExecuteAllLifecycles()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, res)
}

func (s *Server) expireCandidates(w http.ResponseWriter, r *http.Request) {
	bucketID := r.URL.Query().Get("bucket_id")
	if bucketID == "" {
		httpx.BadRequest(w, "bucket_id 不能为空")
		return
	}
	days := 0
	if v, err := strconv.Atoi(r.URL.Query().Get("expire_days")); err == nil {
		days = v
	}
	httpx.OK(w, s.svc.ExpireCandidates(bucketID, days))
}
