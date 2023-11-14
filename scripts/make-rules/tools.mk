TOOLS ?=$(BLOCKER_TOOLS) $(CRITICAL_TOOLS) $(TRIVIAL_TOOLS)

.PHONY: tools.install
# 安装所有工具
tools.install: $(addprefix tools.install., $(TOOLS))

.PHONY: tools.install.%
tools.install.%:
	@echo "==============> Installing $*"
	@$(MAKE) install.$*

.PHONY: tools.verify.%
# 验证，如果没有安装则安装相关工具
tools.verify.%:
	@if ! which $* &>/dev/null; then $(MAKE) tools.install.$*; fi


.PHONY: install.swagger
# 生成 Swagger 格式的 API 文档
install.swagger:
	@$(GO) install github.com/go-swagger/go-swagger/cmd/swagger@latest

.PHONY: install.golangci-lint
install.golangci-lint:
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2
	@golangci-lint completion bash > $(HOME)/.golangci-lint.bash
	@if ! grep -q .golangci-lint.bash $(HOME)/.bashrc; then echo "source \$$HOME/.golangci-lint.bash" >> $(HOME)/.bashrc; fi

.PHONY: install.go-junit-report
# 将 Go 测试的输出转化为 junit.xml文件
install.go-junit-report:
	@$(GO) install github.com/jstemmer/go-junit-report@latest

.PHONY: install.gsemver
# 根据 git commit message 命令规范自动生成语义化版本
install.gsemver:
	@$(GO) install github.com/arnaud-deprez/gsemver@latest

.PHONY: install.git-chglog
# 根据 git commit 命令自动生成 CHANGELOG 日志
install.git-chglog:
	@$(GO) install github.com/git-chglog/git-chglog/cmd/git-chglog@latest

.PHONY: install.github-release
# 命令行工具，用来创建、修改 Github 版本
install.github-release:
	@$(GO) install github.com/github-release/github-release@latest

.PHONY: install.coscli
install.coscli:
	@wget -q https://github.com/tencentyun/coscli/releases/download/v0.10.2-beta/coscli-linux -O ${HOME}/bin/coscli
	@chmod +x ${HOME}/bin/coscli

.PHONY: install.coscmd
install.coscmd:
	@if which pip &>/dev/null; then pip install coscmd; else pip3 install coscmd; fi

.PHONY: install.golines
# 格式化Go代码中的长行为短行
install.golines:
	@$(GO) install github.com/segmentio/golines@latest

.PHONY: install.go-mod-outdated
# 检查依赖包是否有更新
install.go-mod-outdated:
	@$(GO) install github.com/psampaz/go-mod-outdated@latest

.PHONY: install.mockgen
# 接口 Mock 工具
install.mockgen:
	@$(GO) install github.com/golang/mock/mockgen@latest

.PHONY: install.gotests
# 根据 Go 代码自动生成单元测试模块
install.gotests:
	@$(GO) install github.com/cweill/gotests/gotests@latest

.PHONY: install.protoc-gen-go
# 生成 pb.go 文件
install.protoc-gen-go:
	@$(GO) install github.com/golang/protobuf/protoc-gen-go@latest

.PHONY: install.cfssl
# Cloudflare 的 PKI 和 TLS 工具集
install.cfssl:
	@$(ROOT_DIR)/scripts/install/install.sh iam::install::install_cfssl

.PHONY: install.addlicense
# 通过扫描自定的文件，来确保源码文件有版权头
install.addlicense:
	@$(GO) install github.com/marmotedu/addlicense@latest

.PHONY: install.goimports
# 自动格式话Go代码并对所有引入的包进行管理，包括自动增删依赖的包，将依赖包按字母排序并分类
install.goimports:
	@$(GO) install golang.org/x/tools/cmd/goimports@latest

.PHONY: install.depth
# 通过分析导入的库，将某个包的依赖关系用树状结构显示出来
install.depth:
	@$(GO) install github.com/KyleBanks/depth/cmd/depth@latest

.PHONY: install.go-callvis
# 可视化显示 Go 调用关系
install.go-callvis:
	@$(GO) install github.com/ofabry/go-callvis@latest

.PHONY: install.gothanks
# 自动在 Github 上 Star 项目的依赖包所在的 Github 资源库
install.gothanks:
	@$(GO) install github.com/psampaz/gothanks@latest

.PHONY: install.richgo
# 用文本装饰丰富 Go 测试输出
install.richgo:
	@$(GO) install github.com/kyoh86/richgo@latest

.PHONY: install.rts
# 用于根据服务器的响应生成 Go 结构体
install.rts:
	@$(GO) install github.com/galeone/rts/cmd/rts@latest

.PHONY: install.codegen
install.codegen:
	@$(GO) install ${ROOT_DIR}/tools/codegen/codegen.go

.PHONY: install.kube-score
install.kube-score:
	@$(GO) install github.com/zegl/kube-score/cmd/kube-score@latest

.PHONY: install.go-gitlint
# 静态代码检查工具
install.go-gitlint:
	@$(GO) install github.com/marmotedu/go-gitlint/cmd/go-gitlint@latest