package mainflags

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestMainflags(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer(), "./...")
}
