package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name       string
		returncode int
		problems   int
	}{
		{
			name:       "dotimport",
			returncode: 2,
			problems:   1,
		},
	}

	var buf bytes.Buffer

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer buf.Reset()

			cfg := &packages.Config{
				Mode: packageMode,
			}
			pkgs, err := packages.Load(cfg, "./testdata/"+tt.name)
			require.NoError(t, err)

			returncode := run(&buf, pkgs)
			assert.Equal(t, tt.returncode, returncode, "run() had unexpected returncode")

			out := buf.String()
			problems := strings.Count(out, "should not be used on the global flagset")
			problems += strings.Count(out, "package flag should not be dot-imported")

			assert.Equal(t, tt.problems, problems, "run() produced the wrong number of problems")
		})
	}
}
