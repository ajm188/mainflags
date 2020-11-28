package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"
)

const (
	packageMode = (packages.NeedName |
		packages.NeedFiles |
		packages.NeedCompiledGoFiles |
		packages.NeedImports |
		packages.NeedSyntax |
		packages.NeedTypes)
	logLevelHelp = `verbosity of logging. Should be LogError <= x < LogWarning for use in golangci-lint.
Set to a negative number to disable logging entirely`
)

var (
	logLevel             = (*LogLevel)(flag.Int("verbosity", 0, logLevelHelp))
	cwd                  = mustString(os.Getwd())
	allowedFlagFunctions = map[string]bool{
		"Arg":           true,
		"Args":          true,
		"NArg":          true,
		"NArgs":         true,
		"Parse":         true,
		"Parsed":        true,
		"PrintDefaults": true,
		"Set":           true,
		"UnquoteUsage":  true,
		"Visit":         true,
		"VisitAll":      true,
		"Lookup":        true,
		"NewFlagSet":    true,
	}
)

func mustString(s string, err error) string {
	if err != nil {
		fatalf("cannot get current directory, err = %s", err)
	}

	return s
}

// ImportOrder sorts a list of packages by standard goimports order. Standard
// library packages come first alphabetically, then third-party packages.
//
// We currently don't attempt to detect what gomodule mainflags is analyzing,
// so module packages and third-party packages may be interspersed.
type ImportOrder []*packages.Package

func (a ImportOrder) Len() int      { return len(a) }
func (a ImportOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ImportOrder) Less(i, j int) bool {
	l, r := a[i].PkgPath, a[j].PkgPath

	lparts := strings.Split(l, "/")
	rparts := strings.Split(r, "/")

	if !strings.Contains(lparts[0], ".") {
		if !strings.Contains(rparts[0], ".") {
			return l < r
		}

		return true
	}

	if !strings.Contains(rparts[0], ".") {
		return false
	}

	return l < r
}

func main() {
	flag.Parse()

	cfg := &packages.Config{
		Mode: packageMode,
	}

	var pkgNames []string

	switch flag.NArg() {
	case 0:
		pkgNames = append(pkgNames, ".")
	default:
		pkgNames = append(pkgNames, flag.Args()...)
	}

	debugf("begin program load")

	pkgs, err := packages.Load(cfg, pkgNames...)
	if err != nil {
		fatalf("cannot load program; err = %s", err)
	}

	debugf("end program load")

	sort.Sort(ImportOrder(pkgs))
	// nolint:prealloc
	var problems []*Problem

	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			debugf("%s is package main, skipping", pkg.PkgPath)

			continue
		}

		problems = append(problems, doPackage(pkg)...)
	}

	if len(problems) == 0 {
		infof("no problems found")
		return
	}

	for _, problem := range problems {
		errorf(formatProblem(problem))
	}

	os.Exit(2)
}

// formatProblem returns a string representation of a problem suitable for
// consumption by golangci-lint.
func formatProblem(problem *Problem) string {
	buf := strings.Builder{}

	pos := problem.pkg.Fset.Position(problem.pos)

	f, err := filepath.Rel(cwd, pos.Filename)
	if err == nil {
		pos.Filename = f
	}

	buf.WriteString(pos.String())
	buf.WriteString(": ")
	buf.WriteString(problem.message)

	return buf.String()
}

// Problem represents a linter problem. It is formatted for consumption by
// golangci-lint in formatProblem.
type Problem struct {
	pkg     *packages.Package
	file    *ast.File
	pos     token.Pos
	message string
}

// doPackage processes one Go package, returning a list of Problems with flag
// usage in that package.
func doPackage(pkg *packages.Package) []*Problem {
	var ( // nolint:prealloc
		importLog sync.Once
		problems  []*Problem
	)

	for _, file := range pkg.Syntax {
		debugf("processing %s\n", file.Name.Name)
		importsFlag, alias, p := processImports(pkg, file)
		problems = append(problems, p...)

		if !importsFlag {
			continue
		}

		importLog.Do(func() {
			// Only log this the first time it happens.
			debugf("package %s imports flag", pkg.PkgPath)
		})

		ast.Inspect(file, func(n ast.Node) bool {
			expr, ok := n.(ast.Expr)
			if !ok {
				return true
			}

			call, ok := expr.(*ast.CallExpr)
			if !ok {
				return true
			}

			msg, ok := checkCall(call, alias)
			if !ok {
				problems = append(problems, &Problem{
					pkg:     pkg,
					file:    file, // nolint:scopelint
					pos:     call.Pos(),
					message: msg,
				})
			}

			// We may miss calls that look like:
			//		flag.StringVar(flag.String("hello", "val", "usage"), "goodbye", "val", "usage"))
			// In this case we miss the inner one, but honestly I am okay with
			// that.
			return false
		})
	}

	importLog.Do(func() {
		// This will only invoke if no file in the package imports "flag"
		debugf("package %s does not import flag", pkg.PkgPath)
	})

	return problems
}

// processImports inspects a file within a given Go package. It checks if the
// file imports package "flag", and if aliased, what the alias is. It reports
// a problem if package flag is dot-imported, because we are unable to
// accurately do any further analysis on flag usage in the file. Technically, we
// are unable to process *any* file in the package at that point, but we still
// try.
func processImports(pkg *packages.Package, file *ast.File) (importsFlag bool, alias string, problems []*Problem) {
	for _, imp := range file.Imports {
		if imp.Path.Value != `"flag"` {
			continue
		}

		alias = "flag"
		if imp.Name != nil {
			alias = imp.Name.Name
		}

		if alias == "." {
			problems = append(problems, &Problem{
				pkg:     pkg,
				file:    file,
				pos:     imp.Pos(),
				message: "package flag should not be dot-imported",
			})

			return false, "", problems
		}

		return true, alias, problems
	}

	return false, "", nil
}

// checkCall returns true if it is a disallowed function call on the flag
// package.
func checkCall(call *ast.CallExpr, flagpkg string) (string, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", true
	}

	if selector.X == nil {
		return "", true
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", true
	}

	if ident.Name != flagpkg {
		return "", true
	}

	funcName := selector.Sel
	if allowedFlagFunctions[funcName.Name] {
		return "", true
	}

	return fmt.Sprintf("%s.%s should not be used on the global flagset", flagpkg, funcName.Name), false
}
