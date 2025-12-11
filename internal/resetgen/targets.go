package resetgen

import (
	"go/ast"
	"go/token"
	"go/types"
)

type resetTarget struct {
	name     string
	typeSpec *ast.TypeSpec
}

func (g *generator) collectTargets() {
	for ident, obj := range g.pkg.TypesInfo.Defs {
		if obj == nil {
			continue
		}
		if typeName, ok := obj.(*types.TypeName); ok {
			g.typeNames[ident.Name] = typeName.Type()
		}
	}

	for _, file := range g.pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if !ok || decl.Tok != token.TYPE {
				return true
			}

			for _, spec := range decl.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if _, ok := ts.Type.(*ast.StructType); !ok {
					continue
				}

				if !hasGenerateResetComment(decl, ts) {
					continue
				}

				g.targets = append(g.targets, resetTarget{name: ts.Name.Name, typeSpec: ts})
			}

			return false
		})
	}
}

func hasGenerateResetComment(decl *ast.GenDecl, spec *ast.TypeSpec) bool {
	if spec.Doc != nil {
		for _, c := range spec.Doc.List {
			if containsResetDirective(c.Text) {
				return true
			}
		}
	}

	if decl.Doc != nil {
		for _, c := range decl.Doc.List {
			if containsResetDirective(c.Text) {
				return true
			}
		}
	}

	return false
}

func containsResetDirective(text string) bool {
	return text == "// generate:reset"
}
