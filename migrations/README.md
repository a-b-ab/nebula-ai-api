# 数据库迁移说明

本项目使用 [go-migrate](https://github.com/golang-migrate/migrate) 进行数据库迁移管理。

## 迁移文件命名规范

迁移文件应该成对创建，格式如下：
- `XXXXXX_description.up.sql` - 应用迁移
- `XXXXXX_description.down.sql` - 回滚迁移

其中 `XXXXXX` 是版本号（通常从 000001 开始），`description` 是对迁移的描述。

示例：
- `000001_create_articles_table.up.sql`
- `000001_create_articles_table.down.sql`

## 使用方法

### 1. 启动数据库
```bash
docker-compose up -d postgres
```

### 2. 运行所有迁移
```bash
make migrate-up
# 或
docker-compose run --rm migrate
```

### 3. 回滚所有迁移
```bash
make migrate-down
```

### 4. 创建新的迁移
```bash
# 创建迁移文件
migrate create -ext sql -dir ./migrations -seq create_users_table

# 这会创建两个文件：
# - migrations/000002_create_users_table.up.sql
# - migrations/000002_create_users_table.down.sql
```

### 5. 查看当前迁移版本
```bash
make migrate-version
```

## Docker Compose 服务说明

`migrate` 服务会在启动时自动运行所有迁移。如果你不想在启动时自动运行，可以：

```bash
# 不自动运行迁移
docker-compose up -d postgres redis

# 手动运行迁移
make migrate-up
```

## 本地安装 go-migrate（可选）

如果你想在本机安装 go-migrate：

```bash
# macOS
brew install golang-migrate

# 或使用 go install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

安装后可以直接运行：
```bash
migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/nebula?sslmode=disable" up
```
