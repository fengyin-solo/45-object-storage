// Package app 负责依赖装配。
package app

import (
	"net/http"

	"objectstore/internal/config"
	"objectstore/internal/handler"
	"objectstore/internal/service"
	"objectstore/internal/store"
	"objectstore/pkg/logger"
)

// App 应用装配结果。
type App struct {
	server *handler.Server
}

// New 装配存储、业务逻辑与处理器。
func New(cfg *config.Config, log *logger.Logger) (*App, error) {
	st := store.NewMemoryStore()
	svc := service.New(st, log, cfg)
	server := handler.NewServer(svc, log, cfg)
	log.Infof("应用装配完成，配置：%s", cfg.String())
	return &App{server: server}, nil
}

// Routes 返回应用根路由处理器。
func (a *App) Routes() http.Handler { return a.server.Routes() }
