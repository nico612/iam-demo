package main

import (
	"github.com/nico612/iam-demo/internal/authzserver"
	_ "go.uber.org/automaxprocs"

	"math/rand"
	"time"
)

func main() {

	rand.Seed(time.Now().UTC().UnixNano())

	authzserver.NewApp("iam-authz-server").Run()
}
