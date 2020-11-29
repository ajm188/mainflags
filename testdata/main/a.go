package main

import (
	"flag"
	"fmt"
)

var (
	one   = flag.Int("one", 1, "one")
	two   = flag.Int("two", 2, "two")
	three = flag.Int("three", 3, "three")
)

func main() {
	four := flag.Int("four", 4, "four")

	flag.Parse()

	fmt.Println(*four)
}
