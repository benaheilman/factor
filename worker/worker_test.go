package worker

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func BenchmarkFactor(b *testing.B) {
	for _, method := range []string{"naive", "disk"} {
		for _, k := range []int{8, 16, 24, 32} {
			name := fmt.Sprintf("%s-%d-bitshift", method, k)
			b.Run(name, func(b *testing.B) {
				rand.Seed(19597341926366851)
				for i := 0; i < b.N; i++ {
					switch method {
					case "naive":
						_ = factorsNaive(rand.Uint64() >> k)
					case "disk":
						_ = factorsDisk(rand.Uint64()>>k, filepath.Join("..", "primes.bin"))
					}
				}
			})
		}
	}
}

func TestDo(t *testing.T) {
	workers := make(chan struct{}, 1)
	results := make(chan Result)
	var wg sync.WaitGroup
	wg.Add(1)
	workers <- struct{}{}
	go do(workers, &wg, Request{Id: 0, Number: 2}, results, "naive")
	<-results
	wg.Wait()
	assert.True(t, true)
}

func TestGen(t *testing.T) {
	var wg sync.WaitGroup

	c := make(chan Request)
	cxt := context.TODO()
	cxt, cancel := context.WithCancel(cxt)

	wg.Add(1)
	go gen(cxt, &wg, c, 1)
	for i := 0; i < 100; i++ {
		<-c
	}
	cancel()
	wg.Wait()
	_, ok := <-c
	assert.False(t, ok)
	assert.Equal(t, "context canceled", cxt.Err().Error())
}

func TestManage(t *testing.T) {
	Manage(time.Second, 10, "naive")
}
