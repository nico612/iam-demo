package clean

import (
	"context"
	"github.com/go-redsync/redsync/v4"
	"github.com/nico612/iam-demo/internal/apiserver/store/mysql"
	"github.com/nico612/iam-demo/internal/watcher/options"
	"github.com/nico612/iam-demo/internal/watcher/watcher"
	"github.com/nico612/iam-demo/pkg/log"
)

// 清理 policy_audit 表中超过 maxReserveDays 天后的授权策略
type cleanWatcher struct {
	ctx            context.Context
	mutex          *redsync.Mutex // 分布式锁
	maxReserveDays int
}

func init() {
	watcher.Register("clean", &cleanWatcher{})
}

// Init initializes the watcher for later execution.
func (cw *cleanWatcher) Init(ctx context.Context, rs *redsync.Mutex, config interface{}) error {
	cfg, ok := config.(*options.WatcherOptions)
	if !ok {
		return watcher.ErrConfigUnavailable
	}

	*cw = cleanWatcher{
		ctx:            ctx,
		mutex:          rs,
		maxReserveDays: cfg.Clean.MaxReserveDays,
	}

	return nil
}

func (cw *cleanWatcher) Spec() string {
	return "@every 1d"
}

func (cw *cleanWatcher) Run() {
	if err := cw.mutex.Lock(); err != nil {
		log.L(cw.ctx).Info("cleanWatcher already run.")

		return
	}

	defer func() {
		if _, err := cw.mutex.Unlock(); err != nil {
			log.L(cw.ctx).Errorf("could not release cleanWatcher lock. err: %v", err)

			return
		}
	}()

	db, _ := mysql.GetMySQLFactoryOr(nil)

	rowsAffected, err := db.PolicyAudits().ClearOutdated(cw.ctx, cw.maxReserveDays)
	if err != nil {
		log.L(cw.ctx).Errorw("clean data from policy_audit failed", "error", err)

		return
	}

	log.L(cw.ctx).Debugf("clean data from policy_audit succ, %d rows affected", rowsAffected)
}
