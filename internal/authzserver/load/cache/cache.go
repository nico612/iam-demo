package cache

import (
	"github.com/dgraph-io/ristretto"
	pb "github.com/marmotedu/api/proto/apiserver/v1"
	"github.com/marmotedu/errors"
	"github.com/nico612/iam-demo/internal/authzserver/store"
	"github.com/ory/ladon"
	"sync"
)

// Cache 从 apiserver 中获取 secrets 和 policies 并缓存
type Cache struct {
	lock     *sync.RWMutex
	cli      store.Factory
	secrets  *ristretto.Cache
	policies *ristretto.Cache
}

var (
	// ErrSecretNotFound defines secret not found error.
	ErrSecretNotFound = errors.New("secret not found")
	// ErrPolicyNotFound defines policy not found error.
	ErrPolicyNotFound = errors.New("policy not found")
)

var (
	onceCache sync.Once
	cacheIns  *Cache
)

func GetCacheInsOr(cli store.Factory) (*Cache, error) {
	var err error
	if cli != nil {
		var (
			secretCache *ristretto.Cache
			policyCache *ristretto.Cache
		)
		onceCache.Do(func() {
			c := &ristretto.Config{
				NumCounters: 1e7,     // number of keys to track frequency of (10M).
				MaxCost:     1 << 30, // maximum cost of cache (1GB).
				BufferItems: 64,      // number of keys per Get buffer.
				Cost:        nil,
			}
			secretCache, err = ristretto.NewCache(c)
			if err != nil {
				return
			}
			policyCache, err = ristretto.NewCache(c)
			if err != nil {
				return
			}

			cacheIns = &Cache{
				cli:      cli,
				lock:     new(sync.RWMutex),
				secrets:  secretCache,
				policies: policyCache,
			}
		})
	}

	return cacheIns, err
}

// GetSecret return secret detail for the given key.
func (c *Cache) GetSecret(key string) (*pb.SecretInfo, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	value, ok := c.secrets.Get(key)
	if !ok {
		return nil, ErrSecretNotFound
	}

	return value.(*pb.SecretInfo), nil
}

// GetPolicy return user's ladon policies for the given user.
func (c *Cache) GetPolicy(key string) ([]*ladon.DefaultPolicy, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	value, ok := c.policies.Get(key)
	if !ok {
		return nil, ErrPolicyNotFound
	}

	return value.([]*ladon.DefaultPolicy), nil
}

// Reload secrets and policies.
func (c *Cache) Reload() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	// reload secrets
	secrets, err := c.cli.Secrets().List()
	if err != nil {
		return errors.Wrap(err, "list secrets failed")
	}
	c.secrets.Clear()

	for key, val := range secrets {
		c.secrets.Set(key, val, 1)
	}

	// reload policies
	policies, err := c.cli.Policies().List()
	if err != nil {
		return errors.Wrap(err, "list policies failed")
	}

	c.policies.Clear()
	for key, value := range policies {
		c.policies.Set(key, value, 1)
	}

	return nil

}
