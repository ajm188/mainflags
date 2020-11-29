package flagset

import "flag"

type Config struct {
	A string
	B string
	C string
}

var cfg = &Config{}

func AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.A, "a", "a", "a")
	fs.StringVar(&cfg.B, "b", "b", "b")
	fs.StringVar(&cfg.C, "c", "c", "c")
}
