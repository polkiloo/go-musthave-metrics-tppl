package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMainDelegatesToRun(t *testing.T) {
	called := false
	originalRun := run
	run = func(dir string) error {
		called = true
		if dir == "" {
			t.Fatalf("expected directory value")
		}
		return nil
	}
	defer func() { run = originalRun }()

	oldFlags := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	defer func() { flag.CommandLine = oldFlags }()

	os.Args = []string{"reset", "-dir", filepath.Dir(os.TempDir())}
	main()

	if !called {
		t.Fatalf("expected run to be called")
	}
}

func TestMainHandlesError(t *testing.T) {
	called := false
	originalRun := run
	originalFatal := fatalf
	run = func(string) error { return fmt.Errorf("boom") }
	fatalf = func(string, ...interface{}) { called = true }
	defer func() {
		run = originalRun
		fatalf = originalFatal
	}()

	oldFlags := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	defer func() { flag.CommandLine = oldFlags }()

	os.Args = []string{"reset"}
	main()

	if !called {
		t.Fatalf("expected fatalf to be invoked on error")
	}
}
