package service

import (
	"fmt"

	"task209-deformgate/internal/field"
	"task209-deformgate/internal/model"
	"task209-deformgate/internal/store"
)

// SmokeResult 是端到端自检的结构化报告。
type SmokeResult struct {
	Steps []string    `json:"steps"`
	Stats store.Stats `json:"stats"`
}

// RunSmokeTest 执行端到端自检：创建影像对 → 追加折叠场并复现 reject →
// 替换健康场复现 pass → 发布 → 关闭重开数据库验证恢复与幂等。
// 任一步失败即返回错误（非 0 退出码，Docker 验证契约）。
func RunSmokeTest(dbPath, smokeDir string) (*SmokeResult, error) {
	res := &SmokeResult{}

	st, err := store.Open(dbPath)
	if err != nil {
		return nil, err
	}
	app, err := NewApp(st)
	if err != nil {
		return nil, err
	}

	dims := model.Dims{8, 8, 4}
	spacing := []float64{1.0, 1.0, 1.0}

	// 1. 登记影像对。
	pair, err := app.CreateImagePair("smoke-pair", dims, spacing, "RAS", "MRI")
	if err != nil {
		return nil, fmt.Errorf("创建影像对: %w", err)
	}
	res.Steps = append(res.Steps, fmt.Sprintf("影像对 #%d 已登记（%dx%dx%d, %s）", pair.ID, dims[0], dims[1], dims[2], pair.Axes))

	// 2. 追加含局部折叠的形变场（中心体素制造反向位移 → det(J)≤0）。
	foldedDisp := field.SynthesizeLocalFold(dims, 4*8*4+4*8+4)
	folded, err := app.AttachField(pair.ID, dims, foldedDisp)
	if err != nil {
		return nil, fmt.Errorf("追加折叠场: %w", err)
	}
	res.Steps = append(res.Steps, fmt.Sprintf("折叠场 #%d 已追加（hash=%s…）", folded.ID, folded.Hash[:12]))

	// 3. 添加逆一致性采样点（自洽点：forward+inverse≈0）。
	if _, err := app.AddSample(folded.ID, model.Vec3{X: 2, Y: 2, Z: 1},
		model.Vec3{X: 0.5, Y: -0.3, Z: 0.1},
		model.Vec3{X: -0.5, Y: 0.3, Z: -0.1}); err != nil {
		return nil, fmt.Errorf("添加采样点: %w", err)
	}

	// 4. 创建并运行检查 → 折叠场应判定 reject。
	t1, err := app.CreateCheck(pair.ID, folded.ID)
	if err != nil {
		return nil, fmt.Errorf("创建检查任务: %w", err)
	}
	r1, err := app.RunCheck(t1.ID)
	if err != nil {
		return nil, fmt.Errorf("运行检查(折叠): %w", err)
	}
	res.Steps = append(res.Steps, fmt.Sprintf("折叠场检查 #%d 判定=%s，折叠体素=%d", t1.ID, r1.Verdict, r1.Jacobian.FoldCount))
	if r1.Verdict != model.VerdictReject {
		return nil, fmt.Errorf("期望折叠场判定 reject，得到 %s", r1.Verdict)
	}

	// 5. 替换健康仿射场（scale 0.02，雅可比≈1.06，无折叠）。
	okDisp := field.SynthesizeAffine(dims, 0.02, model.Vec3{})
	ok, err := app.AttachField(pair.ID, dims, okDisp)
	if err != nil {
		return nil, fmt.Errorf("追加健康场: %w", err)
	}
	// 幂等验证：相同内容重复追加返回同一记录。
	ok2, err := app.AttachField(pair.ID, dims, okDisp)
	if err != nil {
		return nil, err
	}
	if ok2.ID != ok.ID {
		return nil, fmt.Errorf("幂等失败：重复追加返回 %d ≠ %d", ok2.ID, ok.ID)
	}
	res.Steps = append(res.Steps, fmt.Sprintf("健康场 #%d 已追加，幂等重复返回同一 #%d", ok.ID, ok2.ID))

	// 6. 健康场检查 → pass。
	t2, err := app.CreateCheck(pair.ID, ok.ID)
	if err != nil {
		return nil, err
	}
	r2, err := app.RunCheck(t2.ID)
	if err != nil {
		return nil, fmt.Errorf("运行检查(健康): %w", err)
	}
	res.Steps = append(res.Steps, fmt.Sprintf("健康场检查 #%d 判定=%s，雅可比[min=%.3f max=%.3f]", t2.ID, r2.Verdict, r2.Jacobian.Min, r2.Jacobian.Max))

	// 7. 发布健康结果。
	pub, err := app.PublishResult(r2.ID)
	if err != nil {
		return nil, fmt.Errorf("发布结果: %w", err)
	}
	res.Steps = append(res.Steps, fmt.Sprintf("结果 #%d 已发布（status=%s）", pub.ID, pub.Status))

	// 8. 关闭并重新打开同一数据库，验证重启恢复。
	if err := st.Close(); err != nil {
		return nil, err
	}
	st2, err := store.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("重开数据库: %w", err)
	}
	defer st2.Close()

	pairAfter, err := st2.GetImagePair(pair.ID)
	if err != nil {
		return nil, err
	}
	res.Steps = append(res.Steps, fmt.Sprintf("重开后影像对 #%d 状态=%s（发布结果绑定）", pairAfter.ID, pairAfter.Status))

	// 折叠证据仍可查询。
	folds, err := st2.ListFolds(r1.ID)
	if err != nil {
		return nil, err
	}
	if len(folds) == 0 {
		return nil, fmt.Errorf("重开后折叠证据丢失")
	}
	res.Steps = append(res.Steps, fmt.Sprintf("重开后折叠证据仍在（%d 条）", len(folds)))

	// 旧结果仍可查询。
	oldRes, err := st2.GetResult(r1.ID)
	if err != nil {
		return nil, err
	}
	res.Steps = append(res.Steps, fmt.Sprintf("折叠场旧结果 #%d 仍可查询（verdict=%s, status=%s）", oldRes.ID, oldRes.Verdict, oldRes.Status))

	stats, err := st2.Stats()
	if err != nil {
		return nil, err
	}
	res.Stats = stats
	return res, nil
}
