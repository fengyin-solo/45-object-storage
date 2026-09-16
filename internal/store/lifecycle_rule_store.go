package store

import (
	"objectstore/internal/model"
)

// CreateLifecycleRule 创建生命周期规则。
func (s *MemoryStore) CreateLifecycleRule(r *model.LifecycleRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lifecycleRules[r.ID] = r
	return nil
}

// GetLifecycleRule 按 ID 查询生命周期规则。
func (s *MemoryStore) GetLifecycleRule(id string) (*model.LifecycleRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.lifecycleRules[id]
	if !ok {
		return nil, ErrNotFound
	}
	return r, nil
}

// ListLifecycleRules 返回全部生命周期规则。
func (s *MemoryStore) ListLifecycleRules() []*model.LifecycleRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.LifecycleRule, 0, len(s.lifecycleRules))
	for _, r := range s.lifecycleRules {
		list = append(list, r)
	}
	return list
}

// UpdateLifecycleRule 更新生命周期规则。
func (s *MemoryStore) UpdateLifecycleRule(r *model.LifecycleRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.lifecycleRules[r.ID]; !ok {
		return ErrNotFound
	}
	s.lifecycleRules[r.ID] = r
	return nil
}

// DeleteLifecycleRule 删除生命周期规则。
func (s *MemoryStore) DeleteLifecycleRule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.lifecycleRules[id]; !ok {
		return ErrNotFound
	}
	delete(s.lifecycleRules, id)
	return nil
}
