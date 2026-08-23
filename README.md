# task209-deformgate：医学影像配准形变场质量门服务

面向**医学影像工程师**的纯后端 Go 服务：在下游分割/分析之前，判断一对影像的配准结果
是否具备可信的形变场。服务登记影像对与形变场，计算三类质量指标并给出 pass / review /
reject 判定，支持参数版本锁定与结果版本发布。

## 业务闭环

1. 登记影像对（维度、体素间距、坐标方向约定）；
2. 追加形变场（位移向量场，按内容哈希幂等去重）；
3. 添加逆一致性采样点（正/逆变换位移）；
4. 创建并运行检查任务，计算三组指标并产出判定；
5. 发布质量结果（不可变，新结果显式替代旧结果）。

## 三类质量指标

- **雅可比行列式** `det(J)`：`J = I + ∇u`。`det(J) ≤ 0` 表示局部折叠（配准不可逆）；
  超出收缩/膨胀阈值记为异常体积变化。
- **逆一致性误差**：采样点处 `|正变换位移 + 逆变换位移|`，衡量正逆变换自洽性。
- **边界覆盖率**：影像边界体素位移向量有效的比例，缺失视为缺边界。

## 标准命令

```bash
# 构建 / 静态检查 / 测试
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...

# 运行服务
go run ./cmd/deformgate --addr :8080 --db ./deformgate.db

# 端到端自检（Docker 验证契约，退出码 0 为通过）
go run ./cmd/deformgate --smoke-test
```

## API 入口（前缀 /api）

- 影像对：`POST/GET /api/imagepairs`、`GET/PUT /api/imagepairs/{id}`、`POST /api/imagepairs/{id}/archive`
- 形变场：`POST /api/imagepairs/{id}/fields`、`GET /api/imagepairs/{id}/fields`、`GET /api/fields/{id}`、`GET /api/fields/{id}/hash`、`POST/GET /api/fields/{id}/samples`
- 检查：`POST/GET /api/checks`、`GET /api/checks/{id}`、`POST /api/checks/{id}/run`、`GET /api/checks/{id}/metrics`、`GET /api/checks/{id}/folds`
- 参数：`GET /api/params`、`GET /api/params/active`、`POST /api/params`、`POST /api/params/{id}/activate`
- 结果：`POST /api/results/{id}/publish`、`GET /api/results`、`GET /api/results/{id}`、`GET /api/imagepairs/{id}/results`、`GET /api/imagepairs/{id}/version-chain`
- 元信息：`GET /api/health`、`GET /api/stats`、`GET /api/version`

## 技术栈与持久化

- Go 1.26.3，纯 Go SQLite 驱动 `modernc.org/sqlite v1.52.0`（CGO 无关，离线可构建）。
- SQLite 建表：`image_pairs` / `deform_fields` / `sample_points` / `check_tasks` /
  `check_params` / `quality_results` / `fold_regions`。
- 重启恢复：中断的 `running` 任务回退为 `queued`，所有实体与已发布结果持久化。

## Docker 双架构

```bash
bash build_benzhi_docker.sh task209-deformgate linux/amd64
bash build_benzhi_docker.sh task209-deformgate linux/arm64
docker run --rm task209-deformgate --smoke-test
```
