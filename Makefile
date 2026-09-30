BINARY := imap-cli

.PHONY: all build install test fmt vet check clean help

all: build

## build: 编译当前平台的 imap-cli 二进制
build:
	go build -o $(BINARY) .

## install: 安装到 $GOPATH/bin
install:
	go install .

## test: 跑全部测试（集成测试基于内存 IMAP 服务器，无需真实邮箱）
test:
	go test ./... -count=1

## fmt: 格式化全部 Go 代码
fmt:
	gofmt -w .

## vet: 静态检查
vet:
	go vet ./...

## check: fmt 检查 + vet + test，提交前跑这个
check:
	@out=$$(gofmt -l . 2>/dev/null); if [ -n "$$out" ]; then echo "需要 gofmt:"; echo "$$out"; exit 1; fi
	go vet ./...
	go test ./... -count=1

## clean: 删除构建产物
clean:
	rm -f $(BINARY)

help: ## 显示本帮助
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /make /'
