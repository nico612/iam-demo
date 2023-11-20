package store

import "github.com/ory/ladon"

type PolicyStore interface {
	List() (map[string][]*ladon.DefaultPolicy, error)
}
