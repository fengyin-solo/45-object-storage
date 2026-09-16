package handler

import (
	"net/http"

	"objectstore/pkg/httpx"
)

// registerObjectRestoreRoutes 注册对象版本恢复相关路由。
func (s *Server) registerObjectRestoreRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/objects/{id}/restore", s.restoreObjectVersion)
	mux.HandleFunc("POST /api/objects/{id}/purge-versions", s.purgeObjectVersions)
	mux.HandleFunc("GET /api/objects/{id}/latest-active-version", s.latestActiveVersion)
}

type restoreObjectVersionRequest struct {
	VersionID string `json:"version_id"`
}

func (s *Server) restoreObjectVersion(w http.ResponseWriter, r *http.Request) {
	var req restoreObjectVersionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	if req.VersionID == "" {
		httpx.BadRequest(w, "version_id 不能为空")
		return
	}
	o, err := s.svc.RestoreObjectVersion(pathID(r), req.VersionID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, o)
}

type purgeObjectVersionsRequest struct {
	Keep int `json:"keep"`
}

func (s *Server) purgeObjectVersions(w http.ResponseWriter, r *http.Request) {
	var req purgeObjectVersionsRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	removed, err := s.svc.PurgeOldVersions(pathID(r), req.Keep)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]int{"removed": removed})
}

func (s *Server) latestActiveVersion(w http.ResponseWriter, r *http.Request) {
	v, err := s.svc.LatestActiveVersion(pathID(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, v)
}
