package watcher

import (
	"github.com/nico612/iam-demo/internal/apiserver/store/mysql"
	genericoptions "github.com/nico612/iam-demo/internal/pkg/options"
	"github.com/nico612/iam-demo/internal/watcher/config"
	"github.com/nico612/iam-demo/internal/watcher/options"
	"github.com/nico612/iam-demo/pkg/log"
	"github.com/nico612/iam-demo/pkg/shutdown"
	"github.com/nico612/iam-demo/pkg/shutdown/shutdownmanagers/posixsignal"
	"time"
)

type watcherServer struct {
	gs             *shutdown.GracefulShutdown // 优雅关闭服务
	cron           *watchJob
	redisOptions   *genericoptions.RedisOptions
	mysqlOptions   *genericoptions.MySQLOptions
	watcherOptions *options.WatcherOptions
}

// preparedGenericAPIServer is a private wrapper that enforces a call of PrepareRun() before Run can be invoked.
type preparedWatcherServer struct {
	*watcherServer
}

func createWatcherServer(cfg *config.Config) *watcherServer {
	gs := shutdown.New()
	gs.AddShutdownManager(posixsignal.NewPosixSignalManager())

	server := &watcherServer{
		gs:             gs,
		redisOptions:   cfg.RedisOptions,
		mysqlOptions:   cfg.MySQLOptions,
		watcherOptions: cfg.WatcherOptions,
	}

	return server
}

// PrepareRun prepares the server to run, by setting up the server instance.
func (s *watcherServer) PrepareRun() preparedWatcherServer {
	mysqlStore, err := mysql.GetMySQLFactoryOr(s.mysqlOptions)
	if err != nil {
		panic(err)
	}

	s.gs.AddShutdownCallback(shutdown.ShutdownFunc(func(string) error {
		return mysqlStore.Close()
	}))

	// 创建定时job
	s.cron = newWatchJob(s.redisOptions, s.watcherOptions).addWatchers()

	return preparedWatcherServer{s}
}

func (s preparedWatcherServer) Run() error {
	stopCh := make(chan struct{})
	s.gs.AddShutdownCallback(shutdown.ShutdownFunc(func(string) error {
		// wait for running jobs to complete.
		ctx := s.cron.Stop()
		select {
		case <-ctx.Done():
			log.Info("cron jobs stopped.")
		case <-time.After(3 * time.Minute):
			log.Error("context was not done after 3 minutes.")
		}

		stopCh <- struct{}{}

		return nil
	}))

	// start shutdown managers
	if err := s.gs.Start(); err != nil {
		log.Fatalf("start shutdown manager failed: %s", err.Error())
	}

	log.Info("star to run cron jobs.")
	s.cron.Start()

	// blocking here via channel to prevents the process exit.
	<-stopCh

	return nil
}
