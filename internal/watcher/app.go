package watcher

import (
	"github.com/nico612/iam-demo/internal/watcher/config"
	"github.com/nico612/iam-demo/internal/watcher/options"
	"github.com/nico612/iam-demo/pkg/app"
	"github.com/nico612/iam-demo/pkg/log"
)

const commandDesc = `IAM Watcher is a pluggable watcher service used to do some periodic work like cron job. 
But the difference with cron job is iam-watcher also support sleep some duration after previous job done.

Find more iam-pump information at:
    https://github.com/marmotedu/iam/blob/master/docs/guide/en-US/cmd/iam-watcher.md`

// NewApp creates an App object with default parameters.
func NewApp(basename string) *app.App {
	opts := options.NewOptions()
	application := app.NewApp("IAM watcher server",
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

		cfg, err := config.CreateConfigFromOptions(opts)
		if err != nil {
			return err
		}

		return Run(cfg)
	}
}
