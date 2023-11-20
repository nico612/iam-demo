package store

import pb "github.com/marmotedu/api/proto/apiserver/v1"

type SecretStore interface {
	List() (map[string]*pb.SecretInfo, error)
}
