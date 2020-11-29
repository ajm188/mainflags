package main

import (
	"bytes"
	"sort"
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
		{
			name:       "flagalias",
			returncode: 2,
			problems:   3,
		},
		{
			// the correct usage
			name:       "flagset",
			returncode: 0,
			problems:   0,
		},
		{
			name:       "flagvars",
			returncode: 2,
			problems:   3,
		},
		{
			// in main, so it's okay
			name:       "main",
			returncode: 0,
			problems:   0,
		},
		{
			name:       "noflags",
			returncode: 0,
			problems:   0,
		},
		{
			name:       "scopedflags",
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

func TestImportOrder(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		out  []string
	}{
		{
			name: "stdlib",
			in: []string{
				"math/rand",
				"io",
				"crypto/rand",
			},
			out: []string{
				"crypto/rand",
				"io",
				"math/rand",
			},
		},
		{
			name: "mixed",
			in: []string{
				"strings",
				"github.com/stretchr/testify",
				"gopkg.in/yaml.v3",
				"flag",
				"google.golang.org/grpc",
			},
			out: []string{
				"flag",
				"strings",
				"github.com/stretchr/testify",
				"google.golang.org/grpc",
				"gopkg.in/yaml.v3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pkgs []*packages.Package
			for _, path := range tt.in {
				pkgs = append(pkgs, &packages.Package{
					PkgPath: path,
				})
			}

			sort.Sort(ImportOrder(pkgs))

			var paths []string
			for _, pkg := range pkgs {
				paths = append(paths, pkg.PkgPath)
			}

			assert.Equal(t, tt.out, paths)
		})
	}
}
