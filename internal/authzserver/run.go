package authzserver

import "github.com/nico612/iam-demo/internal/authzserver/config"

func Run(cfg *config.Config) error {
	server, err := createAuthzServer(cfg)
	if err != nil {
		return nil
	}

	return server.PrepareRun().Run()
}
