package flagset

import "flag"

type config struct {
	A string
	B string
	C string
}

var cfg = &config{}

func addFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.A, "a", "a", "a")
	fs.StringVar(&cfg.B, "b", "b", "b")
	fs.StringVar(&cfg.C, "c", "c", "c")
}
