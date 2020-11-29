package flagalias

import ff "flag"

var (
	one   = ff.Int("one", 1, "one")     // want `ff.Int should not be used on the global flagset`
	two   = ff.Int("two", 2, "two")     // want `ff.Int should not be used on the global flagset`
	three = ff.Int("three", 3, "three") // want `ff.Int should not be used on the global flagset`
)
