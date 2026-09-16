package handler

import (
	"net/http"

	"objectstore/internal/model"
	"objectstore/pkg/httpx"
)

// registerBucketPolicyRoutes 注册桶策略相关路由。
func (s *Server) registerBucketPolicyRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/bucket-policies", s.createBucketPolicy)
	mux.HandleFunc("GET /api/bucket-policies", s.listBucketPolicies)
	mux.HandleFunc("GET /api/bucket-policies/{id}", s.getBucketPolicy)
	mux.HandleFunc("PUT /api/bucket-policies/{id}", s.updateBucketPolicy)
	mux.HandleFunc("DELETE /api/bucket-policies/{id}", s.deleteBucketPolicy)
	mux.HandleFunc("POST /api/bucket-policies/{id}/status", s.setBucketPolicyStatus)
	mux.HandleFunc("POST /api/bucket-policies/evaluate", s.evaluateBucketPolicy)
}

type createBucketPolicyRequest struct {
	BucketID  string `json:"bucket_id"`
	Principal string `json:"principal"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	Effect    string `json:"effect"`
}

func (s *Server) createBucketPolicy(w http.ResponseWriter, r *http.Request) {
	var req createBucketPolicyRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.CreateBucketPolicy(model.BucketPolicy{
		BucketID:  req.BucketID,
		Principal: req.Principal,
		Action:    req.Action,
		Resource:  req.Resource,
		Effect:    req.Effect,
		Status:    model.PolicyEnabled,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, p)
}

func (s *Server) listBucketPolicies(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.BucketPolicyFilter{
		BucketID:  r.URL.Query().Get("bucket_id"),
		Principal: r.URL.Query().Get("principal"),
		Effect:    r.URL.Query().Get("effect"),
		Status:    r.URL.Query().Get("status"),
	}
	items, total, err := s.svc.ListBucketPolicies(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getBucketPolicy(w http.ResponseWriter, r *http.Request) {
	p, err := s.svc.GetBucketPolicy(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

type updateBucketPolicyRequest struct {
	Principal string `json:"principal"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	Effect    string `json:"effect"`
}

func (s *Server) updateBucketPolicy(w http.ResponseWriter, r *http.Request) {
	var req updateBucketPolicyRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.UpdateBucketPolicy(pathID(r), model.BucketPolicy{
		Principal: req.Principal,
		Action:    req.Action,
		Resource:  req.Resource,
		Effect:    req.Effect,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

func (s *Server) deleteBucketPolicy(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteBucketPolicy(pathID(r)); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type setBucketPolicyStatusRequest struct {
	Status string `json:"status"`
}

func (s *Server) setBucketPolicyStatus(w http.ResponseWriter, r *http.Request) {
	var req setBucketPolicyStatusRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.SetBucketPolicyStatus(pathID(r), req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

type evaluateBucketPolicyRequest struct {
	BucketID  string `json:"bucket_id"`
	Principal string `json:"principal"`
	Action    string `json:"action"`
}

func (s *Server) evaluateBucketPolicy(w http.ResponseWriter, r *http.Request) {
	var req evaluateBucketPolicyRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	allowed, reason := s.svc.EvaluateBucketPolicy(req.BucketID, req.Principal, req.Action)
	httpx.OK(w, map[string]interface{}{
		"allowed": allowed,
		"reason":  reason,
	})
}
