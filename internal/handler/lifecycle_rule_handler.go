package handler

import (
	"net/http"

	"objectstore/internal/model"
	"objectstore/pkg/httpx"
)

// registerLifecycleRuleRoutes 注册生命周期规则相关路由。
func (s *Server) registerLifecycleRuleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/lifecycle-rules", s.createLifecycleRule)
	mux.HandleFunc("GET /api/lifecycle-rules", s.listLifecycleRules)
	mux.HandleFunc("GET /api/lifecycle-rules/{id}", s.getLifecycleRule)
	mux.HandleFunc("PUT /api/lifecycle-rules/{id}", s.updateLifecycleRule)
	mux.HandleFunc("DELETE /api/lifecycle-rules/{id}", s.deleteLifecycleRule)
	mux.HandleFunc("POST /api/lifecycle-rules/{id}/status", s.setLifecycleRuleStatus)
	mux.HandleFunc("POST /api/lifecycle-rules/batch-status", s.batchSetLifecycleRuleStatus)
}

type createLifecycleRuleRequest struct {
	BucketID       string `json:"bucket_id"`
	Prefix         string `json:"prefix"`
	ExpireDays     int    `json:"expire_days"`
	TransitionDays int    `json:"transition_days"`
	Status         string `json:"status"`
}

func (s *Server) createLifecycleRule(w http.ResponseWriter, r *http.Request) {
	var req createLifecycleRuleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	rule, err := s.svc.CreateLifecycleRule(model.LifecycleRule{
		BucketID:       req.BucketID,
		Prefix:         req.Prefix,
		ExpireDays:     req.ExpireDays,
		TransitionDays: req.TransitionDays,
		Status:         req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, rule)
}

func (s *Server) listLifecycleRules(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.LifecycleRuleFilter{
		BucketID: r.URL.Query().Get("bucket_id"),
		Status:   r.URL.Query().Get("status"),
		Prefix:   r.URL.Query().Get("prefix"),
	}
	items, total, err := s.svc.ListLifecycleRules(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getLifecycleRule(w http.ResponseWriter, r *http.Request) {
	rule, err := s.svc.GetLifecycleRule(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, rule)
}

type updateLifecycleRuleRequest struct {
	Prefix         string `json:"prefix"`
	ExpireDays     int    `json:"expire_days"`
	TransitionDays int    `json:"transition_days"`
}

func (s *Server) updateLifecycleRule(w http.ResponseWriter, r *http.Request) {
	var req updateLifecycleRuleRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	rule, err := s.svc.UpdateLifecycleRule(pathID(r), model.LifecycleRule{
		Prefix:         req.Prefix,
		ExpireDays:     req.ExpireDays,
		TransitionDays: req.TransitionDays,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, rule)
}

func (s *Server) deleteLifecycleRule(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteLifecycleRule(pathID(r)); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type setLifecycleRuleStatusRequest struct {
	Status string `json:"status"`
}

func (s *Server) setLifecycleRuleStatus(w http.ResponseWriter, r *http.Request) {
	var req setLifecycleRuleStatusRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	rule, err := s.svc.SetLifecycleRuleStatus(pathID(r), req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, rule)
}

type batchSetLifecycleRuleStatusRequest struct {
	IDs    []string `json:"ids"`
	Status string   `json:"status"`
}

func (s *Server) batchSetLifecycleRuleStatus(w http.ResponseWriter, r *http.Request) {
	var req batchSetLifecycleRuleStatusRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	if len(req.IDs) == 0 {
		httpx.BadRequest(w, "ids 不能为空")
		return
	}
	updated, err := s.svc.BatchUpdateLifecycleRuleStatus(req.IDs, req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]int{"updated": updated})
}
