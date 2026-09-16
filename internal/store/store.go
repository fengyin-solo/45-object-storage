// Package store 定义数据访问接口与内存实现。
package store

import (
	"errors"

	"objectstore/internal/model"
)

var (
	// ErrNotFound 表示记录不存在。
	ErrNotFound = errors.New("记录不存在")
	// ErrConflict 表示记录已存在或状态冲突。
	ErrConflict = errors.New("记录已存在或状态冲突")
	// ErrStateTransition 表示状态流转非法。
	ErrStateTransition = errors.New("状态流转非法")
)

// Store 聚合全部实体的数据访问方法，便于测试时替换实现。
type Store interface {
	// Bucket 存储桶。
	CreateBucket(b *model.Bucket) error
	GetBucket(id string) (*model.Bucket, error)
	GetBucketByName(name string) (*model.Bucket, error)
	ListBuckets() []*model.Bucket
	UpdateBucket(b *model.Bucket) error
	DeleteBucket(id string) error

	// Object 对象。
	CreateObject(o *model.Object) error
	GetObject(id string) (*model.Object, error)
	GetObjectByKey(bucketID, key string) (*model.Object, error)
	ListObjects() []*model.Object
	UpdateObject(o *model.Object) error
	DeleteObject(id string) error

	// ObjectVersion 对象版本。
	CreateObjectVersion(v *model.ObjectVersion) error
	GetObjectVersion(id string) (*model.ObjectVersion, error)
	ListObjectVersions() []*model.ObjectVersion
	DeleteObjectVersion(id string) error

	// LifecycleRule 生命周期规则。
	CreateLifecycleRule(r *model.LifecycleRule) error
	GetLifecycleRule(id string) (*model.LifecycleRule, error)
	ListLifecycleRules() []*model.LifecycleRule
	UpdateLifecycleRule(r *model.LifecycleRule) error
	DeleteLifecycleRule(id string) error

	// MultipartUpload 分片上传。
	CreateMultipartUpload(u *model.MultipartUpload) error
	GetMultipartUpload(id string) (*model.MultipartUpload, error)
	ListMultipartUploads() []*model.MultipartUpload
	UpdateMultipartUpload(u *model.MultipartUpload) error
	DeleteMultipartUpload(id string) error

	// UploadPart 分片。
	CreateUploadPart(p *model.UploadPart) error
	GetUploadPart(id string) (*model.UploadPart, error)
	ListUploadParts() []*model.UploadPart
	DeleteUploadPart(id string) error

	// BucketPolicy 桶策略。
	CreateBucketPolicy(p *model.BucketPolicy) error
	GetBucketPolicy(id string) (*model.BucketPolicy, error)
	ListBucketPolicies() []*model.BucketPolicy
	UpdateBucketPolicy(p *model.BucketPolicy) error
	DeleteBucketPolicy(id string) error

	// AccessLog 访问日志。
	CreateAccessLog(l *model.AccessLog) error
	GetAccessLog(id string) (*model.AccessLog, error)
	ListAccessLogs() []*model.AccessLog
	DeleteAccessLog(id string) error

	// BucketQuota 配额。
	CreateBucketQuota(q *model.BucketQuota) error
	GetBucketQuota(id string) (*model.BucketQuota, error)
	GetBucketQuotaByBucket(bucketID string) (*model.BucketQuota, error)
	ListBucketQuotas() []*model.BucketQuota
	UpdateBucketQuota(q *model.BucketQuota) error
	DeleteBucketQuota(id string) error
}
