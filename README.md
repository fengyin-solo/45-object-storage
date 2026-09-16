# objectstore 对象存储服务

纯 Go 标准库实现的对象存储后端服务，零第三方依赖（仅 `net/http` + 标准库）。

## 功能特性

- **存储桶**：名称全局唯一、区域/所有者/配额管理、挂起/恢复状态机
- **对象**：桶内 Key 唯一、大小/ETag/ContentType、软删除（active→deleted）
- **版本管理**：同一 Key 覆盖写入自动生成新版本，删除生成 delete marker
- **分片上传**：init → upload part → complete（校验分片齐全、聚合大小）/ abort 完整流程
- **生命周期规则**：前缀过期/转换、启用/禁用状态机、批量更新状态
- **桶策略**：allow/deny 授权评估、启用/禁用
- **访问日志**：操作记录、多条件筛选
- **配额**：已用容量跟踪、超配额检测
- **多报表统计**：全局概览、按桶/区域统计、访问量 TOP N、生命周期规则计数
- **中间件**：X-API-Key 鉴权、固定窗口限流、请求日志、panic 恢复
- **数据导出**：`/api/export` 汇总快照 JSON
- **前端看板**：`web/` 纯 HTML+原生 JS，零 CDN

## 运行

```bash
# 进入项目目录
cd origin

# 启动（默认监听 :8080）
go run ./cmd/server

# 可选环境变量
PORT=9090 ADDR=:9090 API_KEY=my-key RATE_LIMIT_PER_MINUTE=600 MAX_PAGE_SIZE=100 LOG_LEVEL=info \
  go run ./cmd/server
```

浏览器访问 `http://localhost:8080/` 查看看板。

所有 `/api/*` 接口需在请求头携带 `X-API-Key: objectstore-secret-key`（默认值，可通过 `API_KEY` 覆盖）。

## 统一响应

```json
{"code":0,"message":"ok","data":...}
```

错误映射：`model.ValidationError`→400、`store.ErrNotFound`→404、`store.ErrConflict`/状态流转非法→409、鉴权失败→401、限流→429、其他→500。

## API 一览

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/buckets | 创建存储桶 |
| GET | /api/buckets | 分页查询桶（region/owner/status/keyword 筛选） |
| GET | /api/buckets/{id} | 查询桶 |
| PUT | /api/buckets/{id} | 更新桶（区域/所有者/配额） |
| DELETE | /api/buckets/{id} | 删除桶（桶内无对象） |
| POST | /api/buckets/{id}/suspend | 挂起桶（active→suspended） |
| POST | /api/buckets/{id}/activate | 恢复桶（suspended→active） |
| PUT | /api/buckets/{id}/quota | 调整桶配额 |
| POST | /api/objects | 写入/覆盖对象（覆盖生成新版本） |
| GET | /api/objects | 分页查询对象（bucket_id/status/keyword/min_size/max_size） |
| GET | /api/objects/{id} | 查询对象 |
| PUT | /api/objects/{id} | 更新对象元数据 |
| DELETE | /api/objects/{id} | 软删除对象（生成 delete marker） |
| POST | /api/objects/batch-delete | 批量删除对象 |
| GET | /api/object-versions | 分页查询版本（object_id/bucket_id/latest_only） |
| GET | /api/object-versions/{id} | 查询版本 |
| POST | /api/lifecycle-rules | 创建生命周期规则 |
| GET | /api/lifecycle-rules | 分页查询规则（bucket_id/status/prefix） |
| GET | /api/lifecycle-rules/{id} | 查询规则 |
| PUT | /api/lifecycle-rules/{id} | 更新规则 |
| DELETE | /api/lifecycle-rules/{id} | 删除规则 |
| POST | /api/lifecycle-rules/{id}/status | 切换规则状态 |
| POST | /api/lifecycle-rules/batch-status | 批量更新规则状态 |
| POST | /api/multipart-uploads | 初始化分片上传 |
| GET | /api/multipart-uploads | 分页查询上传（bucket_id/status/keyword） |
| GET | /api/multipart-uploads/{id} | 查询上传 |
| POST | /api/multipart-uploads/{id}/parts | 上传分片 |
| GET | /api/multipart-uploads/{id}/parts | 查询上传的全部分片 |
| POST | /api/multipart-uploads/{id}/complete | 完成上传（校验分片齐全） |
| POST | /api/multipart-uploads/{id}/abort | 中止上传 |
| GET | /api/upload-parts | 分页查询分片（upload_id/min_part_number/max_part_number） |
| GET | /api/upload-parts/{id} | 查询分片 |
| POST | /api/bucket-policies | 创建桶策略 |
| GET | /api/bucket-policies | 分页查询策略（bucket_id/principal/effect/status） |
| GET | /api/bucket-policies/{id} | 查询策略 |
| PUT | /api/bucket-policies/{id} | 更新策略 |
| DELETE | /api/bucket-policies/{id} | 删除策略 |
| POST | /api/bucket-policies/{id}/status | 切换策略状态 |
| POST | /api/bucket-policies/evaluate | 评估授权（deny 优先） |
| POST | /api/access-logs | 记录访问日志 |
| GET | /api/access-logs | 分页查询日志（bucket_id/object_key/operation/ip/result） |
| GET | /api/access-logs/{id} | 查询日志 |
| DELETE | /api/access-logs/{id} | 删除日志 |
| POST | /api/bucket-quotas | 创建配额 |
| GET | /api/bucket-quotas | 分页查询配额（bucket_id/exceeded_only） |
| GET | /api/bucket-quotas/{id} | 查询配额 |
| PUT | /api/bucket-quotas/{id} | 更新配额 |
| DELETE | /api/bucket-quotas/{id} | 删除配额 |
| GET | /api/stats/overview | 全局概览统计 |
| GET | /api/stats/by-bucket | 按桶统计对象数与容量 |
| GET | /api/stats/by-region | 按区域统计 |
| GET | /api/stats/top-access?limit=10 | 访问量 TOP N |
| GET | /api/stats/lifecycle-count | 生命周期规则数量 |
| GET | /api/export | 导出汇总快照 JSON |

## 测试

```bash
go test ./...
```

## 目录结构

```
origin/
├── go.mod
├── cmd/server/main.go
├── internal/
│   ├── config/config.go
│   ├── app/app.go
│   ├── model/          # 实体 + 状态机 + 校验 + 筛选
│   ├── store/          # 接口 + 内存实现
│   ├── service/        # 业务逻辑 + 统计
│   └── handler/        # 路由 + 中间件 + 静态服务
├── pkg/httpx|idgen|logger
└── web/                # 前端看板
```
