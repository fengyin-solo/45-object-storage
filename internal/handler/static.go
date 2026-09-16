package handler

import (
	"net/http"
	"os"
)

// registerStaticRoutes 挂载前端静态页面服务。
func (s *Server) registerStaticRoutes(mux *http.ServeMux) {
	// 若 web 目录不存在则退化为返回提示，避免启动失败。
	if _, err := os.Stat("web"); err == nil {
		mux.Handle("GET /", http.FileServer(http.Dir("web")))
		return
	}
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_, _ = w.Write([]byte("objectstore 服务运行中，前端资源目录 web/ 不存在。"))
			return
		}
		http.NotFound(w, r)
	})
}
