package model

import (
	"strings"
	"time"
)

// UploadPart 分片实体。
type UploadPart struct {
	ID         string    `json:"id"`
	UploadID   string    `json:"upload_id"`
	PartNumber int       `json:"part_number"`
	Size       int64     `json:"size"`
	ETag       string    `json:"etag"`
	CreatedAt  time.Time `json:"created_at"`
}

// Validate 校验并规范化分片字段。
func (p *UploadPart) Validate() error {
	p.UploadID = strings.TrimSpace(p.UploadID)
	p.ETag = strings.TrimSpace(p.ETag)
	if p.UploadID == "" {
		return NewValidationError("upload_id", "上传 ID 不能为空")
	}
	if p.PartNumber < 1 || p.PartNumber > 10000 {
		return NewValidationError("part_number", "分片编号须在 1 到 10000 之间")
	}
	if p.Size < 0 {
		return NewValidationError("size", "分片大小不能为负")
	}
	if p.ETag == "" {
		return NewValidationError("etag", "分片 ETag 不能为空")
	}
	return nil
}

// UploadPartFilter 分片列表筛选条件。
type UploadPartFilter struct {
	UploadID string
	// MinPartNumber / MaxPartNumber 按分片编号区间筛选。
	MinPartNumber int
	MaxPartNumber int
}

// Match 判断分片是否命中筛选条件。
func (f UploadPartFilter) Match(p *UploadPart) bool {
	if f.UploadID != "" && p.UploadID != f.UploadID {
		return false
	}
	if f.MinPartNumber > 0 && p.PartNumber < f.MinPartNumber {
		return false
	}
	if f.MaxPartNumber > 0 && p.PartNumber > f.MaxPartNumber {
		return false
	}
	return true
}
