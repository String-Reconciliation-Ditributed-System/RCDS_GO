package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/String-Reconciliation-Ditributed-System/RCDS_GO/pkg/lib/algorithm/full_sync"
)

func TestRunVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "RCDS version") {
		t.Fatalf("expected version output, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"wat"}, &stdout, &stderr)

	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if !strings.Contains(stderr.String(), "Unknown command") {
		t.Fatalf("expected unknown command error, got %q", stderr.String())
	}
}

func TestParseCommandFlags(t *testing.T) {
	var stderr bytes.Buffer

	config, err := parseCommandFlags("client", []string{
		"--host", "0.0.0.0",
		"--port", "9090",
		"--mode", "file",
		"--file", "copy.bin",
		"--timeout", "5s",
		"--chunk-size", "1024",
	}, &stderr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if config.host != "0.0.0.0" || config.port != 9090 || config.mode != "file" {
		t.Fatalf("unexpected parsed config: %+v", config)
	}
	if config.timeout != 5*time.Second || config.chunkSize != 1024 {
		t.Fatalf("unexpected file options: %+v", config)
	}
}

func TestParseCommandFlagsRejectsInvalidValues(t *testing.T) {
	tests := [][]string{
		{"--port", "0"},
		{"--mode", "invalid"},
		{"--algorithm", "invalid"},
		{"--expected-diff", "0"},
		{"--max-retries", "-1"},
		{"--chunk-size", "0"},
		{"unexpected"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			var stderr bytes.Buffer
			if _, err := parseCommandFlags("client", args, &stderr); err == nil {
				t.Fatalf("expected error for args %v", args)
			}
		})
	}
}

func TestConfiguredElementsFromItemsAndInput(t *testing.T) {
	tmp := t.TempDir()
	input := filepath.Join(tmp, "set.txt")
	if err := os.WriteFile(input, []byte("from-file\nshared\n\n"), 0644); err != nil {
		t.Fatal(err)
	}

	elements, err := configuredElements(&cliConfig{
		mode:  "set",
		items: "inline, shared ,",
		input: input,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []string{"inline", "shared", "from-file", "shared"}
	if strings.Join(elements, "|") != strings.Join(expected, "|") {
		t.Fatalf("expected %v, got %v", expected, elements)
	}
}

func TestWriteSetOutputSorted(t *testing.T) {
	syncer, err := full_sync.NewFullSetSync()
	if err != nil {
		t.Fatal(err)
	}
	for _, elem := range []string{"zeta", "alpha", "middle"} {
		if err := syncer.AddElement([]byte(elem)); err != nil {
			t.Fatal(err)
		}
	}

	output := filepath.Join(t.TempDir(), "set.out")
	if err := writeSetOutput(syncer, output); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "alpha\nmiddle\nzeta\n" {
		t.Fatalf("unexpected output: %q", string(data))
	}
}
