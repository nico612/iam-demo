package main

import (
	"math/rand"

	"github.com/nico612/iam-demo/internal/apiserver"
	_ "go.uber.org/automaxprocs"

	"time"
)

func main() {
	// 设置随机数生成器的种子，以确保每次程序运行时生成的随机数序列都是不同的。
	rand.Seed(time.Now().UTC().UnixNano())
	apiserver.NewApp("apiserver").Run()
}
