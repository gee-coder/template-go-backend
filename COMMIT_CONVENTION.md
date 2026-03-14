# 提交规范

建议采用 Conventional Commits 风格，便于后续生成变更日志与自动化发布。

## 1. 推荐格式

```text
type(scope): subject
```

示例：

```text
feat(auth): add refresh token rotation
fix(user): handle duplicate username error
docs(readme): add deployment diagram
refactor(service): split menu tree builder
```

## 2. 常用类型

- `feat`：新功能
- `fix`：缺陷修复
- `docs`：文档调整
- `refactor`：重构，不改变外部行为
- `test`：测试相关
- `chore`：杂项维护
- `build`：构建、依赖、打包相关
- `ci`：持续集成相关

## 3. Subject 书写建议

- 使用简短动词短语
- 尽量说明变更对象
- 不要只写“update”或“modify”

## 4. 推荐示例

- `feat(rbac): add role menu replacement`
- `fix(auth): return unauthorized on invalid refresh token`
- `docs(api): update startup instructions`

