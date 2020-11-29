package flagvars

import "flag"

var (
	one   = flag.Int("one", 1, "one")
	two   = flag.Int("two", 2, "two")
	three = flag.Int("three", 3, "three")
)
