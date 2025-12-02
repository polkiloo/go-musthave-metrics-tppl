package resetgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func TestRunGeneratesResetMethods(t *testing.T) {
	dir := t.TempDir()
	copyTestdata(t, dir, "run")

	if err := Run(dir); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	generated := readFile(t, filepath.Join(dir, "reset.gen.go"))

	expected := readFile(t, filepath.Join(TestData(), "run", "reset.gen.want"))
	formattedExpected, err := formatSourceFunc([]byte(expected))
	if err != nil {
		t.Fatalf("failed to format expected source: %v", err)
	}

	formattedGenerated, err := formatSourceFunc([]byte(generated))
	if err != nil {
		t.Fatalf("failed to format generated source: %v", err)
	}

	if string(formattedGenerated) != string(formattedExpected) {
		t.Fatalf("generated reset mismatch\nExpected:\n%s\nGot:\n%s", formattedExpected, formattedGenerated)
	}
}

func TestRunSkipsPackagesWithoutTargets(t *testing.T) {
	dir := t.TempDir()
	copyTestdata(t, dir, "empty")

	if err := Run(dir); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "reset.gen.go")); err == nil || !os.IsNotExist(err) {
		t.Fatalf("unexpected generated file presence, err=%v", err)
	}
}

func TestErrorPaths(t *testing.T) {
	dir := t.TempDir()
	copyTestdata(t, dir, "errorcase")

	loaderErr := fmt.Errorf("boom")
	originalLoader := loadPackages
	loadPackages = func(_ *packages.Config, _ ...string) ([]*packages.Package, error) {
		return nil, loaderErr
	}
	t.Cleanup(func() { loadPackages = originalLoader })

	if err := Run(dir); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected loader error, got %v", err)
	}

	g := &generator{pkg: &packages.Package{PkgPath: "example.com/errorcase", Name: "errorcase"}, imports: map[string]string{}, fset: token.NewFileSet(), typeNames: map[string]types.Type{}}
	g.filePath = filepath.Join(dir, "reset.gen.go")

	originalFormatter := formatSourceFunc
	formatSourceFunc = func(_ []byte) ([]byte, error) { return nil, fmt.Errorf("format failure") }
	originalWriter := writeFileFunc
	writeFileFunc = writeFile
	t.Cleanup(func() {
		formatSourceFunc = originalFormatter
		writeFileFunc = originalWriter
	})
	if err := g.generate(); err == nil || !strings.Contains(err.Error(), "format failure") {
		t.Fatalf("expected format failure, got %v", err)
	}

	formatSourceFunc = formatSource
	writeFileFunc = func(string, []byte) error { return fmt.Errorf("write failure") }

	if err := g.generate(); err == nil || !strings.Contains(err.Error(), "write failure") {
		t.Fatalf("expected write failure, got %v", err)
	}
}
