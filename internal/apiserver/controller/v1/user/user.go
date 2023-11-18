package user

import (
	srvv1 "github.com/nico612/iam-demo/internal/apiserver/service/v1"
	"github.com/nico612/iam-demo/internal/apiserver/store"
)

type UserController struct {
	srv srvv1.Service
}

func NewUserController(store store.Factory) *UserController {
	return &UserController{srv: srvv1.NewService(store)}
}
