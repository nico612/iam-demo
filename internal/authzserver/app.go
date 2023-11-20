package authzserver

import (
	"github.com/nico612/iam-demo/internal/authzserver/config"
	"github.com/nico612/iam-demo/internal/authzserver/options"
	"github.com/nico612/iam-demo/pkg/app"
	"github.com/nico612/iam-demo/pkg/log"
)

const commandDesc = `Authorization server to run ladon policies which can protecting your resources.
It is written inspired by AWS IAM policiis.

Find more iam-authz-server information at:
    https://github.com/marmotedu/iam/blob/master/docs/guide/en-US/cmd/iam-authz-server.md,

Find more ladon information at:
    https://github.com/ory/ladon`

func NewApp(basename string) *app.App {
	opts := options.NewOptions()
	application := app.NewApp(
		"IAM Authorization Server",
		basename,
		app.WithOptions(opts),
		app.WithDescription(commandDesc),
		app.WithDefaultValidArgs(),
		app.WithRunFunc(run(opts)),
	)

	return application
}

func run(opts *options.Options) app.RunFunc {

	return func(basename string) error {

		log.Init(opts.Log)
		defer log.Flush()

		// 从配置文件中创建 服务可用的配置，单独隔离开来，方便扩展
		cfg, err := config.CreateConfigFromOptions(opts)
		if err != nil {
			return err
		}

		return Run(cfg)

	}

}
