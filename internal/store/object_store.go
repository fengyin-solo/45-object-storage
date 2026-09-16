package store

import (
	"objectstore/internal/model"
)

// CreateObject 创建对象，同一桶内 Key 唯一。
func (s *MemoryStore) CreateObject(o *model.Object) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.objects {
		if exist.BucketID == o.BucketID && exist.Key == o.Key {
			return ErrConflict
		}
	}
	s.objects[o.ID] = o
	return nil
}

// GetObject 按 ID 查询对象。
func (s *MemoryStore) GetObject(id string) (*model.Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.objects[id]
	if !ok {
		return nil, ErrNotFound
	}
	return o, nil
}

// GetObjectByKey 按桶 ID + 对象键查询对象。
func (s *MemoryStore) GetObjectByKey(bucketID, key string) (*model.Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, o := range s.objects {
		if o.BucketID == bucketID && o.Key == key {
			return o, nil
		}
	}
	return nil, ErrNotFound
}

// ListObjects 返回全部对象。
func (s *MemoryStore) ListObjects() []*model.Object {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Object, 0, len(s.objects))
	for _, o := range s.objects {
		list = append(list, o)
	}
	return list
}

// UpdateObject 更新对象，桶内 Key 唯一性校验排除自身。
func (s *MemoryStore) UpdateObject(o *model.Object) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.objects[o.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.objects {
		if exist.ID != o.ID && exist.BucketID == o.BucketID && exist.Key == o.Key {
			return ErrConflict
		}
	}
	s.objects[o.ID] = o
	return nil
}

// DeleteObject 删除对象。
func (s *MemoryStore) DeleteObject(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.objects[id]; !ok {
		return ErrNotFound
	}
	delete(s.objects, id)
	return nil
}
