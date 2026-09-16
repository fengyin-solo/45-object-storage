package model

import (
	"strings"
	"time"
)

// Object 状态常量。
const (
	ObjectActive  = "active"
	ObjectDeleted = "deleted"
)

// objectTransitions 定义对象状态机合法流转。
var objectTransitions = map[string]map[string]bool{
	ObjectActive: {ObjectDeleted: true},
}

// ObjectCanTransition 判断对象状态是否可流转。
func ObjectCanTransition(from, to string) bool {
	if m, ok := objectTransitions[from]; ok {
		return m[to]
	}
	return false
}

// Object 对象实体（对应存储桶中的单个对象键）。
type Object struct {
	ID          string    `json:"id"`
	BucketID    string    `json:"bucket_id"`
	Key         string    `json:"key"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	ETag        string    `json:"etag"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate 校验并规范化对象字段。
func (o *Object) Validate() error {
	o.BucketID = strings.TrimSpace(o.BucketID)
	o.Key = strings.TrimSpace(o.Key)
	o.ContentType = strings.TrimSpace(o.ContentType)
	o.ETag = strings.TrimSpace(o.ETag)
	if o.BucketID == "" {
		return NewValidationError("bucket_id", "桶 ID 不能为空")
	}
	if o.Key == "" {
		return NewValidationError("key", "对象键不能为空")
	}
	if len(o.Key) > 1024 {
		return NewValidationError("key", "对象键长度不能超过 1024")
	}
	if o.Size < 0 {
		return NewValidationError("size", "对象大小不能为负")
	}
	if o.ContentType == "" {
		o.ContentType = "application/octet-stream"
	}
	if o.Status == "" {
		o.Status = ObjectActive
	}
	if o.Status != ObjectActive && o.Status != ObjectDeleted {
		return NewValidationError("status", "对象状态不合法")
	}
	return nil
}

// ObjectFilter 对象列表筛选条件。
type ObjectFilter struct {
	BucketID string
	Status   string
	Keyword  string
	// MinSize / MaxSize 按大小字节区间筛选。
	MinSize int64
	MaxSize int64
}

// Match 判断对象是否命中筛选条件。
func (f ObjectFilter) Match(o *Object) bool {
	if f.BucketID != "" && o.BucketID != f.BucketID {
		return false
	}
	if f.Status != "" && o.Status != f.Status {
		return false
	}
	if f.MinSize > 0 && o.Size < f.MinSize {
		return false
	}
	if f.MaxSize > 0 && o.Size > f.MaxSize {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(o.Key), k) &&
			!strings.Contains(strings.ToLower(o.ContentType), k) {
			return false
		}
	}
	return true
}
