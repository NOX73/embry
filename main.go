package main

import (
	"embry/client"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"time"
)

const (
	MaxExecutionTime = time.Minute
	ExpireKeyTime    = 10 * time.Second
	WorkersCount     = 10
	MaxNum           = 100010
	MinNum           = 100000
)

type Result struct {
	Num int `json:"num"`
	// do not send all data to memcache, it's so big
	Value *big.Int `json:"-"`
}

func fact(n int) *big.Int {
	var f = big.NewInt(int64(1))

	if n == 0 {
		return f
	}

	for i := 2; i <= n; i++ {
		f.Mul(f, big.NewInt(int64(i)))
	}

	return f
}

func worker(n int) {

	var servers = []string{
		strings.Trim(os.Getenv("MEMCACHED1_1_PORT_11211_TCP"), "tcp://"),
		strings.Trim(os.Getenv("MEMCACHED2_1_PORT_11211_TCP"), "tcp://"),
		strings.Trim(os.Getenv("MEMCACHED3_1_PORT_11211_TCP"), "tcp://"),
	}

	var client = client.NewClient(servers...)

	for {

		i := rand.Intn(MaxNum-MinNum) + MinNum

		var key = fmt.Sprintf("%v", i)

		var result = &Result{}

		log.Printf("Worker %v start geting fact(%v) ... ", n, i)
		var err = client.Get(key, result, MaxExecutionTime, ExpireKeyTime, func() {
			log.Printf("Worker %v calulating fact(%v) ... ", n, i)
			result.Num = i
			result.Value = fact(i)
		})

		if err != nil {
			log.Printf("Worker %v got error: %v", n, err)
		} else {
			log.Printf("Worker %v got fact(%v). Number of digist = %v.", n, i, len(result.Value.String()))
		}

	}
}

func main() {
	for i := 1; i <= WorkersCount; i++ {
		go worker(i)
	}

	log.Println("Started.")
	select {}
}
