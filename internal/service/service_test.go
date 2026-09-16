package service

import (
	"testing"

	"objectstore/internal/config"
	"objectstore/internal/model"
	"objectstore/internal/store"
	"objectstore/pkg/logger"
)

func newTestService() *Service {
	cfg := &config.Config{MaxPageSize: 100}
	log := logger.NewLevel(logger.LevelError)
	return New(store.NewMemoryStore(), log, cfg)
}

func newTestBucket(s *Service, name, region string) *model.Bucket {
	b, err := s.CreateBucket(model.Bucket{Name: name, Region: region, Owner: "tester", QuotaBytes: 1024 * 1024})
	if err != nil {
		panic(err)
	}
	return b
}

func TestMultipartUploadFullFlow(t *testing.T) {
	s := newTestService()
	b := newTestBucket(s, "mpu-bucket", "cn-north")

	u, err := s.InitMultipartUpload(b.ID, "large.bin")
	if err != nil {
		t.Fatalf("初始化上传失败: %v", err)
	}
	if u.Status != model.UploadInitiated {
		t.Fatalf("期望 initiated，实际 %s", u.Status)
	}

	// 上传 3 个分片。
	if _, err := s.UploadPart(u.ID, 1, 100, "e1"); err != nil {
		t.Fatalf("上传分片 1 失败: %v", err)
	}
	if _, err := s.UploadPart(u.ID, 2, 200, "e2"); err != nil {
		t.Fatalf("上传分片 2 失败: %v", err)
	}
	if _, err := s.UploadPart(u.ID, 3, 300, "e3"); err != nil {
		t.Fatalf("上传分片 3 失败: %v", err)
	}

	got, _ := s.GetMultipartUpload(u.ID)
	if got.Status != model.UploadUploading || got.PartCount != 3 || got.Size != 600 {
		t.Fatalf("上传状态/计数错误: %+v", got)
	}

	// 缺少分片时完成应失败。
	if _, err := s.CompleteMultipartUpload(u.ID, []CompletePart{{PartNumber: 1, ETag: "e1"}}); err == nil {
		t.Fatalf("分片不齐全应失败")
	}

	// 完整完成。
	completed, err := s.CompleteMultipartUpload(u.ID, []CompletePart{
		{PartNumber: 1, ETag: "e1"},
		{PartNumber: 2, ETag: "e2"},
		{PartNumber: 3, ETag: "e3"},
	})
	if err != nil {
		t.Fatalf("完成上传失败: %v", err)
	}
	if completed.Status != model.UploadCompleted || completed.Size != 600 {
		t.Fatalf("完成状态/大小错误: %+v", completed)
	}
	// 完成后应生成对象。
	if _, err := s.store.GetObjectByKey(b.ID, "large.bin"); err != nil {
		t.Fatalf("完成后未生成对象: %v", err)
	}
}

func TestAbortMultipartUpload(t *testing.T) {
	s := newTestService()
	b := newTestBucket(s, "abort-bucket", "cn-north")
	u, _ := s.InitMultipartUpload(b.ID, "x.bin")
	_, _ = s.UploadPart(u.ID, 1, 10, "e1")

	aborted, err := s.AbortMultipartUpload(u.ID)
	if err != nil {
		t.Fatalf("中止上传失败: %v", err)
	}
	if aborted.Status != model.UploadAborted {
		t.Fatalf("期望 aborted，实际 %s", aborted.Status)
	}
	// 已中止不能继续上传。
	if _, err := s.UploadPart(u.ID, 2, 10, "e2"); err == nil {
		t.Fatalf("已中止不应能继续上传")
	}
}

func TestObjectVersioning(t *testing.T) {
	s := newTestService()
	b := newTestBucket(s, "ver-bucket", "cn-north")

	o1, err := s.PutObject(b.ID, "doc.txt", "text/plain", "etag-1", 100)
	if err != nil {
		t.Fatalf("首次写入失败: %v", err)
	}
	// 覆盖写入生成新版本。
	o2, err := s.PutObject(b.ID, "doc.txt", "text/plain", "etag-2", 200)
	if err != nil {
		t.Fatalf("覆盖写入失败: %v", err)
	}
	if o1.ID != o2.ID {
		t.Fatalf("覆盖写入应复用同一对象")
	}

	versions, _, err := s.ListObjectVersions(model.ObjectVersionFilter{ObjectID: o1.ID}, 1, 100)
	if err != nil || len(versions) != 2 {
		t.Fatalf("期望 2 个版本，实际 %d (err=%v)", len(versions), err)
	}

	// 删除生成 delete marker。
	deleted, err := s.DeleteObject(o1.ID)
	if err != nil {
		t.Fatalf("删除失败: %v", err)
	}
	if deleted.Status != model.ObjectDeleted {
		t.Fatalf("期望 deleted 状态")
	}
	versions, _, _ = s.ListObjectVersions(model.ObjectVersionFilter{ObjectID: o1.ID}, 1, 100)
	if len(versions) != 3 {
		t.Fatalf("期望 3 个版本（含 delete marker），实际 %d", len(versions))
	}
	hasMarker := false
	for _, v := range versions {
		if v.IsDeleteMarker {
			hasMarker = true
		}
	}
	if !hasMarker {
		t.Fatalf("缺少 delete marker")
	}
}

