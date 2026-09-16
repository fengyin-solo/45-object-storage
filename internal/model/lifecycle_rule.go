package model

import (
	"strings"
	"time"
)

// LifecycleRule 状态常量。
const (
	LifecycleEnabled  = "enabled"
	LifecycleDisabled = "disabled"
)

// lifecycleTransitions 定义生命周期规则状态机合法流转。
var lifecycleTransitions = map[string]map[string]bool{
	LifecycleEnabled:  {LifecycleDisabled: true},
	LifecycleDisabled: {LifecycleEnabled: true},
}

// LifecycleCanTransition 判断生命周期规则状态是否可流转。
func LifecycleCanTransition(from, to string) bool {
	if m, ok := lifecycleTransitions[from]; ok {
		return m[to]
	}
	return false
}

// LifecycleRule 生命周期规则实体。
type LifecycleRule struct {
	ID             string    `json:"id"`
	BucketID       string    `json:"bucket_id"`
	Prefix         string    `json:"prefix"`
	ExpireDays     int       `json:"expire_days"`
	TransitionDays int       `json:"transition_days"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Validate 校验并规范化生命周期规则字段。
func (r *LifecycleRule) Validate() error {
	r.BucketID = strings.TrimSpace(r.BucketID)
	r.Prefix = strings.TrimSpace(r.Prefix)
	if r.BucketID == "" {
		return NewValidationError("bucket_id", "桶 ID 不能为空")
	}
	if r.ExpireDays < 0 {
		return NewValidationError("expire_days", "过期天数不能为负")
	}
	if r.TransitionDays < 0 {
		return NewValidationError("transition_days", "转换天数不能为负")
	}
	if r.ExpireDays == 0 && r.TransitionDays == 0 {
		return NewValidationError("expire_days", "过期天数与转换天数不能同时为空")
	}
	if r.TransitionDays > 0 && r.ExpireDays > 0 && r.TransitionDays >= r.ExpireDays {
		return NewValidationError("transition_days", "转换天数必须小于过期天数")
	}
	if r.Status == "" {
		r.Status = LifecycleEnabled
	}
	if r.Status != LifecycleEnabled && r.Status != LifecycleDisabled {
		return NewValidationError("status", "规则状态不合法")
	}
	return nil
}

// LifecycleRuleFilter 生命周期规则列表筛选条件。
type LifecycleRuleFilter struct {
	BucketID string
	Status   string
	Prefix   string
}

// Match 判断生命周期规则是否命中筛选条件。
func (f LifecycleRuleFilter) Match(r *LifecycleRule) bool {
	if f.BucketID != "" && r.BucketID != f.BucketID {
		return false
	}
	if f.Status != "" && r.Status != f.Status {
		return false
	}
	if f.Prefix != "" && !strings.HasPrefix(r.Prefix, f.Prefix) {
		return false
	}
	return true
}
