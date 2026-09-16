package store

import (
	"objectstore/internal/model"
)

// CreateBucketPolicy 创建桶策略。
func (s *MemoryStore) CreateBucketPolicy(p *model.BucketPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bucketPolicies[p.ID] = p
	return nil
}

// GetBucketPolicy 按 ID 查询桶策略。
func (s *MemoryStore) GetBucketPolicy(id string) (*model.BucketPolicy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.bucketPolicies[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

// ListBucketPolicies 返回全部桶策略。
func (s *MemoryStore) ListBucketPolicies() []*model.BucketPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.BucketPolicy, 0, len(s.bucketPolicies))
	for _, p := range s.bucketPolicies {
		list = append(list, p)
	}
	return list
}

// UpdateBucketPolicy 更新桶策略。
func (s *MemoryStore) UpdateBucketPolicy(p *model.BucketPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.bucketPolicies[p.ID]; !ok {
		return ErrNotFound
	}
	s.bucketPolicies[p.ID] = p
	return nil
}

// DeleteBucketPolicy 删除桶策略。
func (s *MemoryStore) DeleteBucketPolicy(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.bucketPolicies[id]; !ok {
		return ErrNotFound
	}
	delete(s.bucketPolicies, id)
	return nil
}
