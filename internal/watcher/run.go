package watcher

import (
	genericapiserver "github.com/nico612/iam-demo/internal/pkg/server"
	"github.com/nico612/iam-demo/internal/watcher/config"
)

// Run runs the specified pump server. This should never exit.
func Run(cfg *config.Config) error {
	go genericapiserver.ServeHealthCheck(cfg.HealthCheckPath, cfg.HealthCheckAddress)

	return createWatcherServer(cfg).PrepareRun().Run()
}
