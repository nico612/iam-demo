
GO := go

GO_BUILD_FLAGS += -ldflags "$(GO_LDFLAGS)"

ifeq ($(GOOS),windows)
	GO_OUT_EXT := .exe
endif

ifeq ($(ROOT_PACKAGE),)
	$(error the variable ROOT_PACKAGE must be set prior to including golang.mk)
endif

GOPATH := $(shell go env GOPATH)
ifeq ($(origin GOBIN), undefined)
	GOBIN := $(GOPATH)/bin
endif

# 返回$(ROOT_DIR)/cmd/ 下所有不以.md 结尾的文件和目录列表
COMMANDS ?= $(filter-out %.md, $(wildcard $(ROOT_DIR)/cmd/*))
# 获取cmd目录下所有的目录或文件名（去除路径）
BINS ?= $(foreach cmd,${COMMANDS},$(notdir $(cmd)))

ifeq ($(COMMANDS),)
  $(error Could not determine COMMANDS, set ROOT_DIR or run in source dir)
endif
ifeq ($(BINS),)
  $(error Could not determine BINS, set ROOT_DIR or run in source dir)
endif

.PHONY: go.build.verify
## 检查 go 命令行工具是否安装.
go.build.verify:
	@if ! which go &>/dev/null; then echo "Cannot found go compile tool. Please install go tool first."; exit 1; fi

.PHONY: go.build.%
## 编译 Go 源码.
go.build.%:
	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	@echo "===========> Building binary $(COMMAND) $(VERSION) for $(OS) $(ARCH)"
	@mkdir -p $(OUTPUT_DIR)/platforms/$(OS)/$(ARCH)
	@CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(GO_BUILD_FLAGS) -o $(OUTPUT_DIR)/platforms/$(OS)/$(ARCH)/$(COMMAND)$(GO_OUT_EXT) $(ROOT_DIR)/cmd/$(COMMAND)

.PHONY: go.build
# 根据指定的平台编译源码. 该指令依赖 go.build.verify 和 go.build.% 两个指令
go.build: go.build.verify $(addprefix go.build., $(addprefix $(PLATFORM)., $(BINS)))

.PHONY: go.build.multiarch
go.build.multiarch: go.build.verify $(foreach p,$(PLATFORMS),$(addprefix go.build., $(addprefix $(p)., $(BINS))))


.PHONY: go.format
## 格式化 Go 源码.
go.format: tools.verify.goimports
	@$(FIND) -type f -name '*.go' | $(XARGS) gofmt -s -w
	@$(FIND) -type f -name '*.go' | $(XARGS) goimports -w -local $(ROOT_PACKAGE)
	@$(GO) mod edit -fmt

.PHNOY: go.tidy
go.tidy:
	@$(GO) mod tidy

## go.test 单元测试解释
### set -o pipefail; ：这是一个 shell 命令，用于设置管道中的命令失败时终止整个命令的执行。它确保如果管道中的任何一个命令失败，整个命令都会失败。
### $(GO) test 这是运行 Golang 测试的命令。 $(GO)  表示通过环境变量或别名引用的 Go 工具。 test  是 Go 工具的子命令，用于运行测试。
### -race ：这是  go test  命令的选项之一，用于启用数据竞争检测。数据竞争是指多个 goroutine 并发访问共享的数据，可能导致未定义的行为。
### -cover ：这是  go test  命令的选项之一，用于生成代码覆盖率报告。它会在测试运行完成后生成一个覆盖率文件。
### -coverprofile=$(OUTPUT_DIR)/coverage.out ：这是  go test  命令的选项之一，用于指定覆盖率文件的输出路径和名称。 $(OUTPUT_DIR)  是一个占位符，表示输出目录的路径。
### -timeout=10m ：这是  go test  命令的选项之一，用于设置测试的超时时间。在这个例子中，超时时间设置为 10 分钟。
### -shuffle=on ：这是  go test  命令的选项之一，用于打开测试用例的随机顺序执行。这样可以减少测试用例之间的依赖关系。
### -short ：这是  go test  命令的选项之一，用于运行短时间运行的测试。它通常用于跳过一些较慢或资源密集型的测试。
### -v ：这是  go test  命令的选项之一，用于显示每个测试函数的详细信息，包括测试函数的名称和运行结果


.PHONY: go.test
## 执行单元测试.
go.test:
	@echo "===========> Run unit test"
	@mkdir -p $(OUTPUT_DIR)
	@set -o pipefail;$(GO) test -race -cover -coverprofile=$(OUTPUT_DIR)/coverage.out -timeout=10m -shuffle=on -short -v `go list ./...`
	@sed -i '/mock_.*.go/d' $(OUTPUT_DIR)/coverage.out # 从 coverage 中删除mock_.*.go 文件
	@sed -i '/internal\/miniblog\/store\/.*.go/d' $(OUTPUT_DIR)/coverage.out # internal/miniblog/store/ 下的 Go 代码不参与覆盖率计算（这部分测试用例稍后补上）



## 用于在 Go 代码覆盖率报告文件上执行操作的命令。
### go tool cover -func=$(OUTPUT_DIR)/coverage.out ：这是 Go 工具的 cover 子命令，用于分析覆盖率报告文件并生成统计信息。
### -func  选项指定要分析的覆盖率报告文件的路径。
### awk -v target=$(COVERAGE) -f $(ROOT_DIR)/scripts/coverage.awk ：这是一个 awk 命令，用于处理和过滤覆盖率报告的输出。
### -v target=$(COVERAGE)  选项将  $(COVERAGE)  的值传递给 awk 脚本中的  target  变量。
### -f $(ROOT_DIR)/scripts/coverage.awk  选项指定要执行的 awk 脚本文件的路径。

.PHONY: go.cover
## 执行单元测试，并校验覆盖率阈值.
go.cover: go.test
	@$(GO) tool cover -func=$(OUTPUT_DIR)/coverage.out | awk -v target=$(COVERAGE) -f $(ROOT_DIR)/scripts/coverage.awk


.PHONY: go.lint
## 执行静态代码检查.
go.lint: tools.verify.golangci-lint
	@echo "===========> Run golangci to lint source codes"
	@golangci-lint run -c $(ROOT_DIR)/.golangci.yaml $(ROOT_DIR)/...
