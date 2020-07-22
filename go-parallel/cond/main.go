package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Hello struct {
	str string
}

func (h Hello) N(n int) {
	for i := 0; i < n; i++ {
		fmt.Println("hello")
	}
}

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
	var h Hello
	h.N(2)
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

	// var wg sync.WaitGroup
	done := make(chan interface{})
	defer close(done)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printGreeting(done); err != nil {
			fmt.Println("%v", err)
			return
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printFarewell(done); err != nil {
			fmt.Printf("%v", err)
			return
		}
	}()

	wg.Wait()

	DoWork := func(done <-chan interface{}, pulseInterval time.Duration, nums ...int) (<-chan interface{}, <-chan int) {
		heartbeat := make(chan interface{}, 1)
		intStream := make(chan int)
		go func() {
			defer close(heartbeat)
			defer close(intStream)

			time.Sleep(2 * time.Second)
			pulse := time.Tick(pulseInterval)
		numLoop:
			for _, n := range nums {
				for {
					select {
					case <-done:
						return
					case <-pulse:
						select {
						case heartbeat <- struct{}{}:
						default:
						}
					case intStream <- n:
						continue numLoop
					}
				}
			}
		}()
		return heartbeat, intStream
	}

	done2 := make(chan interface{})
	hb, intStream := DoWork(done2, 1, []int{1, 2, 3, 4, 5}...)
	for {
		select {
		case <-hb:
		case v, ok := <-intStream:
			if !ok {
				fmt.Fprintf(os.Stderr, "not ok")
				return
			}
			fmt.Printf("%v ", v)
		case <-time.After(10 * time.Second):
			log.Fatal("test timed out")
			return
		}
	}

}

func printGreeting(done <-chan interface{}) error {
	greeting, err := genGreeting(done)
	if err != nil {
		return err
	}
	fmt.Printf("%s world\n", greeting)
	return nil
}
func printFarewell(done <-chan interface{}) error {
	greeting, err := genFarewell(done)
	if err != nil {
		return err
	}
	fmt.Printf("%s world\n", greeting)
	return nil
}
func genGreeting(done <-chan interface{}) (string, error) {
	switch locale, err := locale(done); {
	case err != nil:
		return "", nil
	case locale == "EN/US":
		return "hello", nil
	}
	return "", fmt.Errorf("unsupported locale")
}
func genFarewell(done <-chan interface{}) (string, error) {
	switch locale, err := locale(done); {
	case err != nil:
		return "", nil
	case locale == "EN/US":
		return "goodbye", nil
	}
	return "", fmt.Errorf("unsupported locale")
}
func locale(done <-chan interface{}) (string, error) {
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
	return "EN/US", nil
}

func or(channels ...<-chan interface{}) <-chan interface{} {
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}

	orDone := make(chan interface{})
	go func() {
		switch len(channels) {
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
		case <-channels[0]:
		case <-channels[1]:
		case <-channels[2]:
		case <-or(append(channels[3:], orDone)...):
		}
	}()
	return orDone
}

type startGoroutineFn func(
	done <-chan interface{},
	pulseInterval time.Duration,
) (heartbeat <-chan interface{})

func newSteward2(
	timeout time.Duration,
	startGoroutine startGoroutineFn,
) startGoroutineFn {
	return func(
		done <-chan interface{},
		pulseInterval time.Duration,
	) <-chan interface{} {
		heartbeat := make(chan interface{})
		go func() {
			defer close(heartbeat)

			// あとから上書きしたり制御するために親で定義する
			var wardDone chan interface{}
			var wardHeartbeat <-chan interface{}
			startWard := func() {
				wardDone = make(chan interface{})
				wardHeartbeat = startGoroutine(or(wardDone, done), timeout/2)
			}
			startWard()
			pulse := time.Tick(pulseInterval)

		monitorLoop:
			for {
				timeoutSignal := time.After(timeout)
				for {
					select {
					case <-pulse:
						select {
						case heartbeat <- struct{}{}:
						default:
						}
					case <-wardHeartbeat:
						continue monitorLoop
					case <-timeoutSignal:
						log.Println("steward: ward unhealthy; restarting")
						close(wardDone)
						startWard()
						continue monitorLoop
					case <-done:
						return
					}
				}
			}
		}()
		return heartbeat
	}
}
