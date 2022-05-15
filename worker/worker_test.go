package worker

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func BenchmarkFactor(b *testing.B) {
	rand.Seed(19597341926366851)
	for k := 16; k < 32; k++ {
		name := fmt.Sprintf("%d-bitshift", k)
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = factors(rand.Uint64() >> k)
			}
		})
	}
}

func TestDo(t *testing.T) {
	workers := make(chan struct{}, 1)
	results := make(chan Result)
	var wg sync.WaitGroup
	wg.Add(1)
	workers <- struct{}{}
	go do(workers, &wg, Request{Id: 0, Number: 2}, results)
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
	Manage(time.Second, 10)
}
