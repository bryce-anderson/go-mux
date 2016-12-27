package main

import (
	"fmt"
	"github.com/bryce-anderson/mux"
	"sync"
	"math"
	"log"
)

func main() {
	fmt.Printf("Hello, world!\n")

	var sessions = 100
	var threads = 1000
	var iterations = 1000000
	waitGroup := sync.WaitGroup{}

	for s := 0; s < sessions; s++ {
		session, err := mux.NewClientSession("localhost:8081", math.MaxInt32)
		if err != nil {
			fmt.Printf("Session %d failed\n", s)
		} else {
			for i := 0; i < threads; i++ {
				waitGroup.Add(1)
				go func() {
					err := doDispatches(iterations, session)
					if err != nil {
						log.Printf("Thread %d of session %d failed. Error: %s\n", i, s, err.Error())
					}
					waitGroup.Done()
				}()
			}

		}
	}
	waitGroup.Wait()


	fmt.Print("Finished.")
}

func doDispatches(iterations int, session mux.ClientSession) error {
	for i := 0; i < iterations; i++ {
		data := fmt.Sprintf("Iteration %d: some data", i)
		_, err := mux.SimpleDispatch(session, []byte(data))

		if err != nil {
			return err
		}

		//fmt.Printf("%d: Received response: %s\n", i, string(response))
	}

	return nil
}
