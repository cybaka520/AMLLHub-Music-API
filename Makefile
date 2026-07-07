.PHONY: test build clean bench vet

# 构建库
build:
	go build ./...

# 运行单元测试
test:
	go test ./tests/... -v

# 运行性能测试
bench:
	go test -bench=. ./tests/...

# 代码静态检查
vet:
	go vet ./...

# 运行测试并生成覆盖率报告
cover:
	go test ./tests/... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# 清理
clean:
	rm -f coverage.out coverage.html

# 格式化代码
fmt:
	go fmt ./...
