# 数据库 SQL 说明

这个目录提供了和当前后端模板配套的一套显式 SQL 文件，方便团队成员在不阅读 Go 代码的前提下直接了解表结构、初始化数据库，或交给 DBA 审核。

## 文件说明

- `schema.sql`：建库和建表脚本，结构与当前 `GORM AutoMigrate` 实际产出保持一致。
- `seed.sql`：默认初始化数据，包括管理员账号、角色、菜单和 RBAC 关联关系。

## 默认初始化数据

- 数据库名：`nex_template`
- 默认账号：`admin`
- 默认密码：`Admin123!`

## 使用方式

### 方式一：手工导入 SQL

适合希望先准备数据库，再启动后端服务的场景。

```powershell
mysql -unex -pnex123456 < .\sql\schema.sql
mysql -unex -pnex123456 < .\sql\seed.sql
```

### 方式二：直接启动后端

后端入口仍然保留 `AutoMigrate + SeedInitialData`，即使不手工导入 SQL，服务首次启动时也会自动建表并写入默认数据。

## 注意事项

- `seed.sql` 面向空数据库初始化，推荐在全新库中执行。
- 如果你已经有同名表和数据，建议先备份，再评估是否需要手工调整主键和初始化数据。
