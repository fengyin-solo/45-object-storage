package handler

import (
	"net/http"
	"strconv"

	"objectstore/internal/model"
	"objectstore/pkg/httpx"
)

// registerUploadPartRoutes 注册分片相关路由。
func (s *Server) registerUploadPartRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/upload-parts", s.listUploadPart)
	mux.HandleFunc("GET /api/upload-parts/{id}", s.getUploadPart)
}

func (s *Server) listUploadPart(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 50, s.maxPageSize())
	filter := model.UploadPartFilter{
		UploadID: r.URL.Query().Get("upload_id"),
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("min_part_number")); err == nil {
		filter.MinPartNumber = v
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("max_part_number")); err == nil {
		filter.MaxPartNumber = v
	}
	items, total, err := s.svc.ListUploadParts(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getUploadPart(w http.ResponseWriter, r *http.Request) {
	p, err := s.svc.GetUploadPart(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}
