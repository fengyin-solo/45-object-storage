package store

import (
	"testing"
	"time"

	"objectstore/internal/model"
)

func newBucket(id, name string) *model.Bucket {
	return &model.Bucket{
		ID:        id,
		Name:      name,
		Region:    "cn-north",
		Owner:     "alice",
		QuotaBytes: 1024,
		Status:    model.BucketActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestBucketCRUDAndConflict(t *testing.T) {
	s := NewMemoryStore()
	b := newBucket("b1", "logs")
	if err := s.CreateBucket(b); err != nil {
		t.Fatalf("创建桶失败: %v", err)
	}
	// 名称冲突。
	if err := s.CreateBucket(newBucket("b2", "logs")); err != ErrConflict {
		t.Fatalf("期望名称冲突，实际: %v", err)
	}
	// 按名查询。
	got, err := s.GetBucketByName("logs")
	if err != nil || got.ID != "b1" {
		t.Fatalf("按名查询失败: %v %v", got, err)
	}
	// 按 ID 查询。
	if _, err := s.GetBucket("b1"); err != nil {
		t.Fatalf("按 ID 查询失败: %v", err)
	}
	// 更新。
	b.Region = "cn-east"
	if err := s.UpdateBucket(b); err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	// 删除后不存在。
	if err := s.DeleteBucket("b1"); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
	if _, err := s.GetBucket("b1"); err != ErrNotFound {
		t.Fatalf("期望不存在: %v", err)
	}
}

func TestObjectCRUDAndKeyUniqueness(t *testing.T) {
	s := NewMemoryStore()
	o1 := &model.Object{ID: "o1", BucketID: "b1", Key: "a.txt", Size: 10, ContentType: "text/plain", ETag: "e1", Status: model.ObjectActive, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.CreateObject(o1); err != nil {
		t.Fatalf("创建对象失败: %v", err)
	}
	o2 := &model.Object{ID: "o2", BucketID: "b1", Key: "a.txt", Size: 20, ContentType: "text/plain", ETag: "e2", Status: model.ObjectActive, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.CreateObject(o2); err != ErrConflict {
		t.Fatalf("期望桶内 Key 冲突，实际: %v", err)
	}
	// 不同桶同名 Key 允许。
	o3 := &model.Object{ID: "o3", BucketID: "b2", Key: "a.txt", Size: 30, ContentType: "text/plain", ETag: "e3", Status: model.ObjectActive, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.CreateObject(o3); err != nil {
		t.Fatalf("跨桶同名应允许: %v", err)
	}
	if _, err := s.GetObjectByKey("b1", "a.txt"); err != nil {
		t.Fatalf("按 Key 查询失败: %v", err)
	}
	if _, err := s.GetObjectByKey("b1", "missing.txt"); err != ErrNotFound {
		t.Fatalf("期望不存在: %v", err)
	}
	if len(s.ListObjects()) != 2 {
		t.Fatalf("期望 2 个对象，实际 %d", len(s.ListObjects()))
	}
}

func TestObjectVersionCRUD(t *testing.T) {
	s := NewMemoryStore()
	v := &model.ObjectVersion{ID: "v1", ObjectID: "o1", BucketID: "b1", VersionNo: 1, Size: 10, ETag: "e1", CreatedAt: time.Now()}
	if err := s.CreateObjectVersion(v); err != nil {
		t.Fatalf("创建版本失败: %v", err)
	}
	if _, err := s.GetObjectVersion("v1"); err != nil {
		t.Fatalf("查询版本失败: %v", err)
	}
	if len(s.ListObjectVersions()) != 1 {
		t.Fatalf("期望 1 个版本")
	}
	if err := s.DeleteObjectVersion("v1"); err != nil {
		t.Fatalf("删除版本失败: %v", err)
	}
}

func TestLifecycleRuleCRUD(t *testing.T) {
	s := NewMemoryStore()
	r := &model.LifecycleRule{ID: "r1", BucketID: "b1", Prefix: "logs/", ExpireDays: 30, Status: model.LifecycleEnabled, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.CreateLifecycleRule(r); err != nil {
		t.Fatalf("创建规则失败: %v", err)
	}
	if _, err := s.GetLifecycleRule("r1"); err != nil {
		t.Fatalf("查询规则失败: %v", err)
	}
	r.Status = model.LifecycleDisabled
	if err := s.UpdateLifecycleRule(r); err != nil {
		t.Fatalf("更新规则失败: %v", err)
	}
	if err := s.DeleteLifecycleRule("r1"); err != nil {
		t.Fatalf("删除规则失败: %v", err)
	}
	if _, err := s.GetLifecycleRule("r1"); err != ErrNotFound {
		t.Fatalf("期望不存在: %v", err)
	}
}

func TestMultipartUploadCRUD(t *testing.T) {
	s := NewMemoryStore()
	u := &model.MultipartUpload{ID: "u1", BucketID: "b1", Key: "big.bin", Status: model.UploadInitiated, InitiatedAt: time.Now()}
	if err := s.CreateMultipartUpload(u); err != nil {
		t.Fatalf("创建上传失败: %v", err)
	}
	if _, err := s.GetMultipartUpload("u1"); err != nil {
		t.Fatalf("查询上传失败: %v", err)
	}
	u.Status = model.UploadUploading
	if err := s.UpdateMultipartUpload(u); err != nil {
		t.Fatalf("更新上传失败: %v", err)
	}
	if err := s.DeleteMultipartUpload("u1"); err != nil {
		t.Fatalf("删除上传失败: %v", err)
	}
}

func TestUploadPartUniqueness(t *testing.T) {
	s := NewMemoryStore()
	p1 := &model.UploadPart{ID: "p1", UploadID: "u1", PartNumber: 1, Size: 100, ETag: "e1", CreatedAt: time.Now()}
	if err := s.CreateUploadPart(p1); err != nil {
		t.Fatalf("创建分片失败: %v", err)
	}
	p2 := &model.UploadPart{ID: "p2", UploadID: "u1", PartNumber: 1, Size: 200, ETag: "e2", CreatedAt: time.Now()}
	if err := s.CreateUploadPart(p2); err != ErrConflict {
		t.Fatalf("期望分片编号冲突，实际: %v", err)
	}
	if err := s.DeleteUploadPart("p1"); err != nil {
		t.Fatalf("删除分片失败: %v", err)
	}
}

func TestBucketPolicyCRUD(t *testing.T) {
	s := NewMemoryStore()
	p := &model.BucketPolicy{ID: "p1", BucketID: "b1", Principal: "alice", Action: "GetObject", Resource: "*", Effect: model.PolicyEffectAllow, Status: model.PolicyEnabled, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.CreateBucketPolicy(p); err != nil {
		t.Fatalf("创建策略失败: %v", err)
	}
	if _, err := s.GetBucketPolicy("p1"); err != nil {
		t.Fatalf("查询策略失败: %v", err)
	}
	p.Effect = model.PolicyEffectDeny
	if err := s.UpdateBucketPolicy(p); err != nil {
		t.Fatalf("更新策略失败: %v", err)
	}
	if err := s.DeleteBucketPolicy("p1"); err != nil {
		t.Fatalf("删除策略失败: %v", err)
	}
}

func TestAccessLogCRUD(t *testing.T) {
	s := NewMemoryStore()
	l := &model.AccessLog{ID: "l1", BucketID: "b1", ObjectKey: "a.txt", Operation: "GetObject", IP: "1.2.3.4", Result: model.LogResultSuccess, Size: 100, CreatedAt: time.Now()}
	if err := s.CreateAccessLog(l); err != nil {
		t.Fatalf("创建日志失败: %v", err)
	}
	if _, err := s.GetAccessLog("l1"); err != nil {
		t.Fatalf("查询日志失败: %v", err)
	}
	if err := s.DeleteAccessLog("l1"); err != nil {
		t.Fatalf("删除日志失败: %v", err)
	}
}

func TestBucketQuotaCRUDAndUniqueness(t *testing.T) {
	s := NewMemoryStore()
	q := &model.BucketQuota{ID: "q1", BucketID: "b1", UsedBytes: 10, QuotaBytes: 100, UpdatedAt: time.Now()}
	if err := s.CreateBucketQuota(q); err != nil {
		t.Fatalf("创建配额失败: %v", err)
	}
	q2 := &model.BucketQuota{ID: "q2", BucketID: "b1", UsedBytes: 0, QuotaBytes: 100, UpdatedAt: time.Now()}
	if err := s.CreateBucketQuota(q2); err != ErrConflict {
		t.Fatalf("期望桶配额唯一冲突，实际: %v", err)
	}
	if got, err := s.GetBucketQuotaByBucket("b1"); err != nil || got.ID != "q1" {
		t.Fatalf("按桶查询配额失败: %v %v", got, err)
	}
	q.UsedBytes = 50
	if err := s.UpdateBucketQuota(q); err != nil {
		t.Fatalf("更新配额失败: %v", err)
	}
	if err := s.DeleteBucketQuota("q1"); err != nil {
		t.Fatalf("删除配额失败: %v", err)
	}
}

func TestNotFoundAcrossEntities(t *testing.T) {
	s := NewMemoryStore()
	if _, err := s.GetBucket("nope"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound: %v", err)
	}
	if _, err := s.GetObject("nope"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound: %v", err)
	}
	if err := s.UpdateBucket(newBucket("nope", "x")); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound: %v", err)
	}
	if err := s.DeleteObject("nope"); err != ErrNotFound {
		t.Fatalf("期望 ErrNotFound: %v", err)
	}
}
