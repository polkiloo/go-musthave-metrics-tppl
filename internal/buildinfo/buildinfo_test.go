package buildinfo

import (
	"bytes"
	"testing"
)

func TestParseVersionFile(t *testing.T) {
	contents := "version=v1.2.3\ncommit=abc123\ndate=2024-01-01T00:00:00Z\n"

	info := parseVersionFile(contents)

	if info.Version != "v1.2.3" {
		t.Fatalf("expected version v1.2.3, got %s", info.Version)
	}

	if info.Commit != "abc123" {
		t.Fatalf("expected commit abc123, got %s", info.Commit)
	}

	if info.Date != "2024-01-01T00:00:00Z" {
		t.Fatalf("expected date 2024-01-01T00:00:00Z, got %s", info.Date)
	}
}

func TestParseVersionFileMissingValues(t *testing.T) {
	contents := "version=\ncommit=\n"

	info := parseVersionFile(contents)

	if info.Version != "" {
		t.Fatalf("expected empty version, got %s", info.Version)
	}

	if info.Date != "" {
		t.Fatalf("expected empty date, got %s", info.Date)
	}

	if info.Commit != "" {
		t.Fatalf("expected empty commit, got %s", info.Commit)
	}
}

func TestPrint(t *testing.T) {
	buf := &bytes.Buffer{}
	info := Info{Version: "v1.0.0", Date: "2024-02-02", Commit: "deadbeef"}

	Print(buf, info)

	expected := "Build version: v1.0.0\nBuild date: 2024-02-02\nBuild commit: deadbeef\n"
	if buf.String() != expected {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestPrintWithMissingValues(t *testing.T) {
	buf := &bytes.Buffer{}
	info := Info{}

	Print(buf, info)

	expected := "Build version: N/A\nBuild date: N/A\nBuild commit: N/A\n"
	if buf.String() != expected {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
