package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("check-workflows-yaml", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := checkWorkflows(*root); err != nil {
		return err
	}

	fmt.Fprintln(stdout, "yaml ok")
	return nil
}

func checkWorkflows(root string) error {
	pattern := filepath.Join(root, ".github", "workflows", "*.yml")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("resolver workflows: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no se encontraron workflows en %s", pattern)
	}

	for _, file := range files {
		body, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("leer %s: %w", file, err)
		}
		var node yaml.Node
		if err := yaml.Unmarshal(body, &node); err != nil {
			return fmt.Errorf("YAML inválido en %s: %w", file, err)
		}
	}
	return nil
}
