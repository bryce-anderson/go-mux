package main

import (
	"fmt"
	"github.com/bryce-anderson/mux"
)

func main() {
	fmt.Printf("Hello, world!\n")

	mux.DecodeFrame(nil)
}