func TestBucketStateMachine(t *testing.T) {
	s := newTestService()
	b := newTestBucket(s, "sm-bucket", "cn-north")

	suspended, err := s.SuspendBucket(b.ID)
	if err != nil {
		t.Fatalf("挂起失败: %v", err)
	}
	if suspended.Status != model.BucketSuspended {
		t.Fatalf("期望 suspended，实际 %s", suspended.Status)
	}
	activated, err := s.ActivateBucket(b.ID)
	if err != nil {
		t.Fatalf("恢复失败: %v", err)
	}
	if activated.Status != model.BucketActive {
		t.Fatalf("期望 active，实际 %s", activated.Status)
	}
}

func TestLifecycleRuleStateMachine(t *testing.T) {
	s := newTestService()
	b := newTestBucket(s, "lc-bucket", "cn-north")
	r, err := s.CreateLifecycleRule(model.LifecycleRule{
		BucketID:   b.ID,
		Prefix:     "logs/",
		ExpireDays: 30,
		Status:     model.LifecycleEnabled,
	})
	if err != nil {
		t.Fatalf("创建规则失败: %v", err)
	}
	disabled, err := s.SetLifecycleRuleStatus(r.ID, model.LifecycleDisabled)
	if err != nil || disabled.Status != model.LifecycleDisabled {
		t.Fatalf("禁用失败: %v %+v", err, disabled)
	}
	// 非法流转（disabled→disabled 同状态不算非法，但 enabled 状态机不含该目标）。
	// 生命周期状态机允许 disabled→enabled。
	enabled, err := s.SetLifecycleRuleStatus(r.ID, model.LifecycleEnabled)
	if err != nil || enabled.Status != model.LifecycleEnabled {
		t.Fatalf("启用失败: %v %+v", err, enabled)
	}
}

func TestBucketDeleteWithObjectsFails(t *testing.T) {
	s := newTestService()
	b := newTestBucket(s, "del-bucket", "cn-north")
	_, _ = s.PutObject(b.ID, "a.txt", "text/plain", "e1", 10)
	if err := s.DeleteBucket(b.ID); err == nil {
		t.Fatalf("桶内有对象时删除应失败")
	}
}

func TestStats(t *testing.T) {
	s := newTestService()
	b1 := newTestBucket(s, "stats-b1", "cn-north")
	b2 := newTestBucket(s, "stats-b2", "cn-east")
	_, _ = s.PutObject(b1.ID, "a.txt", "text/plain", "e1", 100)
	_, _ = s.PutObject(b1.ID, "b.txt", "text/plain", "e2", 200)
	_, _ = s.PutObject(b2.ID, "c.txt", "text/plain", "e3", 300)
	_, _ = s.RecordAccess(b1.ID, "a.txt", "GetObject", "1.1.1.1", model.LogResultSuccess, 100)
	_, _ = s.RecordAccess(b1.ID, "b.txt", "GetObject", "1.1.1.1", model.LogResultSuccess, 200)

	ov := s.Overview()
	if ov.BucketCount != 2 || ov.ObjectCount != 3 || ov.TotalBytes != 600 {
		t.Fatalf("概览统计错误: %+v", ov)
	}
	byRegion := s.StatsByRegion()
	if len(byRegion) != 2 {
		t.Fatalf("期望 2 个区域，实际 %d", len(byRegion))
	}
	top := s.TopAccessBuckets(1)
	if len(top) != 1 || top[0].BucketID != b1.ID || top[0].AccessCount != 2 {
		t.Fatalf("访问排行错误: %+v", top)
	}
	snap := s.ExportSnapshot()
	if snap.Overview == nil || len(snap.Buckets) != 2 {
		t.Fatalf("导出快照错误: %+v", snap)
	}
}

func TestBucketPolicyEvaluate(t *testing.T) {
	s := newTestService()
	b := newTestBucket(s, "policy-bucket", "cn-north")
	_, err := s.CreateBucketPolicy(model.BucketPolicy{
		BucketID:  b.ID,
		Principal: "alice",
		Action:    "GetObject",
		Resource:  "*",
		Effect:    model.PolicyEffectAllow,
	})
	if err != nil {
		t.Fatalf("创建策略失败: %v", err)
	}
	allowed, _ := s.EvaluateBucketPolicy(b.ID, "alice", "GetObject")
	if !allowed {
		t.Fatalf("alice 应被允许")
	}
	allowed, _ = s.EvaluateBucketPolicy(b.ID, "bob", "GetObject")
	if allowed {
		t.Fatalf("bob 不应被允许")
	}
	// deny 优先。
	_, _ = s.CreateBucketPolicy(model.BucketPolicy{
		BucketID:  b.ID,
		Principal: "*",
		Action:    "GetObject",
		Resource:  "*",
		Effect:    model.PolicyEffectDeny,
	})
	allowed, _ = s.EvaluateBucketPolicy(b.ID, "alice", "GetObject")
	if allowed {
		t.Fatalf("deny 应优先")
	}
}

func TestListPagination(t *testing.T) {
	s := newTestService()
	b := newTestBucket(s, "page-bucket", "cn-north")
	for i := 0; i < 5; i++ {
		key := "k" + string(rune('a'+i)) + ".txt"
		_, _ = s.PutObject(b.ID, key, "text/plain", "e", int64((i+1)*10))
	}
	items, total, err := s.ListObjects(model.ObjectFilter{BucketID: b.ID}, 1, 2)
	if err != nil || total != 5 || len(items) != 2 {
		t.Fatalf("分页错误: total=%d len=%d err=%v", total, len(items), err)
	}
	items, total, _ = s.ListObjects(model.ObjectFilter{BucketID: b.ID}, 3, 2)
	if total != 5 || len(items) != 1 {
		t.Fatalf("末页错误: total=%d len=%d", total, len(items))
	}
}
