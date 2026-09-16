package store

import (
	"objectstore/internal/model"
)

// CreateObjectVersion 创建对象版本。
func (s *MemoryStore) CreateObjectVersion(v *model.ObjectVersion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objectVersions[v.ID] = v
	return nil
}

// GetObjectVersion 按 ID 查询对象版本。
func (s *MemoryStore) GetObjectVersion(id string) (*model.ObjectVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.objectVersions[id]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

// ListObjectVersions 返回全部对象版本。
func (s *MemoryStore) ListObjectVersions() []*model.ObjectVersion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.ObjectVersion, 0, len(s.objectVersions))
	for _, v := range s.objectVersions {
		list = append(list, v)
	}
	return list
}

// DeleteObjectVersion 删除对象版本。
func (s *MemoryStore) DeleteObjectVersion(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.objectVersions[id]; !ok {
		return ErrNotFound
	}
	delete(s.objectVersions, id)
	return nil
}
