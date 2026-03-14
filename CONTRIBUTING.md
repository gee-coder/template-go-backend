# 贡献指南

欢迎基于此模板继续扩展业务能力。为了让仓库长期保持清晰、可维护，请在提交前遵循以下约定。

## 1. 基本原则

- 保持分层边界：`Handler -> Service -> Repository`
- 不要把业务逻辑写进 `cmd/api/main.go`
- 不要在 Handler 中直接操作数据库
- 新增接口时，优先补请求结构体、Service 与 Repository 抽象

## 2. 本地开发流程

1. 拉取最新代码
2. 创建功能分支
3. 修改代码并运行测试
4. 提交代码并发起 PR

推荐分支命名：

- `feat/<name>`
- `fix/<name>`
- `docs/<name>`
- `refactor/<name>`

## 3. 提交前检查

请至少确保以下命令通过：

```powershell
go test ./...
go build ./cmd/api
```

## 4. 代码风格建议

- 保持 Go 文件单一职责
- Handler 只负责参数校验、调用 Service、返回响应
- Service 负责业务规则
- Repository 负责数据持久化
- 新增结构体时尽量补充注释，方便生成文档

## 5. PR 检查清单

- 是否遵守分层规范
- 是否补充或更新了 README / 文档
- 是否补充了必要的测试
- 是否影响已有接口契约
- 是否需要更新配置说明或部署说明

## 6. 提交规范

请参考 [`COMMIT_CONVENTION.md`](./COMMIT_CONVENTION.md)

