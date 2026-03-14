# 发布流程

## 1. 发布目标

本仓库建议使用语义化版本：

- `MAJOR`：引入破坏性变更
- `MINOR`：新增向后兼容的模板能力
- `PATCH`：修复缺陷、文档更新或非破坏性调整

## 2. 发布前检查清单

### 必做检查

- 更新 [CHANGELOG.md](./CHANGELOG.md)
- 确认 README、配置示例、Swagger 注释与当前实现一致
- 确认数据库迁移、初始化数据与 Redis 使用说明没有遗漏

### 本地校验命令

```powershell
go test ./...
go build ./cmd/api
```

### 容器校验建议

```powershell
docker compose up -d mysql redis
```

在具备 Docker 环境时，建议至少验证一次本地依赖服务启动与配置连通性。

## 3. 建议的发布步骤

1. 合并待发布改动到 `master`
2. 补齐 `CHANGELOG.md` 的本次版本说明
3. 执行本地测试与构建
4. 创建标签，例如 `v0.1.0`
5. 在 GitHub Release 中整理升级说明、风险提示与迁移事项

## 4. 发布说明建议包含

- 本次新增能力
- 修复的问题
- 是否包含破坏性变更
- 升级时需要修改的环境变量、配置或数据库结构
- 与前端仓库约定是否有变化
