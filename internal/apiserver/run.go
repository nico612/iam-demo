package apiserver

import (
	"github.com/nico612/iam-demo/internal/apiserver/config"
)

func Run(cfg *config.Config) error {
	// 根据应用配置创建 API Server 服务实例
	server, err := createAPIServer(cfg)
	if err != nil {
		return err
	}

	// 应用初始化 => 启动服务

	return nil
}
