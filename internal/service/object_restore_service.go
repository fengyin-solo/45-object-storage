package service

import (
	"sort"
	"time"

	"objectstore/internal/model"
)

// RestoreObjectVersion 从指定历史版本恢复对象内容（软删除对象也可恢复为活跃）。
func (s *Service) RestoreObjectVersion(objectID, versionID string) (*model.Object, error) {
	o, err := s.store.GetObject(objectID)
	if err != nil {
		return nil, err
	}
	v, err := s.store.GetObjectVersion(versionID)
	if err != nil {
		return nil, err
	}
	if v.ObjectID != objectID {
		return nil, model.NewValidationError("version", "版本不属于该对象")
	}
	if v.IsDeleteMarker {
		return nil, model.NewValidationError("version", "删除标记版本不可用于恢复")
	}
	// 已删除对象恢复到活跃状态，无需经过状态机校验（恢复为特殊操作）。
	o.Size = v.Size
	o.ETag = v.ETag
	o.Status = model.ObjectActive
	o.UpdatedAt = time.Now()
	if err := s.store.UpdateObject(o); err != nil {
		return nil, err
	}
	// 恢复后追加一个新版本记录，保留审计轨迹。
	if err := s.appendVersion(o, v.Size, v.ETag, false, time.Now()); err != nil {
		return nil, err
	}
	s.adjustQuota(o.BucketID, v.Size)
	return o, nil
}

// LatestActiveVersion 返回对象最近的非删除标记版本。
func (s *Service) LatestActiveVersion(objectID string) (*model.ObjectVersion, error) {
	var versions []*model.ObjectVersion
	for _, v := range s.store.ListObjectVersions() {
		if v.ObjectID == objectID {
			versions = append(versions, v)
		}
	}
	if len(versions) == 0 {
		return nil, model.NewValidationError("version", "对象无可用版本")
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].VersionNo > versions[j].VersionNo
	})
	for _, v := range versions {
		if !v.IsDeleteMarker {
			return v, nil
		}
	}
	return nil, model.NewValidationError("version", "对象仅剩删除标记版本")
}

// PurgeOldVersions 清理对象保留版本之外的旧版本（保留最近 N 个非删除标记版本）。
func (s *Service) PurgeOldVersions(objectID string, keep int) (int, error) {
	if _, err := s.store.GetObject(objectID); err != nil {
		return 0, err
	}
	if keep < 1 {
		return 0, model.NewValidationError("keep", "保留版本数至少为 1")
	}
	var versions []*model.ObjectVersion
	for _, v := range s.store.ListObjectVersions() {
		if v.ObjectID == objectID {
			versions = append(versions, v)
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].VersionNo > versions[j].VersionNo
	})
	removed := 0
	activeKept := 0
	for _, v := range versions {
		if v.IsDeleteMarker {
			continue
		}
		if activeKept < keep {
			activeKept++
			continue
		}
		if err := s.store.DeleteObjectVersion(v.ID); err == nil {
			removed++
		}
	}
	return removed, nil
}
