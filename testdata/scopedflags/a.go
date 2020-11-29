package scopedflags

import "flag"

func x() {
	one := flag.String("one", "one", "one")
}
