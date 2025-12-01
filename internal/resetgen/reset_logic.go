package resetgen

import (
	"bytes"
	"fmt"
	"go/types"
)

func (g *generator) generateResetMethod(target resetTarget) string {
	typeObj, ok := g.typeNames[target.name]
	if !ok {
		return ""
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "func (obj *%s) Reset() {\n", target.name)
	buf.WriteString("\tif obj == nil {\n\t\treturn\n\t}\n\n")

	structType, ok := typeObj.Underlying().(*types.Struct)
	if !ok {
		return ""
	}

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		if !field.Exported() && field.Pkg() != nil && field.Pkg().Path() != g.pkg.PkgPath {
			continue
		}

		fieldName := field.Name()
		fieldType := field.Type()

		statements := g.resetStatements(fmt.Sprintf("obj.%s", fieldName), fieldType)
		for _, stmt := range statements {
			fmt.Fprintf(&buf, "\t%s\n", stmt)
		}
	}

	buf.WriteString("}\n")

	return buf.String()
}

func (g *generator) resetStatements(expr string, t types.Type) []string {
	if ptr, ok := t.(*types.Pointer); ok {
		return g.resetPointer(expr, ptr)
	}

	switch t.Underlying().(type) {
	case *types.Slice:
		return []string{fmt.Sprintf("%s = %s[:0]", expr, g.wrapForIndex(expr))}
	case *types.Map:
		return []string{fmt.Sprintf("clear(%s)", expr)}
	case *types.Struct:
		if g.hasResetMethod(t) {
			return []string{fmt.Sprintf("%s.Reset()", expr)}
		}
		return []string{fmt.Sprintf("%s = %s{}", expr, g.typeString(t))}
	case *types.Interface:
		if g.hasResetMethod(t) {
			return []string{fmt.Sprintf("if %s != nil { %s.Reset() }", expr, expr)}
		}
		return []string{
			fmt.Sprintf("if %s != nil {", expr),
			"\tif resetter, ok := " + expr + ".(interface{ Reset() }); ok {",
			"\t\tresetter.Reset()",
			"\t}",
			"}",
		}
	default:
		return []string{fmt.Sprintf("%s = %s", expr, g.zeroValue(t))}
	}
}

func (g *generator) resetPointer(expr string, ptr *types.Pointer) []string {
	elem := ptr.Elem()
	var lines []string
	lines = append(lines, fmt.Sprintf("if %s != nil {", expr))

	if g.hasResetMethod(ptr) {
		lines = append(lines, fmt.Sprintf("\t%s.Reset()", expr))
	} else {
		switch elem.Underlying().(type) {
		case *types.Slice:
			lines = append(lines, fmt.Sprintf("\t*%s = %s[:0]", expr, g.wrapForIndex("*"+expr)))
		case *types.Map:
			lines = append(lines, fmt.Sprintf("\tclear(*%s)", expr))
		default:
			nested := g.resetStatements("*"+expr, elem)
			for _, stmt := range nested {
				lines = append(lines, "\t"+stmt)
			}
		}
	}

	lines = append(lines, "}")
	return lines
}

func (g *generator) zeroValue(t types.Type) string {
	switch u := t.Underlying().(type) {
	case *types.Basic:
		switch u.Kind() {
		case types.String:
			return "\"\""
		case types.Bool:
			return "false"
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr,
			types.Float32, types.Float64, types.Complex64, types.Complex128,
			types.UnsafePointer, types.UntypedInt, types.UntypedRune, types.UntypedFloat, types.UntypedComplex:
			return "0"
		default:
			return "nil"
		}
	case *types.Interface:
		return "nil"
	default:
		return fmt.Sprintf("%s{}", g.typeString(t))
	}
}

func (g *generator) wrapForIndex(expr string) string {
	return fmt.Sprintf("(%s)", expr)
}

func (g *generator) hasResetMethod(t types.Type) bool {
	if t == nil {
		return false
	}

	if methodSetHasReset(types.NewMethodSet(t)) {
		return true
	}

	if _, ok := t.(*types.Pointer); !ok {
		if methodSetHasReset(types.NewMethodSet(types.NewPointer(t))) {
			return true
		}
	}

	return false
}

func methodSetHasReset(ms *types.MethodSet) bool {
	for i := 0; i < ms.Len(); i++ {
		obj := ms.At(i)
		if obj.Obj().Name() != "Reset" {
			continue
		}

		if sig, ok := obj.Obj().Type().(*types.Signature); ok {
			if sig.Params().Len() == 0 && sig.Results().Len() == 0 {
				return true
			}
		}
	}
	return false
}
