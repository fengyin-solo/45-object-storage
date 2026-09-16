package store

import (
	"objectstore/internal/model"
)

// CreateUploadPart 创建分片，同一上传内分片编号唯一。
func (s *MemoryStore) CreateUploadPart(p *model.UploadPart) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.uploadParts {
		if exist.UploadID == p.UploadID && exist.PartNumber == p.PartNumber {
			return ErrConflict
		}
	}
	s.uploadParts[p.ID] = p
	return nil
}

// GetUploadPart 按 ID 查询分片。
func (s *MemoryStore) GetUploadPart(id string) (*model.UploadPart, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.uploadParts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

// ListUploadParts 返回全部分片。
func (s *MemoryStore) ListUploadParts() []*model.UploadPart {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.UploadPart, 0, len(s.uploadParts))
	for _, p := range s.uploadParts {
		list = append(list, p)
	}
	return list
}

// DeleteUploadPart 删除分片。
func (s *MemoryStore) DeleteUploadPart(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.uploadParts[id]; !ok {
		return ErrNotFound
	}
	delete(s.uploadParts, id)
	return nil
}
