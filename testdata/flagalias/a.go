package flagalias

import ff "flag"

var (
	one   = ff.Int("one", 1, "one")
	two   = ff.Int("two", 2, "two")
	three = ff.Int("three", 3, "three")
)
