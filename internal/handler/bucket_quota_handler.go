package handler

import (
	"net/http"
	"strconv"

	"objectstore/internal/model"
	"objectstore/pkg/httpx"
)

// registerBucketQuotaRoutes 注册配额相关路由。
func (s *Server) registerBucketQuotaRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/bucket-quotas", s.createBucketQuota)
	mux.HandleFunc("GET /api/bucket-quotas", s.listBucketQuotas)
	mux.HandleFunc("GET /api/bucket-quotas/{id}", s.getBucketQuota)
	mux.HandleFunc("PUT /api/bucket-quotas/{id}", s.updateBucketQuota)
	mux.HandleFunc("DELETE /api/bucket-quotas/{id}", s.deleteBucketQuota)
	mux.HandleFunc("POST /api/bucket-quotas/recalculate", s.recalculateQuotas)
	mux.HandleFunc("GET /api/bucket-quotas/exceeded", s.listExceededQuotas)
}

type createBucketQuotaRequest struct {
	BucketID   string `json:"bucket_id"`
	UsedBytes  int64  `json:"used_bytes"`
	QuotaBytes int64  `json:"quota_bytes"`
}

func (s *Server) createBucketQuota(w http.ResponseWriter, r *http.Request) {
	var req createBucketQuotaRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	q, err := s.svc.CreateBucketQuota(model.BucketQuota{
		BucketID:   req.BucketID,
		UsedBytes:  req.UsedBytes,
		QuotaBytes: req.QuotaBytes,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, q)
}

func (s *Server) listBucketQuotas(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.BucketQuotaFilter{
		BucketID: r.URL.Query().Get("bucket_id"),
	}
	if v, err := strconv.ParseBool(r.URL.Query().Get("exceeded_only")); err == nil {
		filter.ExceededOnly = v
	}
	items, total, err := s.svc.ListBucketQuotas(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getBucketQuota(w http.ResponseWriter, r *http.Request) {
	q, err := s.svc.GetBucketQuota(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, q)
}

type updateBucketQuotaRequest struct {
	UsedBytes  int64 `json:"used_bytes"`
	QuotaBytes int64 `json:"quota_bytes"`
}

func (s *Server) updateBucketQuota(w http.ResponseWriter, r *http.Request) {
	var req updateBucketQuotaRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	q, err := s.svc.UpdateBucketQuota(pathID(r), model.BucketQuota{
		UsedBytes:  req.UsedBytes,
		QuotaBytes: req.QuotaBytes,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, q)
}

func (s *Server) deleteBucketQuota(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteBucketQuota(pathID(r)); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) recalculateQuotas(w http.ResponseWriter, r *http.Request) {
	count, err := s.svc.RecalculateQuotas()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]int{"recalculated": count})
}

func (s *Server) listExceededQuotas(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.QuotaExceededBuckets())
}
