.PHONY: migrate-up migrate-down migrate-create migrate-force

# 运行所有迁移
migrate-up:
	docker-compose run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/nebula?sslmode=disable" up

# 回滚所有迁移
migrate-down:
	docker-compose run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/nebula?sslmode=disable" down

# 创建新的迁移文件（使用：make migrate-create NAME=create_users_table）
migrate-create:
	docker-compose run --rm migrate create -ext sql -dir /migrations $(NAME)

# 强制设置版本（用于修复问题）
migrate-force:
	docker-compose run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/nebula?sslmode=disable" force $(VERSION)

# 查看当前迁移状态
migrate-version:
	docker-compose run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/nebula?sslmode=disable" version
