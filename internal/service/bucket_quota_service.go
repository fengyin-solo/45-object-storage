package service

import (
	"sort"
	"time"

	"objectstore/internal/model"
	"objectstore/pkg/idgen"
)

// CreateBucketQuota 创建配额记录。
func (s *Service) CreateBucketQuota(input model.BucketQuota) (*model.BucketQuota, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetBucket(input.BucketID); err != nil {
		return nil, err
	}
	q := &model.BucketQuota{
		ID:         idgen.Hex(),
		BucketID:   input.BucketID,
		UsedBytes:  input.UsedBytes,
		QuotaBytes: input.QuotaBytes,
		UpdatedAt:  time.Now(),
	}
	if err := s.store.CreateBucketQuota(q); err != nil {
		return nil, err
	}
	return q, nil
}

// GetBucketQuota 按 ID 查询配额记录。
func (s *Service) GetBucketQuota(id string) (*model.BucketQuota, error) {
	return s.store.GetBucketQuota(id)
}

// GetBucketQuotaByBucket 按桶 ID 查询配额记录。
func (s *Service) GetBucketQuotaByBucket(bucketID string) (*model.BucketQuota, error) {
	return s.store.GetBucketQuotaByBucket(bucketID)
}

// ListBucketQuotas 分页 + 多条件筛选查询配额记录。
func (s *Service) ListBucketQuotas(filter model.BucketQuotaFilter, page, size int) ([]*model.BucketQuota, int, error) {
	all := s.store.ListBucketQuotas()
	matched := make([]*model.BucketQuota, 0, len(all))
	for _, q := range all {
		if filter.Match(q) {
			matched = append(matched, q)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].UsedBytes > matched[j].UsedBytes
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.BucketQuota{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateBucketQuota 更新配额记录。
func (s *Service) UpdateBucketQuota(id string, input model.BucketQuota) (*model.BucketQuota, error) {
	q, err := s.store.GetBucketQuota(id)
	if err != nil {
		return nil, err
	}
	q.QuotaBytes = input.QuotaBytes
	if input.UsedBytes >= 0 {
		q.UsedBytes = input.UsedBytes
	}
	if err := q.Validate(); err != nil {
		return nil, err
	}
	q.UpdatedAt = time.Now()
	if err := s.store.UpdateBucketQuota(q); err != nil {
		return nil, err
	}
	return q, nil
}

// DeleteBucketQuota 删除配额记录。
func (s *Service) DeleteBucketQuota(id string) error {
	return s.store.DeleteBucketQuota(id)
}
