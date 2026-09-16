package service

import (
	"sort"
	"time"

	"objectstore/internal/model"
)

// Overview 全局概览统计。
type Overview struct {
	BucketCount        int   `json:"bucket_count"`
	ObjectCount        int   `json:"object_count"`
	ActiveObjectCount  int   `json:"active_object_count"`
	TotalBytes         int64 `json:"total_bytes"`
	MultipartUploadCount int `json:"multipart_upload_count"`
	LifecycleRuleCount int   `json:"lifecycle_rule_count"`
	PolicyCount        int   `json:"policy_count"`
	AccessLogCount     int   `json:"access_log_count"`
	GeneratedAt        time.Time `json:"generated_at"`
}

// BucketStat 单桶统计结果。
type BucketStat struct {
	BucketID    string `json:"bucket_id"`
	BucketName  string `json:"bucket_name"`
	Region      string `json:"region"`
	ObjectCount int    `json:"object_count"`
	TotalBytes  int64  `json:"total_bytes"`
	UsedBytes   int64  `json:"used_bytes"`
	QuotaBytes  int64  `json:"quota_bytes"`
}

// RegionStat 区域统计结果。
type RegionStat struct {
	Region      string `json:"region"`
	BucketCount int    `json:"bucket_count"`
	ObjectCount int    `json:"object_count"`
	TotalBytes  int64  `json:"total_bytes"`
}

// AccessRank 访问量排行项。
type AccessRank struct {
	BucketID    string `json:"bucket_id"`
	BucketName  string `json:"bucket_name"`
	AccessCount int    `json:"access_count"`
	TotalBytes  int64  `json:"total_bytes"`
}

// Overview 计算全局概览统计。
func (s *Service) Overview() *Overview {
	buckets := s.store.ListBuckets()
	objects := s.store.ListObjects()
	active := 0
	var totalBytes int64
	for _, o := range objects {
		if o.Status == model.ObjectActive {
			active++
			totalBytes += o.Size
		}
	}
	return &Overview{
		BucketCount:          len(buckets),
		ObjectCount:          len(objects),
		ActiveObjectCount:    active,
		TotalBytes:           totalBytes,
		MultipartUploadCount: len(s.store.ListMultipartUploads()),
		LifecycleRuleCount:   len(s.store.ListLifecycleRules()),
		PolicyCount:          len(s.store.ListBucketPolicies()),
		AccessLogCount:       len(s.store.ListAccessLogs()),
		GeneratedAt:          time.Now(),
	}
}

// StatsByBucket 按桶统计对象数与容量。
func (s *Service) StatsByBucket() []*BucketStat {
	buckets := s.store.ListBuckets()
	objects := s.store.ListObjects()
	nameByID := make(map[string]string, len(buckets))
	regionByID := make(map[string]string, len(buckets))
	for _, b := range buckets {
		nameByID[b.ID] = b.Name
		regionByID[b.ID] = b.Region
	}
	quotaByBucket := make(map[string]int64)
	for _, q := range s.store.ListBucketQuotas() {
		quotaByBucket[q.BucketID] = q.QuotaBytes
	}
	countByBucket := make(map[string]int)
	bytesByBucket := make(map[string]int64)
	for _, o := range objects {
		countByBucket[o.BucketID]++
		bytesByBucket[o.BucketID] += o.Size
	}
	result := make([]*BucketStat, 0, len(buckets))
	for _, b := range buckets {
		result = append(result, &BucketStat{
			BucketID:    b.ID,
			BucketName:  b.Name,
			Region:      b.Region,
			ObjectCount: countByBucket[b.ID],
			TotalBytes:  bytesByBucket[b.ID],
			UsedBytes:   s.usedBytesOf(b.ID),
			QuotaBytes:  quotaByBucket[b.ID],
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalBytes > result[j].TotalBytes
	})
	return result
}

// StatsByRegion 按区域统计桶数、对象数与容量。
func (s *Service) StatsByRegion() []*RegionStat {
	buckets := s.store.ListBuckets()
	objects := s.store.ListObjects()
	regionByBucket := make(map[string]string, len(buckets))
	for _, b := range buckets {
		regionByBucket[b.ID] = b.Region
	}
	byRegion := make(map[string]*RegionStat)
	for _, b := range buckets {
		st, ok := byRegion[b.Region]
		if !ok {
			st = &RegionStat{Region: b.Region}
			byRegion[b.Region] = st
		}
		st.BucketCount++
	}
	for _, o := range objects {
		region := regionByBucket[o.BucketID]
		st, ok := byRegion[region]
		if !ok {
			st = &RegionStat{Region: region}
			byRegion[region] = st
		}
		st.ObjectCount++
		st.TotalBytes += o.Size
	}
	result := make([]*RegionStat, 0, len(byRegion))
	for _, st := range byRegion {
		result = append(result, st)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalBytes > result[j].TotalBytes
	})
	return result
}

// TopAccessBuckets 返回访问量 TOP N 的桶。
func (s *Service) TopAccessBuckets(limit int) []*AccessRank {
	logs := s.store.ListAccessLogs()
	countByBucket := make(map[string]int)
	bytesByBucket := make(map[string]int64)
	for _, l := range logs {
		countByBucket[l.BucketID]++
		bytesByBucket[l.BucketID] += l.Size
	}
	nameByID := make(map[string]string)
	for _, b := range s.store.ListBuckets() {
		nameByID[b.ID] = b.Name
	}
	ranks := make([]*AccessRank, 0, len(countByBucket))
	for bucketID, count := range countByBucket {
		ranks = append(ranks, &AccessRank{
			BucketID:    bucketID,
			BucketName:  nameByID[bucketID],
			AccessCount: count,
			TotalBytes:  bytesByBucket[bucketID],
		})
	}
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].AccessCount > ranks[j].AccessCount
	})
	if limit > 0 && len(ranks) > limit {
		ranks = ranks[:limit]
	}
	return ranks
}

// LifecycleRuleCount 统计生命周期规则数量（按桶分组）。
func (s *Service) LifecycleRuleCount() map[string]int {
	result := make(map[string]int)
	for _, r := range s.store.ListLifecycleRules() {
		result[r.BucketID]++
	}
	return result
}

// usedBytesOf 返回桶的已用容量（配额记录口径）。
func (s *Service) usedBytesOf(bucketID string) int64 {
	if q, err := s.store.GetBucketQuotaByBucket(bucketID); err == nil {
		return q.UsedBytes
	}
	return 0
}

// Snapshot 汇总导出快照。
type Snapshot struct {
	GeneratedAt  time.Time       `json:"generated_at"`
	Overview     *Overview       `json:"overview"`
	Buckets      []*model.Bucket      `json:"buckets"`
	ByBucket     []*BucketStat   `json:"by_bucket"`
	ByRegion     []*RegionStat   `json:"by_region"`
	TopAccess    []*AccessRank   `json:"top_access"`
	ObjectCounts int             `json:"object_counts"`
}

// ExportSnapshot 导出汇总快照。
func (s *Service) ExportSnapshot() *Snapshot {
	return &Snapshot{
		GeneratedAt:  time.Now(),
		Overview:     s.Overview(),
		Buckets:      s.store.ListBuckets(),
		ByBucket:     s.StatsByBucket(),
		ByRegion:     s.StatsByRegion(),
		TopAccess:    s.TopAccessBuckets(10),
		ObjectCounts: len(s.store.ListObjects()),
	}
}
