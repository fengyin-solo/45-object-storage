package handler

import (
	"net/http"

	"objectstore/internal/model"
	"objectstore/internal/service"
	"objectstore/pkg/httpx"
)

// registerMultipartUploadRoutes 注册分片上传相关路由。
func (s *Server) registerMultipartUploadRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/multipart-uploads", s.initMultipartUpload)
	mux.HandleFunc("GET /api/multipart-uploads", s.listMultipartUploads)
	mux.HandleFunc("GET /api/multipart-uploads/{id}", s.getMultipartUpload)
	mux.HandleFunc("POST /api/multipart-uploads/{id}/parts", s.uploadPart)
	mux.HandleFunc("GET /api/multipart-uploads/{id}/parts", s.listUploadParts)
	mux.HandleFunc("POST /api/multipart-uploads/{id}/complete", s.completeMultipartUpload)
	mux.HandleFunc("POST /api/multipart-uploads/{id}/abort", s.abortMultipartUpload)
}

type initMultipartUploadRequest struct {
	BucketID string `json:"bucket_id"`
	Key      string `json:"key"`
}

func (s *Server) initMultipartUpload(w http.ResponseWriter, r *http.Request) {
	var req initMultipartUploadRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	u, err := s.svc.InitMultipartUpload(req.BucketID, req.Key)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, u)
}

func (s *Server) listMultipartUploads(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.MultipartUploadFilter{
		BucketID: r.URL.Query().Get("bucket_id"),
		Status:   r.URL.Query().Get("status"),
		Keyword:  r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListMultipartUploads(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getMultipartUpload(w http.ResponseWriter, r *http.Request) {
	u, err := s.svc.GetMultipartUpload(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, u)
}

type uploadPartRequest struct {
	PartNumber int    `json:"part_number"`
	Size       int64  `json:"size"`
	ETag       string `json:"etag"`
}

func (s *Server) uploadPart(w http.ResponseWriter, r *http.Request) {
	var req uploadPartRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.UploadPart(pathID(r), req.PartNumber, req.Size, req.ETag)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, p)
}

type completeMultipartUploadRequest struct {
	Parts []service.CompletePart `json:"parts"`
}

func (s *Server) completeMultipartUpload(w http.ResponseWriter, r *http.Request) {
	var req completeMultipartUploadRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	u, err := s.svc.CompleteMultipartUpload(pathID(r), req.Parts)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, u)
}

func (s *Server) abortMultipartUpload(w http.ResponseWriter, r *http.Request) {
	u, err := s.svc.AbortMultipartUpload(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, u)
}

func (s *Server) listUploadParts(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 50, s.maxPageSize())
	filter := model.UploadPartFilter{
		UploadID: pathID(r),
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
