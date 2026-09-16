package service

import (
	"sort"
	"time"

	"objectstore/internal/model"
	"objectstore/pkg/idgen"
)

// RecordAccess 记录一条访问日志（跨实体校验桶存在）。
func (s *Service) RecordAccess(bucketID, objectKey, operation, ip, result string, size int64) (*model.AccessLog, error) {
	if _, err := s.store.GetBucket(bucketID); err != nil {
		return nil, err
	}
	now := time.Now()
	l := &model.AccessLog{
		ID:        idgen.Hex(),
		BucketID:  bucketID,
		ObjectKey: objectKey,
		Operation: operation,
		IP:        ip,
		Result:    result,
		Size:      size,
		CreatedAt: now,
	}
	if err := l.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.CreateAccessLog(l); err != nil {
		return nil, err
	}
	return l, nil
}

// GetAccessLog 按 ID 查询访问日志。
func (s *Service) GetAccessLog(id string) (*model.AccessLog, error) {
	return s.store.GetAccessLog(id)
}

// ListAccessLogs 分页 + 多条件筛选查询访问日志。
func (s *Service) ListAccessLogs(filter model.AccessLogFilter, page, size int) ([]*model.AccessLog, int, error) {
	all := s.store.ListAccessLogs()
	matched := make([]*model.AccessLog, 0, len(all))
	for _, l := range all {
		if filter.Match(l) {
			matched = append(matched, l)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.AccessLog{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// DeleteAccessLog 删除访问日志。
func (s *Service) DeleteAccessLog(id string) error {
	return s.store.DeleteAccessLog(id)
}
