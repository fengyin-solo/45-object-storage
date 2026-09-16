package service

import (
	"sort"
	"time"

	"objectstore/internal/model"
	"objectstore/pkg/idgen"
)

// CreateBucketPolicy 创建桶策略。
func (s *Service) CreateBucketPolicy(input model.BucketPolicy) (*model.BucketPolicy, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetBucket(input.BucketID); err != nil {
		return nil, err
	}
	now := time.Now()
	p := &model.BucketPolicy{
		ID:        idgen.Hex(),
		BucketID:  input.BucketID,
		Principal: input.Principal,
		Action:    input.Action,
		Resource:  input.Resource,
		Effect:    input.Effect,
		Status:    input.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.CreateBucketPolicy(p); err != nil {
		return nil, err
	}
	return p, nil
}

// GetBucketPolicy 按 ID 查询桶策略。
func (s *Service) GetBucketPolicy(id string) (*model.BucketPolicy, error) {
	return s.store.GetBucketPolicy(id)
}

// ListBucketPolicies 分页 + 多条件筛选查询桶策略。
func (s *Service) ListBucketPolicies(filter model.BucketPolicyFilter, page, size int) ([]*model.BucketPolicy, int, error) {
	all := s.store.ListBucketPolicies()
	matched := make([]*model.BucketPolicy, 0, len(all))
	for _, p := range all {
		if filter.Match(p) {
			matched = append(matched, p)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.BucketPolicy{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateBucketPolicy 更新桶策略可编辑字段。
func (s *Service) UpdateBucketPolicy(id string, input model.BucketPolicy) (*model.BucketPolicy, error) {
	p, err := s.store.GetBucketPolicy(id)
	if err != nil {
		return nil, err
	}
	p.Principal = input.Principal
	p.Action = input.Action
	p.Resource = input.Resource
	p.Effect = input.Effect
	if err := p.Validate(); err != nil {
		return nil, err
	}
	p.UpdatedAt = time.Now()
	if err := s.store.UpdateBucketPolicy(p); err != nil {
		return nil, err
	}
	return p, nil
}

// SetBucketPolicyStatus 切换桶策略启用状态。
func (s *Service) SetBucketPolicyStatus(id, to string) (*model.BucketPolicy, error) {
	p, err := s.store.GetBucketPolicy(id)
	if err != nil {
		return nil, err
	}
	if to != model.PolicyEnabled && to != model.PolicyDisabled {
		return nil, model.NewValidationError("status", "策略状态不合法")
	}
	p.Status = to
	p.UpdatedAt = time.Now()
	if err := s.store.UpdateBucketPolicy(p); err != nil {
		return nil, err
	}
	return p, nil
}

// DeleteBucketPolicy 删除桶策略。
func (s *Service) DeleteBucketPolicy(id string) error {
	return s.store.DeleteBucketPolicy(id)
}

// EvaluateBucketPolicy 评估某主体对桶执行某动作是否被允许。
// 规则：deny 优先于 allow；无匹配策略时默认拒绝。
func (s *Service) EvaluateBucketPolicy(bucketID, principal, action string) (bool, string) {
	allowed := false
	for _, p := range s.store.ListBucketPolicies() {
		if p.BucketID != bucketID || p.Status != model.PolicyEnabled {
			continue
		}
		if p.Principal != "*" && p.Principal != principal {
			continue
		}
		if p.Action != "*" && p.Action != action {
			continue
		}
		if p.Effect == model.PolicyEffectDeny {
			return false, "denied by policy " + p.ID
		}
		if p.Effect == model.PolicyEffectAllow {
			allowed = true
		}
	}
	if !allowed {
		return false, "no matching allow policy"
	}
	return true, ""
}
