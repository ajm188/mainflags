package flagvars

import "flag"

var (
	one   = flag.Int("one", 1, "one")     // want `flag.Int should not be used on the global flagset`
	two   = flag.Int("two", 2, "two")     // want `flag.Int should not be used on the global flagset`
	three = flag.Int("three", 3, "three") // want `flag.Int should not be used on the global flagset`
)
