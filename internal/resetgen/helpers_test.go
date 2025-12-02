package resetgen

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"go/ast"
	"go/parser"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func TestFormatSourceError(t *testing.T) {
	if _, err := formatSource([]byte("not go code")); err == nil {
		t.Fatalf("expected format error")
	}
}

func TestGenerateResetMethodEdgeCases(t *testing.T) {
	g := &generator{pkg: &packages.Package{PkgPath: "example.com/p", Name: "p", Types: types.NewPackage("example.com/p", "p")}, imports: map[string]string{}, fset: token.NewFileSet(), typeNames: map[string]types.Type{}}

	if got := g.generateResetMethod(resetTarget{name: "Missing"}); got != "" {
		t.Fatalf("expected empty generation for missing type, got %q", got)
	}

	g.typeNames["Alias"] = types.Typ[types.Int]
	if got := g.generateResetMethod(resetTarget{name: "Alias"}); got != "" {
		t.Fatalf("expected empty generation for non-struct type, got %q", got)
	}

	otherPkg := types.NewPackage("example.com/other", "other")
	hidden := types.NewVar(token.NoPos, otherPkg, "hidden", types.Typ[types.Int])
	structType := types.NewStruct([]*types.Var{hidden}, []string{""})
	name := types.NewTypeName(token.NoPos, g.pkg.Types, "WithHidden", nil)
	g.typeNames["WithHidden"] = types.NewNamed(name, structType, nil)

	generated := g.generateResetMethod(resetTarget{name: "WithHidden"})
	if strings.Contains(generated, "hidden") {
		t.Fatalf("expected hidden field to be skipped, got %s", generated)
	}
}

func TestZeroValueAndTypeStringBranches(t *testing.T) {
	pkgOne := types.NewPackage("example.com/one", "dup")
	pkgTwo := types.NewPackage("example.com/two", "dup")
	namedOne := types.NewNamed(types.NewTypeName(0, pkgOne, "Thing", nil), types.NewStruct(nil, nil), nil)
	namedTwo := types.NewNamed(types.NewTypeName(0, pkgTwo, "Other", nil), types.NewStruct(nil, nil), nil)

	g := &generator{pkg: &packages.Package{PkgPath: "example.com/p", Name: "p", Types: types.NewPackage("example.com/p", "p")}, imports: map[string]string{"dup": "example.com/one"}, fset: token.NewFileSet(), typeNames: map[string]types.Type{}}

	if got := g.zeroValue(types.Typ[types.UntypedNil]); got != "nil" {
		t.Fatalf("unexpected zero value for untyped nil: %s", got)
	}

	typeStr := g.zeroValue(namedTwo)
	if typeStr != "dup1.Other{}" {
		t.Fatalf("unexpected zero value for named struct: %s", typeStr)
	}

	if g.imports["dup1"] != "example.com/two" {
		t.Fatalf("expected alias dup1 recorded, got %v", g.imports)
	}

	if g.typeString(namedOne) != "dup.Thing" {
		t.Fatalf("expected reuse of existing alias for first package")
	}

	if got := g.zeroValue(types.NewInterfaceType(nil, nil)); got != "nil" {
		t.Fatalf("expected nil for interface zero value, got %s", got)
	}
}

func TestHasResetMethodVariants(t *testing.T) {
	pkg := types.NewPackage("example.com/p", "p")
	g := &generator{pkg: &packages.Package{PkgPath: "example.com/p", Name: "p", Types: pkg}, imports: map[string]string{}, fset: token.NewFileSet(), typeNames: map[string]types.Type{}}

	if g.hasResetMethod(nil) {
		t.Fatalf("expected nil type to have no Reset method")
	}

	strct := types.NewNamed(types.NewTypeName(0, pkg, "Local", nil), types.NewStruct(nil, nil), nil)
	recv := types.NewVar(token.NoPos, pkg, "", types.NewPointer(strct))
	sig := types.NewSignatureType(recv, nil, nil, nil, nil, false)
	ptrRecv := types.NewFunc(token.NoPos, pkg, "Reset", sig)
	strct.AddMethod(ptrRecv)

	if !g.hasResetMethod(strct) {
		t.Fatalf("expected pointer receiver reset to be detected on value type")
	}
}

func TestHasGenerateResetCommentVariants(t *testing.T) {
	declDoc := &ast.GenDecl{Tok: token.TYPE, Doc: &ast.CommentGroup{List: []*ast.Comment{{Text: "// generate:reset"}}}}
	spec := &ast.TypeSpec{Name: ast.NewIdent("A")}
	if !hasGenerateResetComment(declDoc, spec) {
		t.Fatalf("expected detection via decl doc")
	}

	specDoc := &ast.TypeSpec{Name: ast.NewIdent("B"), Doc: &ast.CommentGroup{List: []*ast.Comment{{Text: "// generate:reset"}}}}
	decl := &ast.GenDecl{Tok: token.TYPE}
	if !hasGenerateResetComment(decl, specDoc) {
		t.Fatalf("expected detection via spec doc")
	}

	none := &ast.GenDecl{Tok: token.TYPE}
	if hasGenerateResetComment(none, &ast.TypeSpec{Name: ast.NewIdent("C")}) {
		t.Fatalf("expected no detection without directive")
	}
}

