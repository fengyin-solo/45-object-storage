package model

import (
	"strings"
	"time"
)

// BucketQuota 配额实体（跟踪桶的已用容量）。
type BucketQuota struct {
	ID         string    `json:"id"`
	BucketID   string    `json:"bucket_id"`
	UsedBytes  int64     `json:"used_bytes"`
	QuotaBytes int64     `json:"quota_bytes"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate 校验并规范化配额字段。
func (q *BucketQuota) Validate() error {
	q.BucketID = strings.TrimSpace(q.BucketID)
	if q.BucketID == "" {
		return NewValidationError("bucket_id", "桶 ID 不能为空")
	}
	if q.UsedBytes < 0 {
		return NewValidationError("used_bytes", "已用容量不能为负")
	}
	if q.QuotaBytes < 0 {
		return NewValidationError("quota_bytes", "配额容量不能为负")
	}
	return nil
}

// Exceeded 判断是否超出配额。
func (q *BucketQuota) Exceeded() bool {
	if q.QuotaBytes <= 0 {
		return false
	}
	return q.UsedBytes > q.QuotaBytes
}

// Remaining 返回剩余可用容量（配额为 0 表示不限制，返回 -1）。
func (q *BucketQuota) Remaining() int64 {
	if q.QuotaBytes <= 0 {
		return -1
	}
	rem := q.QuotaBytes - q.UsedBytes
	if rem < 0 {
		return 0
	}
	return rem
}

// BucketQuotaFilter 配额列表筛选条件。
type BucketQuotaFilter struct {
	BucketID string
	// ExceededOnly 是否只筛选已超配额的桶。
	ExceededOnly bool
}

// Match 判断配额是否命中筛选条件。
func (f BucketQuotaFilter) Match(q *BucketQuota) bool {
	if f.BucketID != "" && q.BucketID != f.BucketID {
		return false
	}
	if f.ExceededOnly && !q.Exceeded() {
		return false
	}
	return true
}
