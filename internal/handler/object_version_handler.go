package handler

import (
	"net/http"
	"strconv"

	"objectstore/internal/model"
	"objectstore/pkg/httpx"
)

// registerObjectVersionRoutes 注册对象版本相关路由。
func (s *Server) registerObjectVersionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/object-versions", s.listObjectVersions)
	mux.HandleFunc("GET /api/object-versions/{id}", s.getObjectVersion)
}

func (s *Server) listObjectVersions(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ObjectVersionFilter{
		ObjectID: r.URL.Query().Get("object_id"),
		BucketID: r.URL.Query().Get("bucket_id"),
	}
	if v, err := strconv.ParseBool(r.URL.Query().Get("latest_only")); err == nil {
		filter.LatestOnly = v
	}
	items, total, err := s.svc.ListObjectVersions(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getObjectVersion(w http.ResponseWriter, r *http.Request) {
	v, err := s.svc.GetObjectVersion(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, v)
}
