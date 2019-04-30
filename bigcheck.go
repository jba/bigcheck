package main

import (
	"flag"
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
		log.Fatal("need -threshold")
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

	isLargeType := func(t types.Type) bool { return stdSizes.Sizeof(t) >= smallestBadSize }

	isLarge := func(e ast.Expr) bool { return isLargeType(pass.TypesInfo.Types[e].Type) }

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
				if isLarge(a) {
					pass.Reportf(e.Pos(), "arg #%d is large", i)
				}
			}
		case *ast.AssignStmt:
			//ast.Print(pass.Fset, e)
			for i, r := range e.Rhs {
				if isLarge(r) {
					if len(e.Rhs) == 1 {
						pass.Reportf(e.Pos(), "rhs is large")
					} else {
						pass.Reportf(e.Pos(), "rhs #%d is large", i)
					}
				}
			}
		case *ast.SendStmt:
			if isLarge(e.Value) {
				pass.Reportf(e.Pos(), "value being sent is large")
			}
		case *ast.UnaryExpr:
			if e.Op == token.ARROW && isLarge(e) {
				pass.Reportf(e.Pos(), "received value is large")
			}
		case *ast.ReturnStmt:
			for i, r := range e.Results {
				if isLarge(r) {
					pass.Reportf(e.Pos(), "return value #%d is large", i)
				}
			}

		case *ast.RangeStmt:
			xt := pass.TypesInfo.Types[e.X].Type
			if e.Key != nil {
				if mt, ok := xt.(*types.Map); ok && isLargeType(mt.Key()) {
					pass.Reportf(e.Pos(), "ranged key is large")
				}
				if ct, ok := xt.(*types.Chan); ok && isLargeType(ct.Elem()) {
					pass.Reportf(e.Pos(), "ranged value is large")
				}
			}
			if e.Value != nil {
				if xt, ok := xt.(interface{ Elem() types.Type }); ok && isLargeType(xt.Elem()) {
					pass.Reportf(e.Pos(), "ranged value is large")
				}
			}

		default:
			panic("unexpected node type")
		}
	})
	return nil, nil
}
