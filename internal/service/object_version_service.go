package service

import (
	"sort"

	"objectstore/internal/model"
)

// GetObjectVersion 按 ID 查询对象版本。
func (s *Service) GetObjectVersion(id string) (*model.ObjectVersion, error) {
	return s.store.GetObjectVersion(id)
}

// ListObjectVersions 分页 + 多条件筛选查询对象版本。
func (s *Service) ListObjectVersions(filter model.ObjectVersionFilter, page, size int) ([]*model.ObjectVersion, int, error) {
	all := s.store.ListObjectVersions()
	matched := make([]*model.ObjectVersion, 0, len(all))
	for _, v := range all {
		if filter.Match(v) {
			matched = append(matched, v)
		}
	}
	// 先按对象 ID 分组，再按版本号降序排序。
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].ObjectID != matched[j].ObjectID {
			return matched[i].ObjectID < matched[j].ObjectID
		}
		return matched[i].VersionNo > matched[j].VersionNo
	})
	if filter.LatestOnly {
		latest := make([]*model.ObjectVersion, 0, len(matched))
		seen := make(map[string]bool)
		for _, v := range matched {
			if !seen[v.ObjectID] {
				seen[v.ObjectID] = true
				latest = append(latest, v)
			}
		}
		matched = latest
	}
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.ObjectVersion{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}
