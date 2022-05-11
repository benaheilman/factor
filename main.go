package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/benaheilman/factor/worker"
)

func main() {
	milliseconds, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	shift, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	rand.Seed(time.Hour.Nanoseconds())
	worker.Manage(time.Millisecond*time.Duration(milliseconds), int(shift))
}
