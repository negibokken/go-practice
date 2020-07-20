package main

import (
	"fmt"
	"sync"
	"time"
)

type A struct{ a string }

func testCh() <-chan string {
	result := make(chan string)
	go func() {
		defer close(result)
		for _, s := range []string{"a", "b", "c"} {
			result <- s
		}
	}()
	return result
}

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

	zeroChan := make(chan struct{}, 0)
	go func() {
		close(zeroChan)
	}()
	<-zeroChan

	ch := make(chan int)
	close(ch)

	for i := 0; i < 5; i++ {
		select {
		case <-ch:
			fmt.Println("hello")
		}
	}

	ch = make(chan int, 0)
	select {
	case <-ch:
	case <-time.After(1 * time.Second):
		fmt.Println("Timed out")
	}

	strs := []string{"a", "b", "c"}
	for _, str := range strs {
		fmt.Printf("str: %s\n", str)
	}

	for str := range testCh() {
		fmt.Printf("str: %s\n", str)
	}
}
