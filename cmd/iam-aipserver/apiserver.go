package main

import (
	"fmt"
	_ "go.uber.org/automaxprocs"
	"math/rand"

	"time"
)

func main() {
	// 设置随机数生成器的种子，以确保每次程序运行时生成的随机数序列都是不同的。
	rand.Seed(time.Now().UTC().UnixNano())
	fmt.Println("apiserver main")
}
