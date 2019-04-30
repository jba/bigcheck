package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/ast/inspector"
)

var size = flag.Int("size", 0, "smallest size that will trigger a message")

func main() {
	flag.Parse()
	if *size == 0 {
		log.Fatal("need -size")
	}
	singlechecker.Main(analyzer)
}

var analyzer = &analysis.Analyzer{
	Name:     "bigcheck",
	Doc:      "check for copying large values",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var stdSizes = types.StdSizes{WordSize: 8, MaxAlign: 1}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	smallestBadSize := int64(*size)

	largetypef := func(t types.Type, pos token.Pos, format string, args ...interface{}) {

		if t == nil {
			// No type information.
			// Could be the variable assigned to in a type switch, for example.
			return
		}
		// Untyped types are never large.
		if bt, ok := t.Underlying().(*types.Basic); ok && bt.Info()&types.IsUntyped != 0 {
			return
		}
		sz := stdSizes.Sizeof(t)
		if sz >= smallestBadSize {
			pass.Reportf(pos, "%s is %d bytes", fmt.Sprintf(format, args...), sz)
		}
	}

	largef := func(e ast.Expr, pos token.Pos, format string, args ...interface{}) {
		largetypef(pass.TypesInfo.Types[e].Type, pos, format, args...)
	}

	nodeFilter := []ast.Node{
		(*ast.AssignStmt)(nil),
		(*ast.CallExpr)(nil),
		(*ast.RangeStmt)(nil),
		(*ast.ReturnStmt)(nil),
		(*ast.SendStmt)(nil),
		(*ast.UnaryExpr)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch e := n.(type) {
		case *ast.CallExpr:
			for i, a := range e.Args {
				largef(a, e.Pos(), "arg #%d", i)
			}
		case *ast.AssignStmt:
			for i, r := range e.Rhs {
				largef(r, e.Pos(), "rhs #%d", i)
			}
		case *ast.SendStmt:
			largef(e.Value, e.Pos(), "value being sent")

		case *ast.UnaryExpr:
			if e.Op == token.ARROW {
				largef(e, e.Pos(), "received value")
			}

		case *ast.ReturnStmt:
			for i, r := range e.Results {
				largef(r, e.Pos(), "return value #%d", i)
			}

		case *ast.RangeStmt:
			xt := pass.TypesInfo.Types[e.X].Type
			if e.Key != nil {
				if mt, ok := xt.(*types.Map); ok {
					largetypef(mt.Key(), e.Pos(), "ranged key")
				}
				if ct, ok := xt.(*types.Chan); ok {
					largetypef(ct.Elem(), e.Pos(), "ranged value")
				}
			}
			if e.Value != nil {
				if xt, ok := xt.(interface{ Elem() types.Type }); ok {
					largetypef(xt.Elem(), e.Pos(), "ranged value")
				}
			}

		default:
			panic("unexpected node type")
		}
	})
	return nil, nil
}
