package handler

import (
	"net/http"

	"objectstore/internal/model"
	"objectstore/pkg/httpx"
)

// registerAccessLogRoutes 注册访问日志相关路由。
func (s *Server) registerAccessLogRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/access-logs", s.createAccessLog)
	mux.HandleFunc("GET /api/access-logs", s.listAccessLogs)
	mux.HandleFunc("GET /api/access-logs/{id}", s.getAccessLog)
	mux.HandleFunc("DELETE /api/access-logs/{id}", s.deleteAccessLog)
}

type createAccessLogRequest struct {
	BucketID  string `json:"bucket_id"`
	ObjectKey string `json:"object_key"`
	Operation string `json:"operation"`
	IP        string `json:"ip"`
	Result    string `json:"result"`
	Size      int64  `json:"size"`
}

func (s *Server) createAccessLog(w http.ResponseWriter, r *http.Request) {
	var req createAccessLogRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	l, err := s.svc.RecordAccess(req.BucketID, req.ObjectKey, req.Operation, req.IP, req.Result, req.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, l)
}

func (s *Server) listAccessLogs(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.AccessLogFilter{
		BucketID:  r.URL.Query().Get("bucket_id"),
		ObjectKey: r.URL.Query().Get("object_key"),
		Operation: r.URL.Query().Get("operation"),
		IP:        r.URL.Query().Get("ip"),
		Result:    r.URL.Query().Get("result"),
	}
	items, total, err := s.svc.ListAccessLogs(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getAccessLog(w http.ResponseWriter, r *http.Request) {
	l, err := s.svc.GetAccessLog(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, l)
}

func (s *Server) deleteAccessLog(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteAccessLog(pathID(r)); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
