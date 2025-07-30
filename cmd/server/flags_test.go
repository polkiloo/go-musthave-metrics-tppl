package main

import (
	"strings"
	"testing"
)

func TestParseFlags_Default(t *testing.T) {
	addr, err := parseFlags([]string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if addr != "localhost:8080" {
		t.Errorf("default addr = %q; want %q", addr, "localhost:8080")
	}
}

func TestParseFlags_EQSyntax(t *testing.T) {
	addr, err := parseFlags([]string{"-a=127.0.0.1:9000"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if addr != "127.0.0.1:9000" {
		t.Errorf("addr = %q; want %q", addr, "127.0.0.1:9000")
	}
}

func TestParseFlags_SpaceSyntax(t *testing.T) {
	addr, err := parseFlags([]string{"-a", "10.0.0.1:8081"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if addr != "10.0.0.1:8081" {
		t.Errorf("addr = %q; want %q", addr, "10.0.0.1:8081")
	}
}

func TestParseFlags_RepeatedFlags(t *testing.T) {
	args := []string{"-a", "first:1111", "-a", "second:2222"}
	addr, err := parseFlags(args)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if addr != "second:2222" {
		t.Errorf("addr = %q; want %q", addr, "second:2222")
	}
}

func TestParseFlags_UnknownFlag(t *testing.T) {
	_, err := parseFlags([]string{"-x"})
	if err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
	if !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Errorf("unexpected error for unknown flag: %v", err)
	}
}

func TestParseFlags_PositionalArg(t *testing.T) {
	_, err := parseFlags([]string{"foo"})
	if err == nil {
		t.Fatal("expected error for positional arg, got nil")
	}
	if !strings.Contains(err.Error(), "unknown arguments") {
		t.Errorf("unexpected error for positional arg: %v", err)
	}
}

func TestParseFlags_MissingValue(t *testing.T) {
	_, err := parseFlags([]string{"-a"})
	if err == nil {
		t.Fatal("expected error for missing -a value, got nil")
	}
	if !strings.Contains(err.Error(), "flag needs an argument") {
		t.Errorf("unexpected error for missing value: %v", err)
	}
}
