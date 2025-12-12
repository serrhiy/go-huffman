package main

import (
	"path/filepath"
	"testing"
)

func TestGetArgumentsEncode(t *testing.T) {
	input := "/tmp/input.txt"

	t.Run("empty input", func(t *testing.T) {
		_, err := getArgumentsEncode("", "/tmp/out.huff")
		if err == nil {
			t.Fatal("expected error for empty input")
		}
	})

	t.Run("output empty", func(t *testing.T) {
		args, err := getArgumentsEncode(input, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Join(filepath.Dir(input), "input"+OutputExtension)
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
}

func TestGetArgumentsDecode(t *testing.T) {
	input := "/tmp/input.huff"

	t.Run("empty input", func(t *testing.T) {
		_, err := getArgumentsDecode("", "/tmp/out.txt")
		if err == nil {
			t.Fatal("expected error for empty input")
		}
	})

	t.Run("empty output", func(t *testing.T) {
		_, err := getArgumentsDecode(input, "")
		if err == nil {
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
	t.Run("invalid service", func(t *testing.T) {
		_, err := getArguments("in", "out", "x")
		if err == nil {
			t.Fatal("expected error for invalid service")
		}
	})

	t.Run("encode service with empty output", func(t *testing.T) {
		input := "/home/user/file.txt"
		args, err := getArguments(input, "", "e")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Join(filepath.Dir(input), "file"+OutputExtension)
		if args.outputFile != expected {
			t.Fatalf("expected output '%s', got '%s'", expected, args.outputFile)
		}
	})

	t.Run("decode service normal", func(t *testing.T) {
		input := "/tmp/input.huff"
		output := "/tmp/out.txt"
		args, err := getArguments(input, output, "d")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if args.inputFile != input || args.outputFile != output {
			t.Fatalf("expected args (%s, %s), got (%s, %s)", input, output, args.inputFile, args.outputFile)
		}
	})
}

func TestGetArgumentsEncode_FileWithoutExt(t *testing.T) {
	input := "/tmp/file_without_ext"
	args, err := getArgumentsEncode(input, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(filepath.Dir(input), "file_without_ext"+OutputExtension)
	if args.outputFile != expected {
		t.Fatalf("expected output '%s', got '%s'", expected, args.outputFile)
	}
}

func TestGetArgumentsEncode_ComplexName(t *testing.T) {
	input := "/some/dir/myfile.data.txt"
	args, err := getArgumentsEncode(input, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(filepath.Dir(input), "myfile.data"+OutputExtension)
	if args.outputFile != expected {
		t.Fatalf("expected output '%s', got '%s'", expected, args.outputFile)
	}
}
