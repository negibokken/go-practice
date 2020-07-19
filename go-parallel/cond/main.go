package main

import (
	"fmt"
	"sync"
)

type A struct{ a string }

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("hello-inside")
	}()
	fmt.Println("hello")
	wg.Wait()
	fmt.Println("hello2")

	myPool := &sync.Pool{
		New: func() interface{} {
			fmt.Println("Creating new instance")
			return struct{}{}
		},
	}

	myPool.Get()
	instance := myPool.Get()
	myPool.Put(instance)

	myPool.Get()
}
