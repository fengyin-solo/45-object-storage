package model

import (
	"strings"
	"time"
)

// Bucket 状态常量。
const (
	BucketActive    = "active"
	BucketSuspended = "suspended"
)

// bucketTransitions 定义存储桶状态机合法流转。
var bucketTransitions = map[string]map[string]bool{
	BucketActive:    {BucketSuspended: true},
	BucketSuspended: {BucketActive: true},
}

// BucketCanTransition 判断存储桶状态是否可流转。
func BucketCanTransition(from, to string) bool {
	if m, ok := bucketTransitions[from]; ok {
		return m[to]
	}
	return false
}

// Bucket 存储桶实体。
type Bucket struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Region     string    `json:"region"`
	Owner      string    `json:"owner"`
	QuotaBytes int64     `json:"quota_bytes"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate 校验并规范化存储桶字段。
func (b *Bucket) Validate() error {
	b.Name = strings.TrimSpace(b.Name)
	b.Region = strings.TrimSpace(b.Region)
	b.Owner = strings.TrimSpace(b.Owner)
	if b.Name == "" {
		return NewValidationError("name", "桶名称不能为空")
	}
	if len(b.Name) < 3 || len(b.Name) > 63 {
		return NewValidationError("name", "桶名称长度须在 3 到 63 之间")
	}
	for _, r := range b.Name {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '-' && r != '.' {
			return NewValidationError("name", "桶名称只能包含小写字母、数字、短横线和点")
		}
	}
	if b.Region == "" {
		return NewValidationError("region", "区域不能为空")
	}
	if b.Owner == "" {
		return NewValidationError("owner", "所有者不能为空")
	}
	if b.QuotaBytes < 0 {
		return NewValidationError("quota_bytes", "配额容量不能为负")
	}
	if b.Status == "" {
		b.Status = BucketActive
	}
	if b.Status != BucketActive && b.Status != BucketSuspended {
		return NewValidationError("status", "桶状态不合法")
	}
	return nil
}

// BucketFilter 存储桶列表筛选条件。
type BucketFilter struct {
	Region  string
	Owner   string
	Status  string
	Keyword string
}

// Match 判断存储桶是否命中筛选条件。
func (f BucketFilter) Match(b *Bucket) bool {
	if f.Region != "" && b.Region != f.Region {
		return false
	}
	if f.Owner != "" && b.Owner != f.Owner {
		return false
	}
	if f.Status != "" && b.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(b.Name), k) &&
			!strings.Contains(strings.ToLower(b.Owner), k) {
			return false
		}
	}
	return true
}
