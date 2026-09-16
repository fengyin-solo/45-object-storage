package service

import (
	"time"

	"objectstore/internal/model"
)

// RecalculateQuotas 根据对象实际大小重算所有桶的已用容量，返回处理的桶数。
func (s *Service) RecalculateQuotas() (int, error) {
	usedByBucket := make(map[string]int64)
	for _, o := range s.store.ListObjects() {
		if o.Status == model.ObjectActive {
			usedByBucket[o.BucketID] += o.Size
		}
	}
	count := 0
	for _, q := range s.store.ListBucketQuotas() {
		used := usedByBucket[q.BucketID]
		if q.UsedBytes == used {
			continue
		}
		q.UsedBytes = used
		q.UpdatedAt = time.Now()
		if err := s.store.UpdateBucketQuota(q); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// RecalculateBucketQuota 重算单个桶的已用容量。
func (s *Service) RecalculateBucketQuota(bucketID string) (*model.BucketQuota, error) {
	if _, err := s.store.GetBucket(bucketID); err != nil {
		return nil, err
	}
	var used int64
	for _, o := range s.store.ListObjects() {
		if o.BucketID == bucketID && o.Status == model.ObjectActive {
			used += o.Size
		}
	}
	q, err := s.store.GetBucketQuotaByBucket(bucketID)
	if err != nil {
		return nil, err
	}
	q.UsedBytes = used
	q.UpdatedAt = time.Now()
	if err := s.store.UpdateBucketQuota(q); err != nil {
		return nil, err
	}
	return q, nil
}

// QuotaExceededBuckets 返回所有已超配额的桶配额记录。
func (s *Service) QuotaExceededBuckets() []*model.BucketQuota {
	var out []*model.BucketQuota
	for _, q := range s.store.ListBucketQuotas() {
		if q.Exceeded() {
			out = append(out, q)
		}
	}
	return out
}
