package handler

import (
	"net/http"
	"strconv"

	"objectstore/internal/model"
	"objectstore/pkg/httpx"
)

// registerBucketRoutes 注册存储桶相关路由。
func (s *Server) registerBucketRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/buckets", s.createBucket)
	mux.HandleFunc("GET /api/buckets", s.listBuckets)
	mux.HandleFunc("GET /api/buckets/{id}", s.getBucket)
	mux.HandleFunc("PUT /api/buckets/{id}", s.updateBucket)
	mux.HandleFunc("DELETE /api/buckets/{id}", s.deleteBucket)
	mux.HandleFunc("POST /api/buckets/{id}/suspend", s.suspendBucket)
	mux.HandleFunc("POST /api/buckets/{id}/activate", s.activateBucket)
	mux.HandleFunc("PUT /api/buckets/{id}/quota", s.setBucketQuota)
}

type createBucketRequest struct {
	Name       string `json:"name"`
	Region     string `json:"region"`
	Owner      string `json:"owner"`
	QuotaBytes int64  `json:"quota_bytes"`
}

func (s *Server) createBucket(w http.ResponseWriter, r *http.Request) {
	var req createBucketRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	b, err := s.svc.CreateBucket(model.Bucket{
		Name:       req.Name,
		Region:     req.Region,
		Owner:      req.Owner,
		QuotaBytes: req.QuotaBytes,
		Status:     model.BucketActive,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, b)
}

func (s *Server) listBuckets(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.BucketFilter{
		Region:  r.URL.Query().Get("region"),
		Owner:   r.URL.Query().Get("owner"),
		Status:  r.URL.Query().Get("status"),
		Keyword: r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListBuckets(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getBucket(w http.ResponseWriter, r *http.Request) {
	b, err := s.svc.GetBucket(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, b)
}

type updateBucketRequest struct {
	Region     string `json:"region"`
	Owner      string `json:"owner"`
	QuotaBytes int64  `json:"quota_bytes"`
}

func (s *Server) updateBucket(w http.ResponseWriter, r *http.Request) {
	var req updateBucketRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	b, err := s.svc.UpdateBucket(pathID(r), model.Bucket{
		Region:     req.Region,
		Owner:      req.Owner,
		QuotaBytes: req.QuotaBytes,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, b)
}

func (s *Server) deleteBucket(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteBucket(pathID(r)); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

func (s *Server) suspendBucket(w http.ResponseWriter, r *http.Request) {
	b, err := s.svc.SuspendBucket(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, b)
}

func (s *Server) activateBucket(w http.ResponseWriter, r *http.Request) {
	b, err := s.svc.ActivateBucket(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, b)
}

type setBucketQuotaRequest struct {
	QuotaBytes int64 `json:"quota_bytes"`
}

func (s *Server) setBucketQuota(w http.ResponseWriter, r *http.Request) {
	var req setBucketQuotaRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	q, err := s.svc.SetBucketQuota(pathID(r), req.QuotaBytes)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, q)
}

// queryInt64 从查询串解析 int64，解析失败返回默认值。
func queryInt64(r *http.Request, key string, def int64) int64 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}
