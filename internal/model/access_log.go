package model

import (
	"strings"
	"time"
)

// AccessLog Result 常量。
const (
	LogResultSuccess = "success"
	LogResultDenied  = "denied"
)

// AccessLog 访问日志实体。
type AccessLog struct {
	ID         string    `json:"id"`
	BucketID   string    `json:"bucket_id"`
	ObjectKey  string    `json:"object_key"`
	Operation  string    `json:"operation"`
	IP         string    `json:"ip"`
	Result     string    `json:"result"`
	Size       int64     `json:"size"`
	CreatedAt  time.Time `json:"created_at"`
}

// Validate 校验并规范化访问日志字段。
func (l *AccessLog) Validate() error {
	l.BucketID = strings.TrimSpace(l.BucketID)
	l.ObjectKey = strings.TrimSpace(l.ObjectKey)
	l.Operation = strings.TrimSpace(l.Operation)
	l.IP = strings.TrimSpace(l.IP)
	l.Result = strings.ToLower(strings.TrimSpace(l.Result))
	if l.BucketID == "" {
		return NewValidationError("bucket_id", "桶 ID 不能为空")
	}
	if l.Operation == "" {
		return NewValidationError("operation", "操作类型不能为空")
	}
	if l.IP == "" {
		return NewValidationError("ip", "来源 IP 不能为空")
	}
	if l.Result == "" {
		l.Result = LogResultSuccess
	}
	if l.Result != LogResultSuccess && l.Result != LogResultDenied {
		return NewValidationError("result", "访问结果不合法")
	}
	if l.Size < 0 {
		return NewValidationError("size", "传输大小不能为负")
	}
	return nil
}

// AccessLogFilter 访问日志列表筛选条件。
type AccessLogFilter struct {
	BucketID  string
	ObjectKey string
	Operation string
	IP        string
	Result    string
}

// Match 判断访问日志是否命中筛选条件。
func (f AccessLogFilter) Match(l *AccessLog) bool {
	if f.BucketID != "" && l.BucketID != f.BucketID {
		return false
	}
	if f.ObjectKey != "" && !strings.Contains(l.ObjectKey, f.ObjectKey) {
		return false
	}
	if f.Operation != "" && l.Operation != f.Operation {
		return false
	}
	if f.IP != "" && l.IP != f.IP {
		return false
	}
	if f.Result != "" && l.Result != f.Result {
		return false
	}
	return true
}
