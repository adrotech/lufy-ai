package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) < 2 {
		return usageError()
	}

	switch args[0] {
	case "get-version":
		if len(args) != 2 {
			return usageError()
		}
		data, err := readUpstream(args[1])
		if err != nil {
			return err
		}
		version, ok := data["effectiveOpenSpecVersion"].(string)
		if !ok || version == "" {
			return fmt.Errorf("%s no contiene effectiveOpenSpecVersion string", args[1])
		}
		fmt.Fprintln(stdout, version)
	case "set-version":
		if len(args) < 3 {
			return usageError()
		}
		version := args[1]
		for _, file := range args[2:] {
			data, err := readUpstream(file)
			if err != nil {
				return err
			}
			data["effectiveOpenSpecVersion"] = version
			if err := writeUpstream(file, data); err != nil {
				return err
			}
		}
	default:
		return usageError()
	}
	return nil
}

func readUpstream(path string) (map[string]any, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("leer %s: %w", path, err)
	}
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("JSON inválido en %s: %w", path, err)
	}
	return data, nil
}

func writeUpstream(path string, data map[string]any) error {
	body, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("serializar %s: %w", path, err)
	}
	body = append(body, '\n')
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("escribir %s: %w", path, err)
	}
	return nil
}

func usageError() error {
	return errors.New("uso: openspec-upstream get-version <file> | set-version <version> <file> [file...]")
}
