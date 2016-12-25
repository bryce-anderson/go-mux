package main

import (
	"fmt"
	"github.com/bryce-anderson/mux"
	"net"
	"sync"
)

func main() {
	fmt.Printf("Hello, world!\n")

	conn,err := net.Dial("tcp", "localhost:8081")

	if err != nil {
		panic("Failed to connect: " + err.Error())
	}

	session := mux.NewClientSession(conn)

	var threads = 1000000
	waitGroup := sync.WaitGroup{}

	waitGroup.Add(threads)

	for i := 0; i < threads; i++ {
		go func() {
			err := do10Dispatches(session)
			if err != nil {
				panic("Error: " + err.Error())
			}
			waitGroup.Done()
		}()
	}

	waitGroup.Wait()


	fmt.Print("Finished.")
}

func do10Dispatches(session mux.ClientSession) error {
	for i := 0; i < 10; i++ {
		data := fmt.Sprintf("Iteration %d: some data", i)
		_, err := session.Dispatch([]byte(data))

		if err != nil {
			return err
		}

		//fmt.Printf("%d: Received response: %s\n", i, string(response))
	}

	return nil
}
