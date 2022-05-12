package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
