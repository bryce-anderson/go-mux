package main

import (
	"fmt"
	"github.com/bryce-anderson/mux"
	"net"
)

func main() {
	fmt.Printf("Hello, world!\n")

	conn,err := net.Dial("tcp", "localhost:8081")

	if err != nil {
		panic("Failed to connect")
	}

	session := mux.NewClientSession(conn)

	response, err := session.Dispatch([]byte("some data"))

	if err != nil {
		panic("Failed to dispatch")
	}

	fmt.Printf("Received response: ", string(response))

	fmt.Print("Finished.")
}
