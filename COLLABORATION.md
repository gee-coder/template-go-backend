# 协作流程说明

## 1. 目标

本文件用于约束仓库在 GitHub 上的日常协作方式，帮助模板仓库在多人维护、多人复用的情况下保持一致性。

## 2. 建议启用的协作能力

如果你是 fork 后自行维护的团队，建议在 GitHub 中启用以下能力：

- Discussions：承接问答、想法、展示与公告
- Projects：管理需求、缺陷与版本排期
- Releases：对外记录版本说明与升级提示

## 3. 建议的 Discussions 分类

- `Q&A`：使用问题、排障讨论
- `Ideas`：通用能力建议、体验优化建议
- `Show and tell`：模板二次开发成果展示
- `Announcements`：版本计划、兼容性变更通知

## 4. 建议的 Projects 看板泳道

- `Inbox`：新进入的需求、缺陷、问题
- `Ready`：已完成分流与优先级确认
- `In Progress`：正在开发或修复
- `Review`：待代码评审或待验证
- `Blocked`：存在依赖或外部阻塞
- `Done`：已完成并合入

## 5. 标签治理建议

建议采用三层标签结构：

- 类型：`type:bug`、`type:feature`、`type:docs`、`type:chore`、`type:breaking`
- 状态：`status:needs-triage`、`status:ready`、`status:blocked`
- 优先级：`priority:p1`、`priority:p2`、`priority:p3`

后端模板额外建议保留的领域标签：

- `area:api`
- `area:auth`
- `area:rbac`
- `area:infra`

## 6. Pull Request 协作链路

1. 开发者提交功能分支并发起 Pull Request
2. `CODEOWNERS` 自动请求维护者评审
3. `pr-labeler` 与 `release-drafter` 自动补充标签与版本草稿信息
4. CI 通过后进入人工评审与验证
5. 合并后由 Release Drafter 持续更新版本草稿

## 7. 发布协作建议

- 正式发布前更新 [CHANGELOG.md](./CHANGELOG.md)
- 参照 [RELEASE.md](./RELEASE.md) 做构建、测试与兼容性确认
- 如涉及前后端契约变化，需同步通知其他模板仓库维护者
