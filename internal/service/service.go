// Package service 实现业务逻辑层。
package service

import (
	"objectstore/internal/config"
	"objectstore/internal/store"
	"objectstore/pkg/logger"
)

// Service 聚合全部实体的业务逻辑。
type Service struct {
	store store.Store
	log   *logger.Logger
	cfg   *config.Config
}

// New 构造业务逻辑服务。
func New(st store.Store, log *logger.Logger, cfg *config.Config) *Service {
	return &Service{store: st, log: log, cfg: cfg}
}

// Store 暴露底层存储（供批量操作等内部方法复用）。
func (s *Service) Store() store.Store { return s.store }

// maxPageSize 返回最大分页大小。
func (s *Service) maxPageSize() int {
	if s.cfg != nil && s.cfg.MaxPageSize > 0 {
		return s.cfg.MaxPageSize
	}
	return 100
}
