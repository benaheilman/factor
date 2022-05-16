package worker

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"
)

// Message to factor a number into primes
type Request struct {
	Id     uint
	Number uint64
}

// Message sent after factoring a number
type Result struct {
	Id      uint
	Time    time.Duration
	Number  uint64
	Factors []uint64
}

// Context used to read/cache primes stored on disk
type cacheReader struct {
	File        *os.File
	ObjectSize  int
	BlockSize   int
	Cache       []byte
	CacheOffset int
}

func newPrimeReader(path string, blockSize int) io.Reader {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	cache := make([]byte, blockSize)
	n, err := file.Read(cache)
	if err != nil {
		log.Fatal(err)
	}
	if n != blockSize {
		log.Fatalf("Only got %d bytes at head of primes file, expected %d", n, blockSize)
	}
	return &cacheReader{File: file, ObjectSize: 8, BlockSize: blockSize, Cache: cache, CacheOffset: 0}
}

func (pr *cacheReader) Read(p []byte) (n int, err error) {
	if pr.CacheOffset >= pr.BlockSize {
		n, err := pr.File.Read(pr.Cache)
		switch err {
		case io.EOF:
			pr.File.Close()
			return 0, io.EOF
		case nil:
			break
		default:
			return n, err
		}
		if n != pr.BlockSize {
			pr.BlockSize = n
		}
		pr.CacheOffset = 0
	}
	n = copy(p, pr.Cache[pr.CacheOffset:])
	pr.CacheOffset = pr.CacheOffset + n*pr.ObjectSize
	return n, nil
}

func factorsNaive(n uint64) []uint64 {
	sqrt := uint64(math.Sqrt(float64(n)))
	for i := uint64(2); i <= sqrt; i++ {
		if n%i == 0 {
			return append([]uint64{i}, factorsNaive(n/i)...)
		}
	}
	return []uint64{n}
}

func factorsDisk(n uint64, path string) []uint64 {
	sqrt := uint64(math.Sqrt(float64(n)))

	var h uint64 = 0
	for i := uint64(2); i < 256*256 && i <= sqrt; i++ {
		if n%i == 0 {
			return append([]uint64{i}, factorsDisk(n/i, path)...)
		}
		h = i
	}
	if h+1 > sqrt {
		return []uint64{n}
	}

	r := newPrimeReader(path, 1024*1024)
	buf := make([]byte, 8)

	var prime uint64 = 0
	for {
		i, err := r.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if i != 8 {
			log.Fatalf("got %d bytes instead of the expected 8", i)
		}
		prime, i = binary.Uvarint(buf)
		if i <= 0 {
			log.Fatalf("decoded %d bytes instead of the expected 8", i)
		}
		if prime > sqrt {
			return []uint64{n}
		}
		if n%prime == 0 {
			return append([]uint64{prime}, factorsDisk(n/prime, path)...)
		}
	}
	for i := prime; i <= sqrt; i++ {
		if n%i == 0 {
			return append([]uint64{i}, factorsDisk(n/i, path)...)
		}
	}
	return []uint64{n}
}

func isPrime(n uint64) bool {
	sqrt := uint64(math.Sqrt(float64(n)))
	for i := uint64(2); i <= sqrt; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func primes(wg *sync.WaitGroup, ch chan<- uint64, until uint64) {
	defer wg.Done()
	defer close(ch)

	for i := uint64(2); i < until; i++ {
		if i%(1<<16) == 0 {
			log.Println(i)
		}
		if isPrime(i) {
			ch <- i
		}
	}
}

func record(wg *sync.WaitGroup, ch <-chan uint64, path string) {
	defer wg.Done()

	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for {
		next, ok := <-ch
		if !ok {
			return
		}
		buf := make([]byte, 8)
		binary.PutUvarint(buf, next)
		nn, err := w.Write(buf)
		if err != nil {
			log.Fatal(err)
		}
		if nn != 8 {
			log.Fatalf("wrote %d bytes instead of 8", nn)
		}
	}
}

func Generate(path string, limit int) {
	wg := sync.WaitGroup{}
	ch := make(chan uint64)
	wg.Add(1)
	go record(&wg, ch, path)
	wg.Add(1)
	go primes(&wg, ch, 1<<limit)
	wg.Wait()
}

func do(workers <-chan struct{}, wg *sync.WaitGroup, r Request, results chan<- Result, method string) {
	defer wg.Done()

	var factors []uint64
	start := time.Now()
	switch method {
	case "naive":
		factors = factorsNaive(r.Number)
	case "disk":
		factors = factorsDisk(r.Number, "primes.bin")
	default:
		log.Fatalf("unknown method: %s", method)
	}
	results <- Result{r.Id, time.Since(start), r.Number, factors}
	<-workers
}

func gen(cxt context.Context, wg *sync.WaitGroup, c chan<- Request, shift int) {
	defer wg.Done()
	defer close(c)

	i := uint(0)
	for {
		r := Request{Id: i, Number: rand.Uint64() >> shift}
		select {
		case c <- r:
			i++
		case <-cxt.Done():
			return
		}
	}
}

func print(results <-chan Result) {
	for r := range results {
		fmt.Printf("%05d (%v): %d factors to primes:\n", r.Id, r.Time, r.Number)
		for _, f := range r.Factors {
			fmt.Printf("\t%d\n", f)
		}
	}
}

func Manage(timeout time.Duration, shift int, method string) {
	var wg sync.WaitGroup

	size := runtime.GOMAXPROCS(0)
	requests := make(chan Request)
	results := make(chan Result)
	defer close(results)
	workers := make(chan struct{}, size)
	defer close(workers)

	cxt, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	wg.Add(1)
	go gen(cxt, &wg, requests, shift)
	go print(results)

	for r := range requests {
		// Block until slots become available
		workers <- struct{}{}
		wg.Add(1)
		go do(workers, &wg, r, results, method)
	}
	wg.Wait()
}
