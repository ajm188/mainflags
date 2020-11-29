package scopedflags

import (
	"flag"
	"fmt"
)

func x() {
	one := flag.String("one", "one", "one") // want `flag.String should not be used on the global flagset`
	flag.Parse()
	fmt.Println(*one)
}
