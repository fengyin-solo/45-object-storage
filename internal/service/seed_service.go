package service

import (
	"objectstore/internal/model"
	"objectstore/pkg/idgen"
)

// SeedDemoData 初始化一批演示数据，用于前端看板与功能演示。
func (s *Service) SeedDemoData() error {
	buckets := []model.Bucket{
		{Name: "web-assets", Region: "cn-north", Owner: "frontend-team", QuotaBytes: 10 * 1024 * 1024},
		{Name: "data-lake", Region: "cn-east", Owner: "data-team", QuotaBytes: 100 * 1024 * 1024},
		{Name: "backup-archive", Region: "cn-south", Owner: "ops-team", QuotaBytes: 50 * 1024 * 1024},
	}
	for _, b := range buckets {
		if _, err := s.CreateBucket(b); err != nil {
			return err
		}
	}
	all := s.store.ListBuckets()
	if len(all) == 0 {
		return nil
	}

	// 为每个桶写入若干对象。
	type seedObject struct {
		key         string
		contentType string
		size        int64
	}
	seeds := [][]seedObject{
		{
			{"index.html", "text/html", 2048},
			{"app.js", "application/javascript", 4096},
			{"style.css", "text/css", 1024},
			{"logo.png", "image/png", 16384},
		},
		{
			{"events/2026-01.log", "text/plain", 1048576},
			{"events/2026-02.log", "text/plain", 2097152},
			{"events/2026-03.log", "text/plain", 3145728},
			{"metrics/cpu.parquet", "application/octet-stream", 8388608},
		},
		{
			{"db/dump-0801.sql", "application/sql", 5242880},
			{"db/dump-0815.sql", "application/sql", 5242880},
			{"config/nginx.conf", "text/plain", 512},
		},
	}
	for i, bucket := range all {
		if i >= len(seeds) {
			break
		}
		for j, obj := range seeds[i] {
			etag := idgen.HexN(8)
			// 部分对象通过分片上传流程写入，演示完整链路。
			if j%3 == 2 {
				if err := s.seedViaMultipart(bucket.ID, obj.key, obj.size); err != nil {
					return err
				}
				continue
			}
			if _, err := s.PutObject(bucket.ID, obj.key, obj.contentType, etag, obj.size); err != nil {
				return err
			}
		}
	}

	// 写入生命周期规则。
	_, _ = s.CreateLifecycleRule(model.LifecycleRule{
		BucketID:       all[1].ID,
		Prefix:         "events/",
		ExpireDays:     90,
		TransitionDays: 30,
		Status:         model.LifecycleEnabled,
	})
	// 写入桶策略。
	_, _ = s.CreateBucketPolicy(model.BucketPolicy{
		BucketID:  all[0].ID,
		Principal: "frontend-team",
		Action:    "GetObject",
		Resource:  "*",
		Effect:    model.PolicyEffectAllow,
	})
	// 写入访问日志。
	for _, bucket := range all[:2] {
		_, _ = s.RecordAccess(bucket.ID, "index.html", "GetObject", "10.0.0.1", model.LogResultSuccess, 2048)
		_, _ = s.RecordAccess(bucket.ID, "app.js", "GetObject", "10.0.0.2", model.LogResultSuccess, 4096)
		_, _ = s.RecordAccess(bucket.ID, "secret.txt", "GetObject", "10.0.0.3", model.LogResultDenied, 0)
	}
	return nil
}

// seedViaMultipart 通过分片上传完整流程写入对象。
func (s *Service) seedViaMultipart(bucketID, key string, size int64) error {
	u, err := s.InitMultipartUpload(bucketID, key)
	if err != nil {
		return err
	}
	const partSize int64 = 1024 * 1024
	parts := make([]CompletePart, 0)
	remaining := size
	num := 1
	for remaining > 0 {
		cur := partSize
		if remaining < cur {
			cur = remaining
		}
		etag := idgen.HexN(8)
		if _, err := s.UploadPart(u.ID, num, cur, etag); err != nil {
			return err
		}
		parts = append(parts, CompletePart{PartNumber: num, ETag: etag})
		remaining -= cur
		num++
	}
	_, err = s.CompleteMultipartUpload(u.ID, parts)
	return err
}
