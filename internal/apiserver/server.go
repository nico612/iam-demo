package apiserver

import (
	"context"
	"fmt"
	pb "github.com/marmotedu/api/proto/apiserver/v1"
	"github.com/nico612/iam-demo/internal/apiserver/config"
	"github.com/nico612/iam-demo/internal/apiserver/store"
	"github.com/nico612/iam-demo/internal/apiserver/store/mysql"
	genericoptions "github.com/nico612/iam-demo/internal/pkg/options"
	"github.com/nico612/iam-demo/pkg/log"
	"github.com/nico612/iam-demo/pkg/shutdown"
	"github.com/nico612/iam-demo/pkg/shutdown/shutdownmanagers/posixsignal"
	"github.com/nico612/iam-demo/pkg/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	cachev1 "github.com/nico612/iam-demo/internal/apiserver/controller/v1/cache"
	genericapiserver "github.com/nico612/iam-demo/internal/pkg/server"
)

// 构建服务

type apiServer struct {
	gs               *shutdown.GracefulShutdown // 优雅关闭服务
	redisOptions     *genericoptions.RedisOptions
	gRPCAPIServer    *grpcAPIServer
	genericapiserver *genericapiserver.GenericAPIServer
}

type preparedAPIServer struct {
	*apiServer
}

// ExtraConfig defines extra configuration for the iam-apiserver.
type ExtraConfig struct {
	Addr         string
	MaxMsgSize   int
	ServerCert   genericoptions.GeneratableKeyCert
	mysqlOptions *genericoptions.MySQLOptions
	// etcdOptions      *genericoptions.EtcdOptions
}

type completedExtraConfig struct {
	*ExtraConfig
}

// Complete fills in any fields not set that are required to have valid data and can be derived from other fields.
func (c *ExtraConfig) complete() *completedExtraConfig {
	if c.Addr == "" {
		c.Addr = "127.0.0.1:8081"
	}

	return &completedExtraConfig{c}
}

func createAPIServer(cfg *config.Config) (*apiServer, error) {
	gs := shutdown.New()                                       // 优雅关闭
	gs.AddShutdownManager(posixsignal.NewPosixSignalManager()) // 添加关闭信号

	genericConfig, err := buildGenericConfig(cfg)
	if err != nil {
		return nil, err
	}

	// grpc 配置
	extraConfig, err := buildExtraConfig(cfg)
	if err != nil {
		return nil, nil
	}

	// 应用补全，然后创建服务
	genericServer, err := genericConfig.Complete().New()
	if err != nil {
		return nil, err
	}

	// grpc 服务
	extraServer, err := extraConfig.complete().New()
	if err != nil {
		return nil, err
	}

	server := &apiServer{
		gs:               gs,
		redisOptions:     cfg.RedisOptions,
		genericapiserver: genericServer, // HTTP HTTPS 服务
		gRPCAPIServer:    extraServer,   // gRPC 服务
	}

	return server, nil

}

// PrepareRun 执行apiServer初始化
func (s *apiServer) PrepareRun() preparedAPIServer {

	s.initRedisStore()

	s.gs.AddShutdownCallback(shutdown.ShutdownFunc(func(string) error {
		mysqlStore, _ := mysql.GetMySQLFactoryOr(nil)
		if mysqlStore != nil {
			_ = mysqlStore.Close()
		}

		s.gRPCAPIServer.Close()
		s.genericapiserver.Close()

		return nil
	}))

	return preparedAPIServer{s}

}

func (s preparedAPIServer) Run() error {

	go s.gRPCAPIServer.Run()

	if err := s.gs.Start(); err != nil {
		log.Fatalf("start shutdown manager failed: %s", err.Error())
	}

	return s.genericapiserver.Run()
}

// New create a grpcAPIServer instance.
func (c *completedExtraConfig) New() (*grpcAPIServer, error) {
	creds, err := credentials.NewServerTLSFromFile(c.ServerCert.CertKey.CertFile, c.ServerCert.CertKey.KeyFile)
	if err != nil {
		log.Fatalf("Failed to generate credentials %s", err.Error())
	}
	opts := []grpc.ServerOption{grpc.MaxRecvMsgSize(c.MaxMsgSize), grpc.Creds(creds)}
	grpcServer := grpc.NewServer(opts...)

	storeIns, _ := mysql.GetMySQLFactoryOr(c.mysqlOptions)
	// storeIns, _ := etcd.GetEtcdFactoryOr(c.etcdOptions, nil)
	store.SetClient(storeIns)

	cacheIns, err := cachev1.GetCacheInsOr(storeIns)
	if err != nil {
		log.Fatalf("Failed to get cache instance: %s", err.Error())
	}

	pb.RegisterCacheServer(grpcServer, cacheIns)

	reflection.Register(grpcServer)

	return &grpcAPIServer{grpcServer, c.Addr}, nil
}

// 构建扩展配置
func buildExtraConfig(cfg *config.Config) (*ExtraConfig, error) {
	return &ExtraConfig{
		Addr:         fmt.Sprintf("%s:%d", cfg.GRPCOptions.BindAddress, cfg.GRPCOptions.BindPort),
		MaxMsgSize:   cfg.GRPCOptions.MaxMsgSize,
		ServerCert:   cfg.SecureServing.ServerCert,
		mysqlOptions: cfg.MySQLOptions,
	}, nil
}

// 根据应用配置来构建服务配置
func buildGenericConfig(cfg *config.Config) (genericConfig *genericapiserver.Config, lastErr error) {
	genericConfig = genericapiserver.NewConfig()

	// 根据应用配置来构建 HTTP/gRPC 服务配置
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

func (s *apiServer) initRedisStore() {
	ctx, cancle := context.WithCancel(context.Background())
	s.gs.AddShutdownCallback(shutdown.ShutdownFunc(func(string) error {
		cancle()
		return nil
	}))

	config := &storage.Config{
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

	go storage.ConnectToRedis(ctx, config)
}
