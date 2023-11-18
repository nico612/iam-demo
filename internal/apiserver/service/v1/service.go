package v1

import "github.com/nico612/iam-demo/internal/apiserver/store"

type Service interface {
	Users() UserSrv
}

var _ Service = &service{}

type service struct {
	store store.Factory
}

func NewService(store store.Factory) Service {
	return &service{store: store}
}

func (s *service) Users() UserSrv {
	return newUsers(s)
}
