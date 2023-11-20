package store

type Factory interface {
	Secrets() SecretStore
	Policies() PolicyStore
}

var client Factory

func Client() Factory {
	return client
}

func SetClient(factory Factory) {
	client = factory
}
