package main

import "testing"

func TestParseArgsAgentCommand(t *testing.T) {
	cfg, command, args, err := parseArgs([]string{"--json", "--limit", "3", "bt", "Pink Floyd", "FLAC"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}
	if !cfg.jsonOutput {
		t.Fatal("expected json output to be enabled")
	}
	if cfg.limit != 3 {
		t.Fatalf("limit = %d, want 3", cfg.limit)
	}
	if cfg.pick != 1 {
		t.Fatalf("pick = %d, want 1", cfg.pick)
	}
	if command != "bt" {
		t.Fatalf("command = %q, want bt", command)
	}
	if got := len(args); got != 2 {
		t.Fatalf("len(args) = %d, want 2", got)
	}
}

func TestParseArgsLegacyOutputDir(t *testing.T) {
	cfg, command, args, err := parseArgs([]string{"/tmp/music"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}
	if command != "" {
		t.Fatalf("command = %q, want empty interactive command", command)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v, want empty", args)
	}
	if cfg.destDir != "/tmp/music" {
		t.Fatalf("destDir = %q, want /tmp/music", cfg.destDir)
	}
}

func TestParseArgsGetCommand(t *testing.T) {
	cfg, command, args, err := parseArgs([]string{"--json", "--pick", "2", "--output-dir", "/tmp/music", "get", "周杰伦", "以父之名"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}
	if !cfg.jsonOutput || cfg.pick != 2 || cfg.destDir != "/tmp/music" {
		t.Fatalf("cfg = %+v, want json pick=2 destDir=/tmp/music", cfg)
	}
	if command != "get" {
		t.Fatalf("command = %q, want get", command)
	}
	if len(args) != 2 {
		t.Fatalf("args = %v, want two keywords", args)
	}
}

func TestSanitizeFilename(t *testing.T) {
	got := sanitizeFilename(`a/b\c:d*e?f"g<h>i|j`)
	want := "a-b-c-d-efghi-j"
	if got != want {
		t.Fatalf("sanitizeFilename() = %q, want %q", got, want)
	}
}
