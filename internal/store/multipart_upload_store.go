package store

import (
	"objectstore/internal/model"
)

// CreateMultipartUpload 创建分片上传。
func (s *MemoryStore) CreateMultipartUpload(u *model.MultipartUpload) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.multipartUploads[u.ID] = u
	return nil
}

// GetMultipartUpload 按 ID 查询分片上传。
func (s *MemoryStore) GetMultipartUpload(id string) (*model.MultipartUpload, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.multipartUploads[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

// ListMultipartUploads 返回全部分片上传。
func (s *MemoryStore) ListMultipartUploads() []*model.MultipartUpload {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.MultipartUpload, 0, len(s.multipartUploads))
	for _, u := range s.multipartUploads {
		list = append(list, u)
	}
	return list
}

// UpdateMultipartUpload 更新分片上传。
func (s *MemoryStore) UpdateMultipartUpload(u *model.MultipartUpload) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.multipartUploads[u.ID]; !ok {
		return ErrNotFound
	}
	s.multipartUploads[u.ID] = u
	return nil
}

// DeleteMultipartUpload 删除分片上传。
func (s *MemoryStore) DeleteMultipartUpload(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.multipartUploads[id]; !ok {
		return ErrNotFound
	}
	delete(s.multipartUploads, id)
	return nil
}
