package resetgen

import (
	"bytes"
	"fmt"
	"go/token"
	"go/types"
	"sort"

	"golang.org/x/tools/go/packages"
)

type generator struct {
	pkg       *packages.Package
	targets   []resetTarget
	imports   map[string]string
	fset      *token.FileSet
	filePath  string
	typeNames map[string]types.Type
}

func newGenerator(pkg *packages.Package) *generator {
	return &generator{
		pkg:       pkg,
		imports:   map[string]string{},
		fset:      pkg.Fset,
		typeNames: map[string]types.Type{},
	}
}

func (g *generator) generate() error {
	var buf bytes.Buffer
	buf.WriteString(generatedHeader)
	fmt.Fprintf(&buf, "package %s\n\n", g.pkg.Name)

	g.buildContent(&buf)

	formatted, err := formatSourceFunc(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format source: %w", err)
	}

	if err := writeFileFunc(g.filePath, formatted); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (g *generator) buildContent(buf *bytes.Buffer) {
	var content bytes.Buffer
	for i, target := range g.targets {
		if i > 0 {
			content.WriteByte('\n')
		}
		content.WriteString(g.generateResetMethod(target))
	}

	if importsBlock := g.renderImports(); importsBlock != "" {
		buf.WriteString(importsBlock)
	}

	buf.Write(content.Bytes())
}

func (g *generator) renderImports() string {
	if len(g.imports) == 0 {
		return ""
	}

	aliases := make([]string, 0, len(g.imports))
	for alias := range g.imports {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)

	var buf bytes.Buffer
	buf.WriteString("import (\n")
	for _, alias := range aliases {
		fmt.Fprintf(&buf, "\t%s \"%s\"\n", alias, g.imports[alias])
	}
	buf.WriteString(")\n\n")
	return buf.String()
}

func (g *generator) typeString(t types.Type) string {
	qualifier := func(p *types.Package) string {
		if p.Path() == g.pkg.PkgPath {
			return ""
		}

		alias := p.Name()
		original := alias
		counter := 1
		for {
			if existing, ok := g.imports[alias]; !ok || existing == p.Path() {
				g.imports[alias] = p.Path()
				break
			}
			alias = fmt.Sprintf("%s%d", original, counter)
			counter++
		}

		return alias
	}

	return types.TypeString(t, qualifier)
}
