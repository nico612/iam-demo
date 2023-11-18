package store

type Factory interface {
	Users() UserStore
	Secrets() SecretStore
	Policies() PolicyStore
	Close() error
}

var client Factory

func Client() Factory {
	return client
}

func SetClient(factory Factory) {
	client = factory
}
