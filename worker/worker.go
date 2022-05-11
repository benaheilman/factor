package worker

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type Request struct {
	Id     uint
	Number uint64
}

type Result struct {
	Id      uint
	Time    time.Duration
	Number  uint64
	Factors []uint64
}

func factors(n uint64) []uint64 {
	for i := uint64(2); i <= uint64(math.Sqrt(float64(n))); i++ {
		if n%i == 0 {
			return append([]uint64{i}, factors(n/i)...)
		}
	}
	return []uint64{n}
}

func do(workers <-chan bool, wg *sync.WaitGroup, r Request, results chan<- Result) {
	start := time.Now()
	factors := factors(r.Number)
	results <- Result{r.Id, time.Since(start), r.Number, factors}
	<-workers
	wg.Done()
}

func gen(c chan<- Request, shift int, timeout time.Duration) {
	i := uint(0)
	t := time.NewTimer(timeout)
loop:
	for {
		r := Request{Id: i, Number: rand.Uint64() >> shift}
		select {
		case c <- r:
			i++
		case <-t.C:
			close(c)
			break loop
		}
	}
}

func print(results <-chan Result) {
	for r := range results {
		fmt.Printf("%05d (%v): %d factors to:\n", r.Id, r.Time, r.Number)
		for _, f := range r.Factors {
			fmt.Printf("\t%d\n", f)
		}
	}
}

func Manage(timeout time.Duration, shift int) {
	var wg sync.WaitGroup

	size := runtime.GOMAXPROCS(0)
	requests := make(chan Request)
	results := make(chan Result)
	workers := make(chan bool, size)

	go gen(requests, shift, timeout)
	go print(results)

	for r := range requests {
		// Block until slots become available
		workers <- true
		wg.Add(1)
		go do(workers, &wg, r, results)
	}
	wg.Wait()
	close(results)
}
