package store

import (
	"sync"

	"objectstore/internal/model"
)

// MemoryStore 基于内存 map 的存储实现，所有读写均加锁保证并发安全。
type MemoryStore struct {
	mu               sync.RWMutex
	buckets          map[string]*model.Bucket
	objects          map[string]*model.Object
	objectVersions   map[string]*model.ObjectVersion
	lifecycleRules   map[string]*model.LifecycleRule
	multipartUploads map[string]*model.MultipartUpload
	uploadParts      map[string]*model.UploadPart
	bucketPolicies   map[string]*model.BucketPolicy
	accessLogs       map[string]*model.AccessLog
	bucketQuotas     map[string]*model.BucketQuota
}

// NewMemoryStore 创建空的内存存储。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		buckets:          make(map[string]*model.Bucket),
		objects:          make(map[string]*model.Object),
		objectVersions:   make(map[string]*model.ObjectVersion),
		lifecycleRules:   make(map[string]*model.LifecycleRule),
		multipartUploads: make(map[string]*model.MultipartUpload),
		uploadParts:      make(map[string]*model.UploadPart),
		bucketPolicies:   make(map[string]*model.BucketPolicy),
		accessLogs:       make(map[string]*model.AccessLog),
		bucketQuotas:     make(map[string]*model.BucketQuota),
	}
}

var _ Store = (*MemoryStore)(nil)
