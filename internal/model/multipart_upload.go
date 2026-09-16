package model

import (
	"strings"
	"time"
)

// MultipartUpload 状态常量。
const (
	UploadInitiated  = "initiated"
	UploadUploading  = "uploading"
	UploadCompleted  = "completed"
	UploadAborted    = "aborted"
)

// uploadTransitions 定义分片上传状态机合法流转。
var uploadTransitions = map[string]map[string]bool{
	UploadInitiated: {UploadUploading: true, UploadCompleted: true, UploadAborted: true},
	UploadUploading: {UploadCompleted: true, UploadAborted: true},
}

// MultipartCanTransition 判断分片上传状态是否可流转。
func MultipartCanTransition(from, to string) bool {
	if m, ok := uploadTransitions[from]; ok {
		return m[to]
	}
	return false
}

// MultipartUpload 分片上传实体。
type MultipartUpload struct {
	ID          string     `json:"id"`
	BucketID    string     `json:"bucket_id"`
	Key         string     `json:"key"`
	Status      string     `json:"status"`
	PartCount   int        `json:"part_count"`
	Size        int64      `json:"size"`
	InitiatedAt time.Time  `json:"initiated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// Validate 校验并规范化分片上传字段。
func (u *MultipartUpload) Validate() error {
	u.BucketID = strings.TrimSpace(u.BucketID)
	u.Key = strings.TrimSpace(u.Key)
	if u.BucketID == "" {
		return NewValidationError("bucket_id", "桶 ID 不能为空")
	}
	if u.Key == "" {
		return NewValidationError("key", "对象键不能为空")
	}
	if u.PartCount < 0 {
		return NewValidationError("part_count", "分片数量不能为负")
	}
	if u.Size < 0 {
		return NewValidationError("size", "上传大小不能为负")
	}
	if u.Status == "" {
		u.Status = UploadInitiated
	}
	if u.Status != UploadInitiated && u.Status != UploadUploading &&
		u.Status != UploadCompleted && u.Status != UploadAborted {
		return NewValidationError("status", "上传状态不合法")
	}
	return nil
}

// MultipartUploadFilter 分片上传列表筛选条件。
type MultipartUploadFilter struct {
	BucketID string
	Status   string
	Keyword  string
}

// Match 判断分片上传是否命中筛选条件。
func (f MultipartUploadFilter) Match(u *MultipartUpload) bool {
	if f.BucketID != "" && u.BucketID != f.BucketID {
		return false
	}
	if f.Status != "" && u.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(u.Key), k) {
			return false
		}
	}
	return true
}
