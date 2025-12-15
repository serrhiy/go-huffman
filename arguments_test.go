package main

import (
	"testing"
)

func TestGetArgumentsEncode(t *testing.T) {
	input := "/tmp/input.txt"

	t.Run("empty input", func(t *testing.T) {
		if _, err := getArgumentsEncode("", "/tmp/out.huff"); err == nil {
			t.Fatal("expected error for empty input")
		}
	})

	t.Run("output empty", func(t *testing.T) {
		args, err := getArgumentsEncode(input, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "/tmp/input.hfm"
		if args.outputFile != expected {
			t.Fatalf("expected output '%s', got '%s'", expected, args.outputFile)
		}
		if args.inputFile != input {
			t.Fatalf("expected input '%s', got '%s'", input, args.inputFile)
		}
	})

	t.Run("output provided", func(t *testing.T) {
		output := "/tmp/out.huff"
		args, err := getArgumentsEncode(input, output)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if args.outputFile != output {
			t.Fatalf("expected output '%s', got '%s'", output, args.outputFile)
		}
	})

	t.Run("auto output without extension", func(t *testing.T) {
		input := "/tmp/file_without_ext"
		args, err := getArgumentsEncode(input, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "/tmp/file_without_ext.hfm"
		if args.outputFile != expected {
			t.Fatalf("expected %q, got %q", expected, args.outputFile)
		}
	})

	t.Run("auto output with complex name", func(t *testing.T) {
		input := "/some/dir/myfile.data.txt"
		args, err := getArgumentsEncode(input, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "/some/dir/myfile.data.hfm"
		if args.outputFile != expected {
			t.Fatalf("expected %q, got %q", expected, args.outputFile)
		}
	})
}

func TestGetArgumentsDecode(t *testing.T) {
	input := "/tmp/input.huff"

	t.Run("empty input", func(t *testing.T) {
		if _, err := getArgumentsDecode("", "/tmp/out.txt"); err == nil {
			t.Fatal("expected error for empty input")
		}
	})

	t.Run("empty output", func(t *testing.T) {
		if _, err := getArgumentsDecode(input, ""); err == nil {
			t.Fatal("expected error for empty output")
		}
	})

	t.Run("both provided", func(t *testing.T) {
		output := "/tmp/out.txt"
		args, err := getArgumentsDecode(input, output)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if args.inputFile != input {
			t.Fatalf("expected input '%s', got '%s'", input, args.inputFile)
		}
		if args.outputFile != output {
			t.Fatalf("expected output '%s', got '%s'", output, args.outputFile)
		}
	})
}

func TestGetArguments(t *testing.T) {
	t.Run("both encode and decode set", func(t *testing.T) {
		_, err := getArguments("in.txt", "in.hfm", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("neither encode nor decode set", func(t *testing.T) {
		_, err := getArguments("", "", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("encode path only", func(t *testing.T) {
		args, err := getArguments("input.txt", "", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if args.inputFile != "input.txt" {
			t.Fatalf("unexpected input: %q", args.inputFile)
		}
	})

	t.Run("decode path only", func(t *testing.T) {
		args, err := getArguments("", "input.hfm", "out.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if args.inputFile != "input.hfm" || args.outputFile != "out.txt" {
			t.Fatalf("unexpected args: %+v", args)
		}
	})
}
