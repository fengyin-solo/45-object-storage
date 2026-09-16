package store

import (
	"objectstore/internal/model"
)

// CreateBucket 创建存储桶，名称全局唯一。
func (s *MemoryStore) CreateBucket(b *model.Bucket) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.buckets {
		if exist.Name == b.Name {
			return ErrConflict
		}
	}
	s.buckets[b.ID] = b
	return nil
}

// GetBucket 按 ID 查询存储桶。
func (s *MemoryStore) GetBucket(id string) (*model.Bucket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.buckets[id]
	if !ok {
		return nil, ErrNotFound
	}
	return b, nil
}

// GetBucketByName 按名称查询存储桶。
func (s *MemoryStore) GetBucketByName(name string) (*model.Bucket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.buckets {
		if b.Name == name {
			return b, nil
		}
	}
	return nil, ErrNotFound
}

// ListBuckets 返回全部存储桶。
func (s *MemoryStore) ListBuckets() []*model.Bucket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Bucket, 0, len(s.buckets))
	for _, b := range s.buckets {
		list = append(list, b)
	}
	return list
}

// UpdateBucket 更新存储桶，名称唯一性校验排除自身。
func (s *MemoryStore) UpdateBucket(b *model.Bucket) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.buckets[b.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.buckets {
		if exist.ID != b.ID && exist.Name == b.Name {
			return ErrConflict
		}
	}
	s.buckets[b.ID] = b
	return nil
}

// DeleteBucket 删除存储桶。
func (s *MemoryStore) DeleteBucket(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.buckets[id]; !ok {
		return ErrNotFound
	}
	delete(s.buckets, id)
	return nil
}
