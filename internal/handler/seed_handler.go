package handler

import (
	"net/http"

	"objectstore/pkg/httpx"
)

// registerSeedRoutes 注册演示数据初始化路由。
func (s *Server) registerSeedRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/seed", s.seedDemoData)
}

func (s *Server) seedDemoData(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.SeedDemoData(); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, s.svc.Overview())
}
