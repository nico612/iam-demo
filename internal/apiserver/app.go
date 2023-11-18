package apiserver

import (
	"github.com/nico612/iam-demo/internal/apiserver/config"
	"github.com/nico612/iam-demo/internal/apiserver/options"
	"github.com/nico612/iam-demo/pkg/app"
	"github.com/nico612/iam-demo/pkg/log"
)

const commandDesc = `The IAM API server validates and configures data
for the api objects which include users, policies, secrets, and
others. The API Server services REST operations to do the api objects management.

Find more iam-apiserver information at:
    https://github.com/marmotedu/iam/blob/master/docs/guide/en-US/cmd/iam-apiserver.md`

func NewApp(basename string) *app.App {

	// 构建命令行参数
	opts := options.NewOptions()
	application := app.NewApp(
		"IAM API Server",
		basename,
		app.WithOptions(opts),
		app.WithDescription(commandDesc),
		app.WithDefaultValidArgs(),
		app.WithRunFunc(run(opts)),
	)
	return application
}

// 应用 iam-apiServer 启动函数,里面封装了自定义的启动逻辑
func run(opts *options.Options) app.RunFunc {
	return func(basename string) error {
		log.Init(opts.Log) // 初始化日志
		defer log.Flush()

		// 构建 API Server 应用配置，应用配置和 Options 配置完全独立，二者可能完全不同，只是在该项目中时相同的配置
		cfg, err := config.CreateConfigFromOptions(opts)
		if err != nil {
			return err
		}

		return Run(cfg)
	}
}
