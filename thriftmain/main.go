package main

import (
	"fmt"
	"github.com/bryce-anderson/mux"
	"github.com/bryce-anderson/mux/thriftmux"
	"github.com/bryce-anderson/mux/thriftmux/tutorial"
	"math"
	"git.apache.org/thrift.git/lib/go/thrift"
	"sync"
)

func main() {
	fmt.Println("Hello, world!")

	group := sync.WaitGroup{}
	sessions := 100
	threads := 100
	dispatches := 1000000

	for i := 0; i < sessions; i++ {
		spawnSession(&group, threads, dispatches)
	}
	group.Wait()
}

func spawnSession(group *sync.WaitGroup, threads int, dispatches int) {
	sess, err := mux.NewClientSession("localhost:8081", math.MaxInt32)
	if err != nil {
		panic("Failed to establish connection: " + err.Error())
	}

	for i := 0; i < threads; i++ {
		group.Add(1)

		go func(i int) {
			trans := thriftmux.NewThriftMuxTransport(sess)
			factory := thrift.NewTBinaryProtocolFactoryDefault()

			client := tutorial.NewMultiplicationServiceClientFactory(trans, factory)

			for j := tutorial.Int(0); j < tutorial.Int(dispatches); j++ {
				r, err := client.Multiply(j, j)
				if err != nil {
					panic("Failed to perform multiplication: " + err.Error())
				}

				if (r != j*j) {
					panic(fmt.Sprintf("Invalid squire: %d * %d /= %d", j, j, r))
				}

				//fmt.Printf("Multiplication %d * %d = %d\n", i, i, r)
			}

			fmt.Printf("Group %d done.\n", i)

			group.Done()
		}(i)

	}
}
