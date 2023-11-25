package watcher

import (
	"context"
	"fmt"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	genericoptions "github.com/nico612/iam-demo/internal/pkg/options"
	"github.com/nico612/iam-demo/internal/watcher/options"
	"github.com/nico612/iam-demo/internal/watcher/watcher"
	"github.com/nico612/iam-demo/pkg/log"
	"github.com/nico612/iam-demo/pkg/log/cronlog"
	"github.com/robfig/cron/v3"
	"time"
)

type watchJob struct {
	*cron.Cron
	config *options.WatcherOptions
	rs     *redsync.Redsync // 用于获取分布式锁
}

func newWatchJob(redisOptions *genericoptions.RedisOptions, watcherOptions *options.WatcherOptions) *watchJob {
	logger := cronlog.NewLogger(log.SugaredLogger())

	client := goredislib.NewClient(&goredislib.Options{
		Addr:     fmt.Sprintf("%s:%d", redisOptions.Host, redisOptions.Port),
		Username: redisOptions.Username,
		Password: redisOptions.Password,
	})

	rs := redsync.New(goredis.NewPool(client)) //初始化redsync 实例，用户获取分布式锁

	// 创建 cron 实例
	cronjob := cron.New(
		cron.WithSeconds(), // 每秒执行一次任务
		cron.WithChain(cron.SkipIfStillRunning(logger), cron.Recover(logger)), // 拦截器
	)

	// 创建作业系统
	return &watchJob{
		Cron:   cronjob,
		config: watcherOptions,
		rs:     rs,
	}
}

func (w *watchJob) addWatchers() *watchJob {

	for name, watch := range watcher.ListWatchers() {
		// log with `{"watcher": "counter"}` key-value to distinguish which watcher the log comes from.
		//nolint: golint,staticcheck
		ctx := context.WithValue(context.Background(), log.KeyWatcherName, name)

		if err := watch.Init(ctx, w.rs.NewMutex(name, redsync.WithExpiry(2*time.Hour)), w.config); err != nil {
			log.Panicf("construct watcher %s failed: %s", name, err.Error())
		}

		_, _ = w.AddJob(watch.Spec(), watch)
	}

	return w
}
