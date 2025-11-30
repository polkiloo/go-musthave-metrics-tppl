package analyzer

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
)

type context struct {
	insideMain bool
}

// Analyzer reports panic usage and restricted fatal exits.
var Analyzer = &analysis.Analyzer{
	Name: "projectlinter",
	Doc:  "checks for disallowed panic/log.Fatal/os.Exit usages",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	pkgIsMain := pass.Pkg.Name() == "main"

	var stack []context
	push := func(isMain bool) { stack = append(stack, context{insideMain: isMain}) }
	pop := func() {
		if len(stack) > 0 {
			stack = stack[:len(stack)-1]
		}
	}
	currentInsideMain := func() bool {
		if len(stack) == 0 {
			return false
		}
		return stack[len(stack)-1].insideMain
	}

	for _, file := range pass.Files {
		astutil.Apply(file, func(c *astutil.Cursor) bool {
			switch node := c.Node().(type) {
			case *ast.FuncDecl:
				inMain := pkgIsMain && node.Recv == nil && node.Name.Name == "main"
				if !inMain && len(stack) > 0 {
					inMain = currentInsideMain()
				}
				push(inMain)
			case *ast.FuncLit:
				push(currentInsideMain())
			case *ast.CallExpr:
				analyzeCall(pass, node, currentInsideMain())
			}
			return true
		}, func(c *astutil.Cursor) bool {
			switch c.Node().(type) {
			case *ast.FuncDecl, *ast.FuncLit:
				pop()
			}
			return true
		})
		stack = stack[:0]
	}

	return nil, nil
}

func analyzeCall(pass *analysis.Pass, call *ast.CallExpr, insideMain bool) {
	if id, ok := call.Fun.(*ast.Ident); ok {
		if obj, ok := pass.TypesInfo.Uses[id]; ok {
			if b, ok := obj.(*types.Builtin); ok && b.Name() == "panic" {
				pass.Reportf(call.Lparen, "avoid using panic")
				return
			}
		}
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	pkgPath, name := lookupSelector(pass, sel)
	switch pkgPath + ":" + name {
	case "log:Fatal":
		if !insideMain {
			pass.Reportf(call.Lparen, "log.Fatal should be called only from main function in package main")
		}
	case "os:Exit":
		if !insideMain {
			pass.Reportf(call.Lparen, "os.Exit should be called only from main function in package main")
		}
	}
}

func lookupSelector(pass *analysis.Pass, sel *ast.SelectorExpr) (string, string) {
	if obj, ok := pass.TypesInfo.Uses[sel.Sel]; ok && obj != nil && obj.Pkg() != nil {
		return obj.Pkg().Path(), obj.Name()
	}

	if pkgIdent, ok := sel.X.(*ast.Ident); ok {
		if pkgObj, ok := pass.TypesInfo.Uses[pkgIdent].(*types.PkgName); ok {
			return pkgObj.Imported().Path(), sel.Sel.Name
		}
		return pkgIdent.Name, sel.Sel.Name
	}

	return "", sel.Sel.Name
}
