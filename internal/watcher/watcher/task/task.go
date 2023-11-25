package task

import (
	"context"
	"github.com/go-redsync/redsync/v4"
	metav1 "github.com/marmotedu/component-base/pkg/meta/v1"
	"github.com/nico612/iam-demo/internal/apiserver/store/mysql"
	"github.com/nico612/iam-demo/internal/watcher/options"
	"github.com/nico612/iam-demo/internal/watcher/watcher"
	"github.com/nico612/iam-demo/pkg/log"
	"time"
)

// 禁用超过 maxInactiveDays 天还没有登录过的用户
type taskWatcher struct {
	ctx             context.Context
	mutex           *redsync.Mutex
	maxInactiveDays int
}

func init() {
	watcher.Register("task", &taskWatcher{})
}

// Init initializes the watcher for later execution.
func (tw *taskWatcher) Init(ctx context.Context, rs *redsync.Mutex, config interface{}) error {
	cfg, ok := config.(*options.WatcherOptions)
	if !ok {
		return watcher.ErrConfigUnavailable
	}

	*tw = taskWatcher{
		ctx:             ctx,
		mutex:           rs,
		maxInactiveDays: cfg.Task.MaxInactiveDays,
	}

	return nil
}

// Spec 指定 job 的间隔时间格式
func (tw *taskWatcher) Spec() string {
	return "@every 1d"
}

// Run runs the watcher job.
func (tw *taskWatcher) Run() {
	if err := tw.mutex.Lock(); err != nil {
		log.L(tw.ctx).Info("taskWatcher already run.")

		return
	}

	defer func() {
		if _, err := tw.mutex.Unlock(); err != nil {
			log.L(tw.ctx).Errorf("could not release taskWatcher lock. err: %v", err)

			return
		}
	}()

	db, _ := mysql.GetMySQLFactoryOr(nil)

	users, err := db.Users().List(tw.ctx, metav1.ListOptions{})
	if err != nil {
		log.L(tw.ctx).Errorf("list user failed", "error", err)

		return
	}

	for _, user := range users.Items {

		// if maxInactiveDays equal to 0, means never forbid
		if tw.maxInactiveDays == 0 {
			continue
		}

		if time.Since(user.LoginedAt) > time.Duration(tw.maxInactiveDays)*(24*time.Hour) {
			log.L(tw.ctx).Infof("user %s not active for %d days, disable his account", user.Name, tw.maxInactiveDays)

			user.Status = 0
			_ = db.Users().Update(tw.ctx, user, metav1.UpdateOptions{})
		}
	}

}
