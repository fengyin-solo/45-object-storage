package service

import (
	"sort"

	"objectstore/internal/model"
)

// OperationStat 按操作类型聚合的访问统计。
type OperationStat struct {
	Operation   string `json:"operation"`
	Count       int    `json:"count"`
	SuccessCount int   `json:"success_count"`
	DeniedCount int    `json:"denied_count"`
	TotalBytes  int64  `json:"total_bytes"`
}

// OperationStats 按操作类型聚合访问日志。
func (s *Service) OperationStats() []*OperationStat {
	byOp := make(map[string]*OperationStat)
	for _, l := range s.store.ListAccessLogs() {
		st, ok := byOp[l.Operation]
		if !ok {
			st = &OperationStat{Operation: l.Operation}
			byOp[l.Operation] = st
		}
		st.Count++
		st.TotalBytes += l.Size
		if l.Result == model.LogResultSuccess {
			st.SuccessCount++
		} else {
			st.DeniedCount++
		}
	}
	out := make([]*OperationStat, 0, len(byOp))
	for _, st := range byOp {
		out = append(out, st)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Count > out[j].Count
	})
	return out
}

// ContentTypeStat 按 ContentType 分布的对象统计。
type ContentTypeStat struct {
	ContentType string `json:"content_type"`
	Count       int    `json:"count"`
	TotalBytes  int64  `json:"total_bytes"`
}

// ContentTypeDistribution 统计对象按 ContentType 的分布。
func (s *Service) ContentTypeDistribution() []*ContentTypeStat {
	byType := make(map[string]*ContentTypeStat)
	for _, o := range s.store.ListObjects() {
		ct := o.ContentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		st, ok := byType[ct]
		if !ok {
			st = &ContentTypeStat{ContentType: ct}
			byType[ct] = st
		}
		st.Count++
		st.TotalBytes += o.Size
	}
	out := make([]*ContentTypeStat, 0, len(byType))
	for _, st := range byType {
		out = append(out, st)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].TotalBytes > out[j].TotalBytes
	})
	return out
}

// TopObjectsBySize 返回大小最大的 N 个对象。
func (s *Service) TopObjectsBySize(limit int) []*model.Object {
	all := s.store.ListObjects()
	sort.Slice(all, func(i, j int) bool {
		return all[i].Size > all[j].Size
	})
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all
}

// DailyAccessStat 按日聚合的访问统计。
type DailyAccessStat struct {
	Date       string `json:"date"`
	Count      int    `json:"count"`
	TotalBytes int64  `json:"total_bytes"`
}

// DailyAccessTrend 按日期聚合访问日志趋势（YYYY-MM-DD）。
func (s *Service) DailyAccessTrend() []*DailyAccessStat {
	byDate := make(map[string]*DailyAccessStat)
	for _, l := range s.store.ListAccessLogs() {
		date := l.CreatedAt.Format("2006-01-02")
		st, ok := byDate[date]
		if !ok {
			st = &DailyAccessStat{Date: date}
			byDate[date] = st
		}
		st.Count++
		st.TotalBytes += l.Size
	}
	out := make([]*DailyAccessStat, 0, len(byDate))
	for _, st := range byDate {
		out = append(out, st)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Date < out[j].Date
	})
	return out
}

// UploadStatusStat 分片上传状态分布。
type UploadStatusStat struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// UploadStatusDistribution 统计分片上传按状态的分布。
func (s *Service) UploadStatusDistribution() []*UploadStatusStat {
	byStatus := make(map[string]int)
	for _, u := range s.store.ListMultipartUploads() {
		byStatus[u.Status]++
	}
	out := make([]*UploadStatusStat, 0, len(byStatus))
	for status, count := range byStatus {
		out = append(out, &UploadStatusStat{Status: status, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Count > out[j].Count
	})
	return out
}

// BucketGrowthStat 桶容量增长统计（活跃对象累计）。
type BucketGrowthStat struct {
	BucketID   string `json:"bucket_id"`
	BucketName string `json:"bucket_name"`
	ObjectCount int   `json:"object_count"`
	TotalBytes int64  `json:"total_bytes"`
}

// BucketGrowth 返回按容量增长的桶排行（与 StatsByBucket 互补，仅统计活跃对象）。
func (s *Service) BucketGrowth() []*BucketGrowthStat {
	nameByID := make(map[string]string)
	for _, b := range s.store.ListBuckets() {
		nameByID[b.ID] = b.Name
	}
	countByBucket := make(map[string]int)
	bytesByBucket := make(map[string]int64)
	for _, o := range s.store.ListObjects() {
		if o.Status != model.ObjectActive {
			continue
		}
		countByBucket[o.BucketID]++
		bytesByBucket[o.BucketID] += o.Size
	}
	out := make([]*BucketGrowthStat, 0, len(countByBucket))
	for bucketID, count := range countByBucket {
		out = append(out, &BucketGrowthStat{
			BucketID:    bucketID,
			BucketName:  nameByID[bucketID],
			ObjectCount: count,
			TotalBytes:  bytesByBucket[bucketID],
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].TotalBytes > out[j].TotalBytes
	})
	return out
}

// VersionCount 统计对象版本总数与删除标记数。
func (s *Service) VersionCount() (int, int) {
	total := 0
	markers := 0
	for _, v := range s.store.ListObjectVersions() {
		total++
		if v.IsDeleteMarker {
			markers++
		}
	}
	return total, markers
}
