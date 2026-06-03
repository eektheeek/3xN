package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/eektheeek/dead-lift-project/math-engine/internal/analysis"
	"github.com/eektheeek/dead-lift-project/math-engine/internal/contracts"
	"github.com/eektheeek/dead-lift-project/math-engine/internal/validate"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	var exitCode int
	switch os.Args[1] {
	case "validate-input":
		exitCode = runValidateInput(os.Args[2:])
	case "run-analysis":
		exitCode = runAnalysis(os.Args[2:])
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		printUsage()
		exitCode = 2
	}
	os.Exit(exitCode)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage:
  core-cli validate-input -f <core_input_v1.json>
  core-cli run-analysis     -f <core_input_v1.json>

`)
}

func runValidateInput(args []string) int {
	fs := flag.NewFlagSet("validate-input", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	file := fs.String("f", "", "path to core_input_v1 JSON file")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *file == "" {
		fmt.Fprintln(os.Stderr, "validate-input: -f is required")
		return 2
	}

	in, err := loadInput(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "validate-input: %v\n", err)
		return 1
	}
	if err := validate.Validate(in); err != nil {
		fmt.Fprintf(os.Stderr, "validate-input: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "ok: %s\n", *file)
	return 0
}

func runAnalysis(args []string) int {
	fs := flag.NewFlagSet("run-analysis", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	file := fs.String("f", "", "path to core_input_v1 JSON file")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *file == "" {
		fmt.Fprintln(os.Stderr, "run-analysis: -f is required")
		return 2
	}

	in, err := loadInput(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run-analysis: %v\n", err)
		return 1
	}
	if err := validate.Validate(in); err != nil {
		fmt.Fprintf(os.Stderr, "run-analysis: validate: %v\n", err)
		return 1
	}

	out, err := analysis.Run(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run-analysis: %v\n", err)
		return 1
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "run-analysis: encode: %v\n", err)
		return 1
	}
	return 0
}

func loadInput(path string) (contracts.CoreInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return contracts.CoreInput{}, err
	}
	var in contracts.CoreInput
	if err := json.Unmarshal(data, &in); err != nil {
		return contracts.CoreInput{}, err
	}
	return in, nil
}
