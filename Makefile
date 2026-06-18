# =============================================================================
# Makefile —— 开发与基础设施命令
# 禁止 AI 自行修改基础设施目标，除非用户明确要求。
# =============================================================================

.PHONY: help up down ps logs api worker build test lint fmt vet tidy \
        ent-gen migrate-apply migrate-diff

help: ## 显示所有可用目标
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# -----------------------------------------------------------------------------
# 基础设施（Docker）
# -----------------------------------------------------------------------------
up: ## 启动 postgres / redis / minio
	docker compose up -d

down: ## 停止所有服务
	docker compose down

ps: ## 查看服务状态
	docker compose ps

logs: ## 查看服务日志
	docker compose logs -f

# -----------------------------------------------------------------------------
# 开发运行
# -----------------------------------------------------------------------------
api: ## 运行 Admin API 服务
	go run ./cmd/admin

worker: ## 运行 asynq Worker
	go run ./cmd/worker

# -----------------------------------------------------------------------------
# 构建与质量
# -----------------------------------------------------------------------------
build: ## 构建所有二进制
	go build ./...

test: ## 运行测试
	go test ./...

fmt: ## 格式化代码
	go fmt ./...

vet: ## 静态检查
	go vet ./...

lint: vet ## 运行 lint（需 golangci-lint，否则仅 vet）
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, vet passed"

tidy: ## 整理依赖
	go mod tidy

# -----------------------------------------------------------------------------
# ent + Atlas 迁移
# -----------------------------------------------------------------------------
ent-gen: ## 生成 ent 代码
	go generate ./internal/ent

migrate-apply: ## 应用迁移到数据库（dev: ent auto-migrate）
	go run ./cmd/migrate

migrate-diff: ## 生成版本化迁移差异（用法: make migrate-diff name=add_product_table）
	@[ -z "$(name)" ] && echo "usage: make migrate-diff name=<migration_name>" && exit 1 || \
	go run ariga.io/atlas/cmd/atlas migrate diff $(name) --dir "file://migrations" --env local --to "ent://internal/ent/schema"
