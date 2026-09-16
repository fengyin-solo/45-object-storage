package model

import (
	"strings"
	"time"
)

// BucketPolicy Effect 常量。
const (
	PolicyEffectAllow = "allow"
	PolicyEffectDeny  = "deny"
)

// BucketPolicy 状态常量。
const (
	PolicyEnabled  = "enabled"
	PolicyDisabled = "disabled"
)

// BucketPolicy 桶策略实体。
type BucketPolicy struct {
	ID        string    `json:"id"`
	BucketID  string    `json:"bucket_id"`
	Principal string    `json:"principal"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Effect    string    `json:"effect"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate 校验并规范化桶策略字段。
func (p *BucketPolicy) Validate() error {
	p.BucketID = strings.TrimSpace(p.BucketID)
	p.Principal = strings.TrimSpace(p.Principal)
	p.Action = strings.TrimSpace(p.Action)
	p.Resource = strings.TrimSpace(p.Resource)
	p.Effect = strings.ToLower(strings.TrimSpace(p.Effect))
	if p.BucketID == "" {
		return NewValidationError("bucket_id", "桶 ID 不能为空")
	}
	if p.Principal == "" {
		return NewValidationError("principal", "授权主体不能为空")
	}
	if p.Action == "" {
		return NewValidationError("action", "授权动作不能为空")
	}
	if p.Resource == "" {
		return NewValidationError("resource", "资源不能为空")
	}
	if p.Effect == "" {
		p.Effect = PolicyEffectAllow
	}
	if p.Effect != PolicyEffectAllow && p.Effect != PolicyEffectDeny {
		return NewValidationError("effect", "策略效果不合法")
	}
	if p.Status == "" {
		p.Status = PolicyEnabled
	}
	if p.Status != PolicyEnabled && p.Status != PolicyDisabled {
		return NewValidationError("status", "策略状态不合法")
	}
	return nil
}

// BucketPolicyFilter 桶策略列表筛选条件。
type BucketPolicyFilter struct {
	BucketID  string
	Principal string
	Effect    string
	Status    string
}

// Match 判断桶策略是否命中筛选条件。
func (f BucketPolicyFilter) Match(p *BucketPolicy) bool {
	if f.BucketID != "" && p.BucketID != f.BucketID {
		return false
	}
	if f.Principal != "" && p.Principal != f.Principal {
		return false
	}
	if f.Effect != "" && p.Effect != f.Effect {
		return false
	}
	if f.Status != "" && p.Status != f.Status {
		return false
	}
	return true
}
