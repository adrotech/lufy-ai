package domain

import (
	"bytes"
	"io"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

func DecodeAssignment(input io.Reader) (Assignment, error) {
	var value Assignment
	if err := decodeAllocationDocument(input, &value); err != nil {
		return Assignment{}, err
	}
	if err := ValidateAssignment(value); err != nil {
		return Assignment{}, err
	}
	return value, nil
}

func DecodeYieldCheckpoint(input io.Reader) (YieldCheckpoint, error) {
	var value YieldCheckpoint
	if err := decodeAllocationDocument(input, &value); err != nil {
		return YieldCheckpoint{}, err
	}
	if err := ValidateYieldCheckpoint(value); err != nil {
		return YieldCheckpoint{}, err
	}
	return value, nil
}

func decodeAllocationDocument(input io.Reader, destination any) error {
	if input == nil {
		return fieldError("$", "proveer YAML o JSON")
	}
	body, err := io.ReadAll(io.LimitReader(input, maxInputBytes+1))
	if err != nil {
		return fieldError("$", "proveer una entrada legible")
	}
	if len(body) > maxInputBytes {
		return fieldError("$", "reducir el documento a 64 KiB o menos")
	}
	if !utf8.Valid(body) {
		return fieldError("$", "usar UTF-8 válido")
	}
	if err := rejectUnsafeYAML(body); err != nil {
		return err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(body))
	decoder.KnownFields(true)
	if err := decoder.Decode(destination); err != nil {
		return fieldError("$", "corregir sintaxis, claves duplicadas o campos no documentados")
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fieldError("$", "proveer exactamente un documento")
	}
	return nil
}
