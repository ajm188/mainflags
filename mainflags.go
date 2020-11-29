package mainflags

import (
	"flag"
	"fmt"
	"go/ast"
	"sync"

	"golang.org/x/tools/go/analysis"
)

var (
	logLevel             LogLevel = LogError
	allowedFlagFunctions          = map[string]bool{
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

const (
	logLevelHelp = `verbosity of logging. Should be LogError <= x < LogWarning for use in golangci-lint.
Set to a negative number to disable logging entirely`
)

func flagset() *flag.FlagSet {
	fs := flag.NewFlagSet("mainflags", flag.ContinueOnError)

	fs.IntVar((*int)(&logLevel), "verbosity", 0, logLevelHelp)

	return fs
}

func Analyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:  "mainflags",
		Doc:   "check that global flags are only defined in main packages",
		Run:   run,
		Flags: *flagset(),
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	var (
		flagsetSkipLog sync.Once
		importLog      sync.Once
		problems       = false
	)

	for _, file := range pass.Files {
		debugf("processing %s", file.Name.String())

		imp, ok := checkImports(file.Imports)
		if !ok {
			continue
		}

		importLog.Do(func() {
			// Only log this the first time it happens.
			debugf("package %s imports flag", pass.Pkg.Path())
		})

		if imp.alias == "." {
			pass.ReportRangef(imp.spec, "package flag should not be dot-imported")
			problems = true // nolint:wsl
			continue        // nolint:wsl
		}

		if pass.Pkg.Name() == "main" {
			flagsetSkipLog.Do(func() {
				debugf("%s is package main, skipping flagset check", pass.Pkg.Path())
			})

			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			problem, recurse := passInspect(pass, imp, n)
			problems = problems || problem
			return recurse
		})
	}

	importLog.Do(func() {
		// This will only invoke if no file in the package imports "flag"
		debugf("package %s does not import flag", pass.Pkg.Path())
	})

	if !problems {
		infof("no problems in package %s", pass.Pkg.Path())
	}

	return nil, nil
}

func passInspect(pass *analysis.Pass, imp *pkgImport, n ast.Node) (problem bool, recurse bool) {
	expr, ok := n.(ast.Expr)
	if !ok {
		return false, true
	}

	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false, true
	}

	msg, ok := checkCall(call, imp.alias)
	if !ok {
		pass.ReportRangef(call, msg)
		problem = true // nolint:wsl
	}

	// We may miss calls that look like:
	//		flag.StringVar(flag.String("hello", "val", "usage"), "goodbye", "val", "usage"))
	//
	// In this case we miss the inner one, but honestly I am okay with
	// that.
	return problem, false
}

type pkgImport struct {
	alias string
	spec  *ast.ImportSpec
}

func checkImports(imports []*ast.ImportSpec) (*pkgImport, bool) {
	for _, imp := range imports {
		if imp.Path == nil {
			continue
		}

		if imp.Path.Value != `"flag"` {
			continue
		}

		alias := "flag"
		if imp.Name != nil {
			alias = imp.Name.Name
		}

		return &pkgImport{
			alias: alias,
			spec:  imp,
		}, true
	}

	return nil, false
}

// checkCall returns false if it is a disallowed function call on the flag
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
