package main

import (
	"github.com/ajm188/mainflags"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(mainflags.Analyzer())
}
