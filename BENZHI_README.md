# BENZHI 评测说明：task209-deformgate

## 构建与运行

```bash
# 构建镜像（双架构）
bash build_benzhi_docker.sh task209-deformgate linux/amd64
bash build_benzhi_docker.sh task209-deformgate linux/arm64

# 自检（唯一判据，退出码 0 为通过）
docker run --rm task209-deformgate --smoke-test

# 长驻服务
docker run --rm -p 8080:8080 task209-deformgate --addr :8080 --db /tmp/deformgate.db
```

## --smoke-test 契约

`--smoke-test` 不启动长驻服务，而是执行端到端自检后以 0 退出码结束：

1. 登记影像对（8x8x4，RAS，spacing 1mm）；
2. 追加含局部折叠的形变场 → 检查判定 `reject`，折叠体素 22 个；
3. 追加健康仿射场 → 幂等去重返回同一记录；
4. 健康场检查判定 `pass`，雅可比行列式约 `1.03~1.06`；
5. 发布健康结果；
6. 关闭并重开数据库，验证影像对状态、折叠证据、旧结果均恢复。

任一步失败即非 0 退出，Docker 双架构验证据此判据。

## API 一览（前缀 /api）

| 分组 | 端点 |
| --- | --- |
| 影像对 | `POST/GET /api/imagepairs`、`GET/PUT /api/imagepairs/{id}`、`POST /api/imagepairs/{id}/archive` |
| 形变场 | `POST /api/imagepairs/{id}/fields`、`GET /api/imagepairs/{id}/fields`、`GET /api/fields/{id}`、`GET /api/fields/{id}/hash`、`POST/GET /api/fields/{id}/samples` |
| 检查 | `POST/GET /api/checks`、`GET /api/checks/{id}`、`POST /api/checks/{id}/run`、`GET /api/checks/{id}/metrics`、`GET /api/checks/{id}/folds` |
| 参数 | `GET /api/params`、`GET /api/params/active`、`POST /api/params`、`POST /api/params/{id}/activate` |
| 结果 | `POST /api/results/{id}/publish`、`GET /api/results`、`GET /api/results/{id}`、`GET /api/imagepairs/{id}/results`、`GET /api/imagepairs/{id}/version-chain` |
| 元信息 | `GET /api/health`、`GET /api/stats`、`GET /api/version` |

## 组件版本

- Go：`1.26.3`
- SQLite：`3.46.1`（`modernc.org/sqlite v1.52.0`）
- 构建镜像：`golang:1.26.3-bookworm`（单阶段，`CGO_ENABLED=0`）
