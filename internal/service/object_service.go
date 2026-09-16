package service

import (
	"sort"
	"time"

	"objectstore/internal/model"
	"objectstore/internal/store"
	"objectstore/pkg/idgen"
)

// PutObject 写入（或覆盖）对象：同一 Key 覆盖写入会生成新版本。
func (s *Service) PutObject(bucketID, key, contentType, etag string, size int64) (*model.Object, error) {
	if _, err := s.store.GetBucket(bucketID); err != nil {
		return nil, err
	}
	now := time.Now()
	existing, err := s.store.GetObjectByKey(bucketID, key)
	if err == store.ErrNotFound {
		return s.createObject(bucketID, key, contentType, etag, size, now)
	}
	if err != nil {
		return nil, err
	}
	// 覆盖写入：更新对象元数据并生成新版本。
	existing.Size = size
	existing.ContentType = contentType
	existing.ETag = etag
	existing.Status = model.ObjectActive
	existing.UpdatedAt = now
	if err := existing.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateObject(existing); err != nil {
		return nil, err
	}
	if err := s.appendVersion(existing, size, etag, false, now); err != nil {
		return nil, err
	}
	s.adjustQuota(bucketID, size-existing.Size)
	return existing, nil
}

func (s *Service) createObject(bucketID, key, contentType, etag string, size int64, now time.Time) (*model.Object, error) {
	o := &model.Object{
		ID:          idgen.Hex(),
		BucketID:    bucketID,
		Key:         key,
		Size:        size,
		ContentType: contentType,
		ETag:        etag,
		Status:      model.ObjectActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := o.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.CreateObject(o); err != nil {
		return nil, err
	}
	if err := s.appendVersion(o, size, etag, false, now); err != nil {
		return nil, err
	}
	s.adjustQuota(bucketID, size)
	return o, nil
}

// appendVersion 为对象追加一个版本记录，版本号在对象内递增。
func (s *Service) appendVersion(o *model.Object, size int64, etag string, isDeleteMarker bool, now time.Time) error {
	next := 1
	for _, v := range s.store.ListObjectVersions() {
		if v.ObjectID == o.ID && v.VersionNo >= next {
			next = v.VersionNo + 1
		}
	}
	v := &model.ObjectVersion{
		ID:             idgen.Hex(),
		ObjectID:       o.ID,
		BucketID:       o.BucketID,
		VersionNo:      next,
		Size:           size,
		ETag:           etag,
		IsDeleteMarker: isDeleteMarker,
		CreatedAt:      now,
	}
	return s.store.CreateObjectVersion(v)
}

// GetObject 按 ID 查询对象。
func (s *Service) GetObject(id string) (*model.Object, error) {
	return s.store.GetObject(id)
}

// ListObjects 分页 + 多条件筛选查询对象。
func (s *Service) ListObjects(filter model.ObjectFilter, page, size int) ([]*model.Object, int, error) {
	all := s.store.ListObjects()
	matched := make([]*model.Object, 0, len(all))
	for _, o := range all {
		if filter.Match(o) {
			matched = append(matched, o)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Object{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateObject 更新对象元数据（Key 不可变）。
func (s *Service) UpdateObject(id, contentType, etag string, size int64) (*model.Object, error) {
	o, err := s.store.GetObject(id)
	if err != nil {
		return nil, err
	}
	if o.Status != model.ObjectActive {
		return nil, model.NewValidationError("status", "已删除对象不可更新")
	}
	oldSize := o.Size
	o.ContentType = contentType
	o.ETag = etag
	o.Size = size
	o.UpdatedAt = time.Now()
	if err := o.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateObject(o); err != nil {
		return nil, err
	}
	s.adjustQuota(o.BucketID, size-oldSize)
	return o, nil
}

// DeleteObject 软删除对象（active→deleted），并生成 delete marker 版本。
func (s *Service) DeleteObject(id string) (*model.Object, error) {
	o, err := s.store.GetObject(id)
	if err != nil {
		return nil, err
	}
	if !model.ObjectCanTransition(o.Status, model.ObjectDeleted) {
		return nil, store.ErrStateTransition
	}
	o.Status = model.ObjectDeleted
	o.UpdatedAt = time.Now()
	if err := s.store.UpdateObject(o); err != nil {
		return nil, err
	}
	_ = s.appendVersion(o, 0, "", true, time.Now())
	s.adjustQuota(o.BucketID, -o.Size)
	return o, nil
}

// BatchDeleteObjects 批量删除对象（软删除，跳过已删除的）。
func (s *Service) BatchDeleteObjects(ids []string) (int, error) {
	deleted := 0
	for _, id := range ids {
		if _, err := s.DeleteObject(id); err == nil {
			deleted++
		}
	}
	return deleted, nil
}

// PutObjectInput 批量写入对象时的单项入参。
type PutObjectInput struct {
	Key         string `json:"key"`
	ContentType string `json:"content_type"`
	ETag        string `json:"etag"`
	Size        int64  `json:"size"`
}

// BatchPutObjects 批量写入对象（键冲突的项跳过）。
func (s *Service) BatchPutObjects(bucketID string, items []PutObjectInput) (int, error) {
	if _, err := s.store.GetBucket(bucketID); err != nil {
		return 0, err
	}
	created := 0
	for _, it := range items {
		if _, err := s.PutObject(bucketID, it.Key, it.ContentType, it.ETag, it.Size); err == nil {
			created++
		}
	}
	return created, nil
}

// adjustQuota 调整桶的已用容量。
func (s *Service) adjustQuota(bucketID string, delta int64) {
	if delta == 0 {
		return
	}
	q, err := s.store.GetBucketQuotaByBucket(bucketID)
	if err != nil {
		return
	}
	q.UsedBytes += delta
	if q.UsedBytes < 0 {
		q.UsedBytes = 0
	}
	q.UpdatedAt = time.Now()
	_ = s.store.UpdateBucketQuota(q)
}
