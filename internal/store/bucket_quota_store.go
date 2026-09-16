package store

import (
	"objectstore/internal/model"
)

// CreateBucketQuota 创建配额记录。
func (s *MemoryStore) CreateBucketQuota(q *model.BucketQuota) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.bucketQuotas {
		if exist.BucketID == q.BucketID {
			return ErrConflict
		}
	}
	s.bucketQuotas[q.ID] = q
	return nil
}

// GetBucketQuota 按 ID 查询配额记录。
func (s *MemoryStore) GetBucketQuota(id string) (*model.BucketQuota, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	q, ok := s.bucketQuotas[id]
	if !ok {
		return nil, ErrNotFound
	}
	return q, nil
}

// GetBucketQuotaByBucket 按桶 ID 查询配额记录。
func (s *MemoryStore) GetBucketQuotaByBucket(bucketID string) (*model.BucketQuota, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, q := range s.bucketQuotas {
		if q.BucketID == bucketID {
			return q, nil
		}
	}
	return nil, ErrNotFound
}

// ListBucketQuotas 返回全部配额记录。
func (s *MemoryStore) ListBucketQuotas() []*model.BucketQuota {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.BucketQuota, 0, len(s.bucketQuotas))
	for _, q := range s.bucketQuotas {
		list = append(list, q)
	}
	return list
}

// UpdateBucketQuota 更新配额记录。
func (s *MemoryStore) UpdateBucketQuota(q *model.BucketQuota) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.bucketQuotas[q.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.bucketQuotas {
		if exist.ID != q.ID && exist.BucketID == q.BucketID {
			return ErrConflict
		}
	}
	s.bucketQuotas[q.ID] = q
	return nil
}

// DeleteBucketQuota 删除配额记录。
func (s *MemoryStore) DeleteBucketQuota(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.bucketQuotas[id]; !ok {
		return ErrNotFound
	}
	delete(s.bucketQuotas, id)
	return nil
}
