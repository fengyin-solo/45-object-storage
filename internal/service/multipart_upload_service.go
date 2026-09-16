package service

import (
	"sort"
	"time"

	"objectstore/internal/model"
	"objectstore/internal/store"
	"objectstore/pkg/idgen"
)

// CompletePart 完成分片上传时客户端上报的分片信息。
type CompletePart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

// InitMultipartUpload 初始化分片上传。
func (s *Service) InitMultipartUpload(bucketID, key string) (*model.MultipartUpload, error) {
	if _, err := s.store.GetBucket(bucketID); err != nil {
		return nil, err
	}
	now := time.Now()
	u := &model.MultipartUpload{
		ID:          idgen.Hex(),
		BucketID:    bucketID,
		Key:         key,
		Status:      model.UploadInitiated,
		PartCount:   0,
		Size:        0,
		InitiatedAt: now,
	}
	if err := u.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.CreateMultipartUpload(u); err != nil {
		return nil, err
	}
	return u, nil
}

// GetMultipartUpload 按 ID 查询分片上传。
func (s *Service) GetMultipartUpload(id string) (*model.MultipartUpload, error) {
	return s.store.GetMultipartUpload(id)
}

// ListMultipartUploads 分页 + 多条件筛选查询分片上传。
func (s *Service) ListMultipartUploads(filter model.MultipartUploadFilter, page, size int) ([]*model.MultipartUpload, int, error) {
	all := s.store.ListMultipartUploads()
	matched := make([]*model.MultipartUpload, 0, len(all))
	for _, u := range all {
		if filter.Match(u) {
			matched = append(matched, u)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].InitiatedAt.After(matched[j].InitiatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.MultipartUpload{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UploadPart 上传单个分片：校验上传状态后登记分片并推进上传状态。
func (s *Service) UploadPart(uploadID string, partNumber int, size int64, etag string) (*model.UploadPart, error) {
	u, err := s.store.GetMultipartUpload(uploadID)
	if err != nil {
		return nil, err
	}
	if u.Status != model.UploadInitiated && u.Status != model.UploadUploading {
		return nil, model.NewValidationError("status", "上传已结束，无法继续上传分片")
	}
	now := time.Now()
	p := &model.UploadPart{
		ID:         idgen.Hex(),
		UploadID:   uploadID,
		PartNumber: partNumber,
		Size:       size,
		ETag:       etag,
		CreatedAt:  now,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.CreateUploadPart(p); err != nil {
		return nil, err
	}
	// 推进上传状态并累计分片数量与大小。
	if u.Status == model.UploadInitiated {
		if !model.MultipartCanTransition(u.Status, model.UploadUploading) {
			return nil, store.ErrStateTransition
		}
		u.Status = model.UploadUploading
	}
	u.PartCount++
	u.Size += size
	if err := s.store.UpdateMultipartUpload(u); err != nil {
		return nil, err
	}
	return p, nil
}

// CompleteMultipartUpload 完成分片上传：校验分片齐全后聚合大小并生成对象。
func (s *Service) CompleteMultipartUpload(uploadID string, parts []CompletePart) (*model.MultipartUpload, error) {
	u, err := s.store.GetMultipartUpload(uploadID)
	if err != nil {
		return nil, err
	}
	if !model.MultipartCanTransition(u.Status, model.UploadCompleted) {
		return nil, store.ErrStateTransition
	}
	uploaded := s.partsOfUpload(uploadID)
	if len(uploaded) == 0 {
		return nil, model.NewValidationError("parts", "尚未上传任何分片")
	}
	// 校验上报分片数量与已存分片一致。
	if len(parts) != len(uploaded) {
		return nil, model.NewValidationError("parts", "上报分片数量与已上传分片不一致")
	}
	// 校验编号连续完整（1..N）且 ETag 匹配。
	etagByNumber := make(map[int]string, len(uploaded))
	maxNumber := 0
	for _, p := range uploaded {
		etagByNumber[p.PartNumber] = p.ETag
		if p.PartNumber > maxNumber {
			maxNumber = p.PartNumber
		}
	}
	if maxNumber != len(uploaded) {
		return nil, model.NewValidationError("parts", "分片编号不连续")
	}
	for i := 1; i <= maxNumber; i++ {
		if _, ok := etagByNumber[i]; !ok {
			return nil, model.NewValidationError("parts", "缺少第 "+itoa(i)+" 个分片")
		}
	}
	for _, cp := range parts {
		etag, ok := etagByNumber[cp.PartNumber]
		if !ok {
			return nil, model.NewValidationError("parts", "上报了不存在的分片编号")
		}
		if etag != cp.ETag {
			return nil, model.NewValidationError("parts", "分片 ETag 校验不匹配")
		}
	}
	// 聚合大小。
	var total int64
	for _, p := range uploaded {
		total += p.Size
	}
	now := time.Now()
	u.Size = total
	u.PartCount = len(uploaded)
	u.Status = model.UploadCompleted
	u.CompletedAt = &now
	if err := s.store.UpdateMultipartUpload(u); err != nil {
		return nil, err
	}
	// 生成对象（key 为上传 key，取聚合大小）。
	if _, err := s.PutObject(u.BucketID, u.Key, "application/octet-stream", idgen.HexN(16), total); err != nil {
		return nil, err
	}
	return u, nil
}

// AbortMultipartUpload 中止分片上传，并清理已上传分片。
func (s *Service) AbortMultipartUpload(uploadID string) (*model.MultipartUpload, error) {
	u, err := s.store.GetMultipartUpload(uploadID)
	if err != nil {
		return nil, err
	}
	if !model.MultipartCanTransition(u.Status, model.UploadAborted) {
		return nil, store.ErrStateTransition
	}
	u.Status = model.UploadAborted
	if err := s.store.UpdateMultipartUpload(u); err != nil {
		return nil, err
	}
	for _, p := range s.partsOfUpload(uploadID) {
		_ = s.store.DeleteUploadPart(p.ID)
	}
	return u, nil
}

// partsOfUpload 返回指定上传的全部分片（按编号升序）。
func (s *Service) partsOfUpload(uploadID string) []*model.UploadPart {
	var list []*model.UploadPart
	for _, p := range s.store.ListUploadParts() {
		if p.UploadID == uploadID {
			list = append(list, p)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].PartNumber < list[j].PartNumber
	})
	return list
}

// ListUploadParts 分页 + 多条件筛选查询分片。
func (s *Service) ListUploadParts(filter model.UploadPartFilter, page, size int) ([]*model.UploadPart, int, error) {
	all := s.store.ListUploadParts()
	matched := make([]*model.UploadPart, 0, len(all))
	for _, p := range all {
		if filter.Match(p) {
			matched = append(matched, p)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].PartNumber < matched[j].PartNumber
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.UploadPart{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// GetUploadPart 按 ID 查询分片。
func (s *Service) GetUploadPart(id string) (*model.UploadPart, error) {
	return s.store.GetUploadPart(id)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
