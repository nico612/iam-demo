package apiserver

import (
	"context"
	"github.com/AlekSi/pointer"
	"github.com/avast/retry-go"
	pb "github.com/marmotedu/api/proto/apiserver/v1"
	"github.com/marmotedu/errors"
	"github.com/nico612/iam-demo/pkg/log"
)

type secrets struct {
	cli pb.CacheClient
}

func newSecrets(ds *datastore) *secrets {
	return &secrets{
		cli: ds.cli,
	}
}

// List returns all the authorization secrets.
func (s *secrets) List() (map[string]*pb.SecretInfo, error) {

	secretInfos := make(map[string]*pb.SecretInfo)
	log.Info("Loading secrets")

	req := &pb.ListSecretsRequest{
		Offset: pointer.ToInt64(0),
		Limit:  pointer.ToInt64(-1),
	}

	var resp *pb.ListSecretsResponse

	// rpc 调用 secrets, 重试3次
	err := retry.Do(func() error {
		var listErr error
		resp, listErr = s.cli.ListSecrets(context.Background(), req)
		if listErr != nil {
			return listErr
		}
		return nil
	}, retry.Attempts(3))

	if err != nil {
		return nil, errors.Wrap(err, "list secrets failed")
	}

	log.Infof("Secrets found (%d total):", len(resp.Items))
	for _, v := range resp.Items {
		log.Infof("- %s:%s", v.Username, v.SecretId)
		secretInfos[v.SecretId] = v
	}

	return secretInfos, nil
}
