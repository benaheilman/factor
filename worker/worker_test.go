package worker

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDo(t *testing.T) {
	workers := make(chan bool, 1)
	results := make(chan Result)
	var wg sync.WaitGroup
	wg.Add(1)
	workers <- true
	go do(workers, &wg, Request{Id: 0, Number: 2}, results)
	<-results
	wg.Wait()
	assert.True(t, true)
}

func TestGen(t *testing.T) {
	c := make(chan Request, 10)
	go gen(c, 1, time.Second)
	for i := 0; i < 100; i++ {
		<-c
	}
	assert.True(t, true)
}

func TestManage(t *testing.T) {
	Manage(time.Second, 10)
}
