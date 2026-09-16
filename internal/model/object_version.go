package model

import (
	"strings"
	"time"
)

// ObjectVersion 对象版本实体。
type ObjectVersion struct {
	ID            string    `json:"id"`
	ObjectID      string    `json:"object_id"`
	BucketID      string    `json:"bucket_id"`
	VersionNo     int       `json:"version_no"`
	Size          int64     `json:"size"`
	ETag          string    `json:"etag"`
	IsDeleteMarker bool     `json:"is_delete_marker"`
	CreatedAt     time.Time `json:"created_at"`
}

// Validate 校验并规范化对象版本字段。
func (v *ObjectVersion) Validate() error {
	v.ObjectID = strings.TrimSpace(v.ObjectID)
	v.BucketID = strings.TrimSpace(v.BucketID)
	v.ETag = strings.TrimSpace(v.ETag)
	if v.ObjectID == "" {
		return NewValidationError("object_id", "对象 ID 不能为空")
	}
	if v.BucketID == "" {
		return NewValidationError("bucket_id", "桶 ID 不能为空")
	}
	if v.VersionNo < 0 {
		return NewValidationError("version_no", "版本号不能为负")
	}
	if !v.IsDeleteMarker && v.Size < 0 {
		return NewValidationError("size", "版本大小不能为负")
	}
	return nil
}

// ObjectVersionFilter 对象版本列表筛选条件。
type ObjectVersionFilter struct {
	ObjectID string
	BucketID string
	// LatestOnly 是否只保留每个对象的最新版本。
	LatestOnly bool
}

// Match 判断对象版本是否命中筛选条件。
func (f ObjectVersionFilter) Match(v *ObjectVersion) bool {
	if f.ObjectID != "" && v.ObjectID != f.ObjectID {
		return false
	}
	if f.BucketID != "" && v.BucketID != f.BucketID {
		return false
	}
	return true
}
