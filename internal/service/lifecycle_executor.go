package service

import (
	"strings"
	"time"

	"objectstore/internal/model"
)

// LifecycleExecutionResult 生命周期规则执行结果。
type LifecycleExecutionResult struct {
	RuleID            string `json:"rule_id"`
	ExpiredCount      int    `json:"expired_count"`
	TransitionedCount int    `json:"transitioned_count"`
	ScannedCount      int    `json:"scanned_count"`
}

// ExecuteLifecycle 对指定桶执行生命周期规则：过期对象软删除，命中转换天数的对象计数。
func (s *Service) ExecuteLifecycle(bucketID string) (*LifecycleExecutionResult, error) {
	if _, err := s.store.GetBucket(bucketID); err != nil {
		return nil, err
	}
	result := &LifecycleExecutionResult{}
	for _, rule := range s.store.ListLifecycleRules() {
		if rule.BucketID != bucketID || rule.Status != model.LifecycleEnabled {
			continue
		}
		result.RuleID = rule.ID
		for _, o := range s.store.ListObjects() {
			if o.BucketID != bucketID || o.Status != model.ObjectActive {
				continue
			}
			if !prefixMatch(o.Key, rule.Prefix) {
				continue
			}
			result.ScannedCount++
			age := daysSince(o.CreatedAt)
			if rule.ExpireDays > 0 && age >= rule.ExpireDays {
				if _, err := s.DeleteObject(o.ID); err == nil {
					result.ExpiredCount++
				}
				continue
			}
			if rule.TransitionDays > 0 && age >= rule.TransitionDays {
				result.TransitionedCount++
			}
		}
	}
	return result, nil
}

// ExecuteAllLifecycles 对所有桶执行生命周期规则。
func (s *Service) ExecuteAllLifecycles() (map[string]*LifecycleExecutionResult, error) {
	out := make(map[string]*LifecycleExecutionResult)
	for _, b := range s.store.ListBuckets() {
		res, err := s.ExecuteLifecycle(b.ID)
		if err != nil {
			continue
		}
		out[b.ID] = res
	}
	return out, nil
}

// ExpireCandidates 返回指定桶中即将过期（或已过期）的对象，供预览。
func (s *Service) ExpireCandidates(bucketID string, expireDays int) []*model.Object {
	var out []*model.Object
	for _, o := range s.store.ListObjects() {
		if o.BucketID != bucketID || o.Status != model.ObjectActive {
			continue
		}
		if expireDays > 0 && daysSince(o.CreatedAt) >= expireDays {
			out = append(out, o)
		}
	}
	return out
}

// prefixMatch 判断对象键是否命中规则前缀（空前缀匹配所有）。
func prefixMatch(key, prefix string) bool {
	if prefix == "" {
		return true
	}
	return strings.HasPrefix(key, prefix)
}

// daysSince 计算自指定时间起经过的整天数。
func daysSince(t time.Time) int {
	if t.IsZero() {
		return 0
	}
	d := time.Since(t)
	if d < 0 {
		return 0
	}
	return int(d.Hours() / 24)
}
