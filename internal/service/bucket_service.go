package service

import (
	"sort"
	"time"

	"objectstore/internal/model"
	"objectstore/internal/store"
	"objectstore/pkg/idgen"
)

// CreateBucket 创建存储桶，并同步初始化其配额记录。
func (s *Service) CreateBucket(input model.Bucket) (*model.Bucket, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetBucketByName(input.Name); err == nil {
		return nil, store.ErrConflict
	}
	now := time.Now()
	b := &model.Bucket{
		ID:         idgen.Hex(),
		Name:       input.Name,
		Region:     input.Region,
		Owner:      input.Owner,
		QuotaBytes: input.QuotaBytes,
		Status:     input.Status,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.store.CreateBucket(b); err != nil {
		return nil, err
	}
	// 初始化配额记录（配额为 0 表示不限制）。
	if _, err := s.store.GetBucketQuotaByBucket(b.ID); err != nil {
		q := &model.BucketQuota{
			ID:         idgen.Hex(),
			BucketID:   b.ID,
			UsedBytes:  0,
			QuotaBytes: input.QuotaBytes,
			UpdatedAt:  now,
		}
		_ = s.store.CreateBucketQuota(q)
	}
	return b, nil
}

// GetBucket 按 ID 查询存储桶。
func (s *Service) GetBucket(id string) (*model.Bucket, error) {
	return s.store.GetBucket(id)
}

// ListBuckets 分页 + 多条件筛选查询存储桶。
func (s *Service) ListBuckets(filter model.BucketFilter, page, size int) ([]*model.Bucket, int, error) {
	all := s.store.ListBuckets()
	matched := make([]*model.Bucket, 0, len(all))
	for _, b := range all {
		if filter.Match(b) {
			matched = append(matched, b)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Bucket{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateBucket 更新存储桶可编辑字段（区域、所有者、配额）。
func (s *Service) UpdateBucket(id string, input model.Bucket) (*model.Bucket, error) {
	exist, err := s.store.GetBucket(id)
	if err != nil {
		return nil, err
	}
	exist.Region = input.Region
	exist.Owner = input.Owner
	if input.QuotaBytes >= 0 {
		exist.QuotaBytes = input.QuotaBytes
	}
	if err := exist.Validate(); err != nil {
		return nil, err
	}
	exist.UpdatedAt = time.Now()
	if err := s.store.UpdateBucket(exist); err != nil {
		return nil, err
	}
	// 同步配额记录。
	if q, err := s.store.GetBucketQuotaByBucket(id); err == nil {
		q.QuotaBytes = exist.QuotaBytes
		q.UpdatedAt = time.Now()
		_ = s.store.UpdateBucketQuota(q)
	}
	return exist, nil
}

// SuspendBucket 挂起存储桶（active→suspended）。
func (s *Service) SuspendBucket(id string) (*model.Bucket, error) {
	return s.transitionBucket(id, model.BucketSuspended)
}

// ActivateBucket 恢复存储桶（suspended→active）。
func (s *Service) ActivateBucket(id string) (*model.Bucket, error) {
	return s.transitionBucket(id, model.BucketActive)
}

func (s *Service) transitionBucket(id, to string) (*model.Bucket, error) {
	b, err := s.store.GetBucket(id)
	if err != nil {
		return nil, err
	}
	if !model.BucketCanTransition(b.Status, to) {
		return nil, store.ErrStateTransition
	}
	b.Status = to
	b.UpdatedAt = time.Now()
	if err := s.store.UpdateBucket(b); err != nil {
		return nil, err
	}
	return b, nil
}

// DeleteBucket 删除存储桶（仅当桶内无对象）。
func (s *Service) DeleteBucket(id string) error {
	if _, err := s.store.GetBucket(id); err != nil {
		return err
	}
	for _, o := range s.store.ListObjects() {
		if o.BucketID == id {
			return model.NewValidationError("bucket", "桶内仍有对象，无法删除")
		}
	}
	return s.store.DeleteBucket(id)
}

// SetBucketQuota 调整桶配额并同步。
func (s *Service) SetBucketQuota(bucketID string, quotaBytes int64) (*model.BucketQuota, error) {
	if _, err := s.store.GetBucket(bucketID); err != nil {
		return nil, err
	}
	if quotaBytes < 0 {
		return nil, model.NewValidationError("quota_bytes", "配额容量不能为负")
	}
	q, err := s.store.GetBucketQuotaByBucket(bucketID)
	if err != nil {
		q = &model.BucketQuota{
			ID:         idgen.Hex(),
			BucketID:   bucketID,
			UsedBytes:  0,
			QuotaBytes: quotaBytes,
			UpdatedAt:  time.Now(),
		}
		if err := s.store.CreateBucketQuota(q); err != nil {
			return nil, err
		}
		return q, nil
	}
	q.QuotaBytes = quotaBytes
	q.UpdatedAt = time.Now()
	if err := s.store.UpdateBucketQuota(q); err != nil {
		return nil, err
	}
	return q, nil
}
