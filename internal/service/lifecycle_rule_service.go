package service

import (
	"sort"
	"time"

	"objectstore/internal/model"
	"objectstore/internal/store"
	"objectstore/pkg/idgen"
)

// CreateLifecycleRule 创建生命周期规则。
func (s *Service) CreateLifecycleRule(input model.LifecycleRule) (*model.LifecycleRule, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetBucket(input.BucketID); err != nil {
		return nil, err
	}
	now := time.Now()
	r := &model.LifecycleRule{
		ID:             idgen.Hex(),
		BucketID:       input.BucketID,
		Prefix:         input.Prefix,
		ExpireDays:     input.ExpireDays,
		TransitionDays: input.TransitionDays,
		Status:         input.Status,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.store.CreateLifecycleRule(r); err != nil {
		return nil, err
	}
	return r, nil
}

// GetLifecycleRule 按 ID 查询生命周期规则。
func (s *Service) GetLifecycleRule(id string) (*model.LifecycleRule, error) {
	return s.store.GetLifecycleRule(id)
}

// ListLifecycleRules 分页 + 多条件筛选查询生命周期规则。
func (s *Service) ListLifecycleRules(filter model.LifecycleRuleFilter, page, size int) ([]*model.LifecycleRule, int, error) {
	all := s.store.ListLifecycleRules()
	matched := make([]*model.LifecycleRule, 0, len(all))
	for _, r := range all {
		if filter.Match(r) {
			matched = append(matched, r)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.LifecycleRule{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateLifecycleRule 更新生命周期规则可编辑字段。
func (s *Service) UpdateLifecycleRule(id string, input model.LifecycleRule) (*model.LifecycleRule, error) {
	r, err := s.store.GetLifecycleRule(id)
	if err != nil {
		return nil, err
	}
	r.Prefix = input.Prefix
	r.ExpireDays = input.ExpireDays
	r.TransitionDays = input.TransitionDays
	if err := r.Validate(); err != nil {
		return nil, err
	}
	r.UpdatedAt = time.Now()
	if err := s.store.UpdateLifecycleRule(r); err != nil {
		return nil, err
	}
	return r, nil
}

// SetLifecycleRuleStatus 转换规则状态（enabled↔disabled）。
func (s *Service) SetLifecycleRuleStatus(id, to string) (*model.LifecycleRule, error) {
	r, err := s.store.GetLifecycleRule(id)
	if err != nil {
		return nil, err
	}
	if !model.LifecycleCanTransition(r.Status, to) {
		return nil, store.ErrStateTransition
	}
	r.Status = to
	r.UpdatedAt = time.Now()
	if err := s.store.UpdateLifecycleRule(r); err != nil {
		return nil, err
	}
	return r, nil
}

// BatchUpdateLifecycleRuleStatus 批量更新规则状态。
func (s *Service) BatchUpdateLifecycleRuleStatus(ids []string, to string) (int, error) {
	updated := 0
	for _, id := range ids {
		if _, err := s.SetLifecycleRuleStatus(id, to); err == nil {
			updated++
		}
	}
	return updated, nil
}

// DeleteLifecycleRule 删除生命周期规则。
func (s *Service) DeleteLifecycleRule(id string) error {
	return s.store.DeleteLifecycleRule(id)
}
