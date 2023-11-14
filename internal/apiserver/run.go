package apiserver

import (
	"fmt"
	"github.com/nico612/iam-demo/internal/apiserver/config"
)

func Run(cfg *config.Config) error {
	fmt.Printf("config = %v", *cfg)
	return nil
}
