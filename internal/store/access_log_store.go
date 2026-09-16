package store

import (
	"objectstore/internal/model"
)

// CreateAccessLog 创建访问日志。
func (s *MemoryStore) CreateAccessLog(l *model.AccessLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessLogs[l.ID] = l
	return nil
}

// GetAccessLog 按 ID 查询访问日志。
func (s *MemoryStore) GetAccessLog(id string) (*model.AccessLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.accessLogs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return l, nil
}

// ListAccessLogs 返回全部访问日志。
func (s *MemoryStore) ListAccessLogs() []*model.AccessLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.AccessLog, 0, len(s.accessLogs))
	for _, l := range s.accessLogs {
		list = append(list, l)
	}
	return list
}

// DeleteAccessLog 删除访问日志。
func (s *MemoryStore) DeleteAccessLog(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.accessLogs[id]; !ok {
		return ErrNotFound
	}
	delete(s.accessLogs, id)
	return nil
}
