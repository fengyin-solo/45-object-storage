package handler

import (
	"net/http"
	"strconv"

	"objectstore/pkg/httpx"
)

// registerStatsRoutes 注册统计与导出接口。
func (s *Server) registerStatsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/stats/overview", s.statsOverview)
	mux.HandleFunc("GET /api/stats/by-bucket", s.statsByBucket)
	mux.HandleFunc("GET /api/stats/by-region", s.statsByRegion)
	mux.HandleFunc("GET /api/stats/top-access", s.statsTopAccess)
	mux.HandleFunc("GET /api/stats/lifecycle-count", s.statsLifecycleCount)
	mux.HandleFunc("GET /api/stats/operation-stats", s.statsOperationStats)
	mux.HandleFunc("GET /api/stats/content-types", s.statsContentTypes)
	mux.HandleFunc("GET /api/stats/top-objects", s.statsTopObjects)
	mux.HandleFunc("GET /api/stats/daily-trend", s.statsDailyTrend)
	mux.HandleFunc("GET /api/stats/upload-status", s.statsUploadStatus)
	mux.HandleFunc("GET /api/stats/bucket-growth", s.statsBucketGrowth)
	mux.HandleFunc("GET /api/stats/version-count", s.statsVersionCount)
	mux.HandleFunc("GET /api/export", s.exportSnapshot)
}

func (s *Server) statsOverview(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.Overview())
}

func (s *Server) statsByBucket(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.StatsByBucket())
}

func (s *Server) statsByRegion(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.StatsByRegion())
}

func (s *Server) statsTopAccess(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 {
		limit = v
	}
	httpx.OK(w, s.svc.TopAccessBuckets(limit))
}

func (s *Server) statsLifecycleCount(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.LifecycleRuleCount())
}

func (s *Server) exportSnapshot(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.ExportSnapshot())
}

func (s *Server) statsOperationStats(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.OperationStats())
}

func (s *Server) statsContentTypes(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.ContentTypeDistribution())
}

func (s *Server) statsTopObjects(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 {
		limit = v
	}
	httpx.OK(w, s.svc.TopObjectsBySize(limit))
}

func (s *Server) statsDailyTrend(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.DailyAccessTrend())
}

func (s *Server) statsUploadStatus(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.UploadStatusDistribution())
}

func (s *Server) statsBucketGrowth(w http.ResponseWriter, r *http.Request) {
	httpx.OK(w, s.svc.BucketGrowth())
}

func (s *Server) statsVersionCount(w http.ResponseWriter, r *http.Request) {
	total, markers := s.svc.VersionCount()
	httpx.OK(w, map[string]int{"total": total, "delete_markers": markers})
}
