package authzserver

import (
	"context"
	"github.com/marmotedu/errors"
	"github.com/nico612/iam-demo/internal/authzserver/config"
	"github.com/nico612/iam-demo/internal/authzserver/load"
	"github.com/nico612/iam-demo/internal/authzserver/load/cache"
	"github.com/nico612/iam-demo/internal/authzserver/store/apiserver"
	"github.com/nico612/iam-demo/pkg/log"
	"github.com/nico612/iam-demo/pkg/storage"

	"github.com/nico612/iam-demo/internal/authzserver/analytics"
	genericoptions "github.com/nico612/iam-demo/internal/pkg/options"
	genericapiserver "github.com/nico612/iam-demo/internal/pkg/server"
	"github.com/nico612/iam-demo/pkg/shutdown"
	"github.com/nico612/iam-demo/pkg/shutdown/shutdownmanagers/posixsignal"
)

// RedisKeyPrefix defines the prefix key in redis for analytics data.
const RedisKeyPrefix = "analytics-"

type authzServer struct {
	gs               *shutdown.GracefulShutdown         // 优雅关闭
	rpcServer        string                             // rpc 服务
	clientCA         string                             // 客户端 CA
	redisOptions     *genericoptions.RedisOptions       // redis
	genericAPIServer *genericapiserver.GenericAPIServer // http/https 服务
	analyticsOptions *analytics.AnalyticsOptions        // 记录分析配置
	redisCancelFunc  context.CancelFunc
}

type preparedAuthzServer struct {
	*authzServer
}

// createAuthzServer(cfg *config.Config) (*authzServer, error) {.
func createAuthzServer(cfg *config.Config) (*authzServer, error) {
	gs := shutdown.New()
	gs.AddShutdownManager(posixsignal.NewPosixSignalManager())

	genericConfig, err := buildGenericConfig(cfg)
	if err != nil {
		return nil, err
	}

	genericServer, err := genericConfig.Complete().New()
	if err != nil {
		return nil, err
	}

	server := &authzServer{
		gs:               gs,
		rpcServer:        cfg.RPCServer,
		clientCA:         cfg.ClientCA,
		redisOptions:     cfg.RedisOptions,
		genericAPIServer: genericServer,
		analyticsOptions: cfg.AnalyticsOptions,
	}

	return server, nil
}

// PrepareRun 初始化相关服务 和 注册路由
func (s *authzServer) PrepareRun() preparedAuthzServer {
	_ = s.initialize()

	installController(s.genericAPIServer.Engine)

	return preparedAuthzServer{s}
}

// Run start to run AuthzServer.
func (s preparedAuthzServer) Run() error {
	// in order to ensure that the reported data is not lost,
	// please ensure the following graceful shutdown sequence
	s.gs.AddShutdownCallback(shutdown.ShutdownFunc(func(string) error {
		s.genericAPIServer.Close()
		if s.analyticsOptions.Enable {
			analytics.GetAnalytics().Stop()
		}
		s.redisCancelFunc() // 取消 redis 连接

		return nil
	}))

	// start shutdown managers
	if err := s.gs.Start(); err != nil {
		log.Fatalf("start shutdown manager failed: %s", err.Error())
	}

	return s.genericAPIServer.Run()
}

func buildGenericConfig(cfg *config.Config) (genericConfig *genericapiserver.Config, lastErr error) {
	genericConfig = genericapiserver.NewConfig()
	if lastErr = cfg.GenericServerRunOptions.ApplyTo(genericConfig); lastErr != nil {
		return
	}

	if lastErr = cfg.FeatureOptions.ApplyTo(genericConfig); lastErr != nil {
		return
	}

	if lastErr = cfg.SecureServing.ApplyTo(genericConfig); lastErr != nil {
		return
	}

	if lastErr = cfg.InsecureServing.ApplyTo(genericConfig); lastErr != nil {
		return
	}

	return
}

// 构建 redis 缓存 config
func (s *authzServer) buildStorageConfig() *storage.Config {
	return &storage.Config{
		Host:                  s.redisOptions.Host,
		Port:                  s.redisOptions.Port,
		Addrs:                 s.redisOptions.Addrs,
		MasterName:            s.redisOptions.MasterName,
		Username:              s.redisOptions.Username,
		Password:              s.redisOptions.Password,
		Database:              s.redisOptions.Database,
		MaxIdle:               s.redisOptions.MaxIdle,
		MaxActive:             s.redisOptions.MaxActive,
		Timeout:               s.redisOptions.Timeout,
		EnableCluster:         s.redisOptions.EnableCluster,
		UseSSL:                s.redisOptions.UseSSL,
		SSLInsecureSkipVerify: s.redisOptions.SSLInsecureSkipVerify,
	}
}

func (s *authzServer) initialize() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.redisCancelFunc = cancel

	// keep redis connected
	go storage.ConnectToRedis(ctx, s.buildStorageConfig())

	// cron to reload all secrets and policies from iam-apiserver
	// 缓存实例
	cacheIns, err := cache.GetCacheInsOr(apiserver.GetAPIServerFactoryOrDie(s.rpcServer, s.clientCA))
	if err != nil {
		return errors.Wrap(err, "get cache instance failed")
	}

	// 初始化 load 并开启 订阅 redis 服务, 当有缓存需要更新时执行更新本地缓存
	load.NewLoader(ctx, cacheIns).Start()

	// 日志上报功能， start analytics service, 并将分析日志写入redis中
	if s.analyticsOptions.Enable {
		analyticsStore := storage.RedisCluster{KeyPrefix: RedisKeyPrefix}

		// 创建数据上报服务
		analyticsIns := analytics.NewAnalytics(s.analyticsOptions, &analyticsStore)
		analyticsIns.Start()
	}

	return nil
}