func TestRunSkipsEmptyPackages(t *testing.T) {
	originalLoader := loadPackages
	defer func() { loadPackages = originalLoader }()

	loadPackages = func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
		return []*packages.Package{{GoFiles: nil}}, nil
	}

	if err := Run(t.TempDir()); err != nil {
		t.Fatalf("expected Run to succeed with empty packages, got %v", err)
	}

	loadPackages = func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
		return []*packages.Package{{GoFiles: []string{"file.go"}, TypesInfo: &types.Info{Defs: map[*ast.Ident]types.Object{}}, Fset: token.NewFileSet()}}, nil
	}

	if err := Run(t.TempDir()); err != nil {
		t.Fatalf("expected Run to skip packages without targets, got %v", err)
	}
}

func TestRunHandlesAbsError(t *testing.T) {
	originalLoader := loadPackages
	originalAbs := absPath
	loadPackages = func(_ *packages.Config, _ ...string) ([]*packages.Package, error) { return nil, nil }
	absPath = func(string) (string, error) { return "", fmt.Errorf("abs fail") }
	defer func() {
		loadPackages = originalLoader
		absPath = originalAbs
	}()

	if err := Run("bad"); err == nil || !strings.Contains(err.Error(), "abs fail") {
		t.Fatalf("expected abs failure, got %v", err)
	}
}

func TestCollectTargetsSkipsUnsupportedTypes(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "skip.go")
	src := readFile(t, filepath.Join(TestData(), "collect", "skip.go"))

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	pkgTypes := types.NewPackage("example.com/sample", "sample")
	identAlias := file.Scope.Objects["Alias"].Decl.(*ast.TypeSpec).Name
	identStruct := file.Scope.Objects["NoDirective"].Decl.(*ast.TypeSpec).Name

	info := &types.Info{Defs: map[*ast.Ident]types.Object{
		identAlias:  nil,
		identStruct: types.NewTypeName(token.NoPos, pkgTypes, "NoDirective", types.NewStruct(nil, nil)),
	}}

	pkg := &packages.Package{PkgPath: pkgTypes.Path(), Name: pkgTypes.Name(), TypesInfo: info, Syntax: []*ast.File{file}, Fset: fset}
	g := newGenerator(pkg)

	g.collectTargets()

	if len(g.targets) != 0 {
		t.Fatalf("expected no targets, got %d", len(g.targets))
	}
}

func TestCollectTargetsHandlesNonTypeSpec(t *testing.T) {
	file := &ast.File{Decls: []ast.Decl{&ast.GenDecl{Tok: token.TYPE, Specs: []ast.Spec{&ast.ValueSpec{}}}}}
	pkg := &packages.Package{TypesInfo: &types.Info{Defs: map[*ast.Ident]types.Object{}}, Syntax: []*ast.File{file}, Fset: token.NewFileSet()}
	g := newGenerator(pkg)

	g.collectTargets()

	if len(g.targets) != 0 {
		t.Fatalf("expected no targets, got %d", len(g.targets))
	}
}

func TestRunPropagatesGenerationError(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "sample.go")

	src := "package sample\n\n// generate:reset\ntype Target struct {\n    Value int\n}"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	pkgTypes := types.NewPackage("example.com/sample", "sample")
	typeName := types.NewTypeName(token.NoPos, pkgTypes, "Target", nil)
	structType := types.NewStruct(nil, nil)
	_ = types.NewNamed(typeName, structType, nil)

	var targetIdent *ast.Ident
	for _, decl := range file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range gen.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.Name == "Target" {
					targetIdent = ts.Name
				}
			}
		}
	}

	if targetIdent == nil {
		t.Fatalf("type ident not found")
	}

	info := &types.Info{Defs: map[*ast.Ident]types.Object{targetIdent: typeName}}

	originalLoader := loadPackages
	originalWriter := writeFileFunc
	loadPackages = func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
		return []*packages.Package{{PkgPath: pkgTypes.Path(), Name: pkgTypes.Name(), GoFiles: []string{filePath}, Syntax: []*ast.File{file}, TypesInfo: info, Fset: fset}}, nil
	}
	writeFileFunc = func(string, []byte) error { return fmt.Errorf("write fail") }
	defer func() {
		loadPackages = originalLoader
		writeFileFunc = originalWriter
	}()

	if err := Run(dir); err == nil || !strings.Contains(err.Error(), "write fail") {
		t.Fatalf("expected write failure propagated, got %v", err)
	}
}
