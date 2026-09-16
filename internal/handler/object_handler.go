package handler

import (
	"net/http"

	"objectstore/internal/model"
	"objectstore/internal/service"
	"objectstore/pkg/httpx"
)

// registerObjectRoutes 注册对象相关路由。
func (s *Server) registerObjectRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/objects", s.putObject)
	mux.HandleFunc("GET /api/objects", s.listObjects)
	mux.HandleFunc("GET /api/objects/{id}", s.getObject)
	mux.HandleFunc("PUT /api/objects/{id}", s.updateObject)
	mux.HandleFunc("DELETE /api/objects/{id}", s.deleteObject)
	mux.HandleFunc("POST /api/objects/batch-delete", s.batchDeleteObjects)
	mux.HandleFunc("POST /api/objects/batch-put", s.batchPutObjects)
}

type putObjectRequest struct {
	BucketID    string `json:"bucket_id"`
	Key         string `json:"key"`
	ContentType string `json:"content_type"`
	ETag        string `json:"etag"`
	Size        int64  `json:"size"`
}

func (s *Server) putObject(w http.ResponseWriter, r *http.Request) {
	var req putObjectRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	o, err := s.svc.PutObject(req.BucketID, req.Key, req.ContentType, req.ETag, req.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, o)
}

func (s *Server) listObjects(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ObjectFilter{
		BucketID: r.URL.Query().Get("bucket_id"),
		Status:   r.URL.Query().Get("status"),
		Keyword:  r.URL.Query().Get("keyword"),
		MinSize:  queryInt64(r, "min_size", 0),
		MaxSize:  queryInt64(r, "max_size", 0),
	}
	items, total, err := s.svc.ListObjects(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getObject(w http.ResponseWriter, r *http.Request) {
	o, err := s.svc.GetObject(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, o)
}

type updateObjectRequest struct {
	ContentType string `json:"content_type"`
	ETag        string `json:"etag"`
	Size        int64  `json:"size"`
}

func (s *Server) updateObject(w http.ResponseWriter, r *http.Request) {
	var req updateObjectRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	o, err := s.svc.UpdateObject(pathID(r), req.ContentType, req.ETag, req.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, o)
}

func (s *Server) deleteObject(w http.ResponseWriter, r *http.Request) {
	o, err := s.svc.DeleteObject(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, o)
}

type batchDeleteObjectsRequest struct {
	IDs []string `json:"ids"`
}

func (s *Server) batchDeleteObjects(w http.ResponseWriter, r *http.Request) {
	var req batchDeleteObjectsRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	if len(req.IDs) == 0 {
		httpx.BadRequest(w, "ids 不能为空")
		return
	}
	deleted, err := s.svc.BatchDeleteObjects(req.IDs)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]int{"deleted": deleted})
}

type batchPutObjectsRequest struct {
	BucketID string                   `json:"bucket_id"`
	Items    []service.PutObjectInput `json:"items"`
}

func (s *Server) batchPutObjects(w http.ResponseWriter, r *http.Request) {
	var req batchPutObjectsRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	if req.BucketID == "" {
		httpx.BadRequest(w, "bucket_id 不能为空")
		return
	}
	if len(req.Items) == 0 {
		httpx.BadRequest(w, "items 不能为空")
		return
	}
	created, err := s.svc.BatchPutObjects(req.BucketID, req.Items)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]int{"created": created})
}
