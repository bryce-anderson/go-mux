package main

import (
	"fmt"
	"github.com/bryce-anderson/mux"
	"time"
)

func main() {
	fmt.Printf("Hello, world!\n")

	a := []int{1,2,3}

	fmt.Print("Data: ", a)

	b := a

	b[0] = 255

	fmt.Print("Data: ", a)

	time.Sleep(10000)

	fmt.Print("Finished.")
}
