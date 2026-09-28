package domain

import (
	"bytes"
	"io"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

func DecodeEvaluation(input io.Reader) (EvaluationRequest, error) {
	if input == nil {
		return EvaluationRequest{}, fieldError("$", "proveer YAML o JSON")
	}
	body, err := io.ReadAll(io.LimitReader(input, maxInputBytes+1))
	if err != nil {
		return EvaluationRequest{}, fieldError("$", "proveer una entrada legible")
	}
	if len(body) > maxInputBytes {
		return EvaluationRequest{}, fieldError("$", "reducir el documento a 64 KiB o menos")
	}
	if !utf8.Valid(body) {
		return EvaluationRequest{}, fieldError("$", "usar UTF-8 válido")
	}
	if err := rejectUnsafeYAML(body); err != nil {
		return EvaluationRequest{}, err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(body))
	decoder.KnownFields(true)
	var request EvaluationRequest
	if err := decoder.Decode(&request); err != nil {
		return EvaluationRequest{}, fieldError("$", "corregir sintaxis, claves duplicadas o campos no documentados")
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); err != io.EOF {
		return EvaluationRequest{}, fieldError("$", "proveer exactamente un documento")
	}
	if err := ValidateEvaluation(request); err != nil {
		return EvaluationRequest{}, err
	}
	return request, nil
}

func rejectUnsafeYAML(body []byte) error {
	var document yaml.Node
	if err := yaml.Unmarshal(body, &document); err != nil {
		return fieldError("$", "corregir la sintaxis YAML o JSON")
	}
	var walk func(*yaml.Node) error
	walk = func(node *yaml.Node) error {
		if node == nil {
			return fieldError("$", "proveer un objeto")
		}
		if node.Kind == yaml.AliasNode || node.Anchor != "" {
			return fieldError("$", "no usar anchors ni aliases")
		}
		switch node.Tag {
		case "", "!!map", "!!seq", "!!str", "!!bool", "!!int", "!!null":
		default:
			return fieldError("$", "usar tipos YAML o JSON estándar")
		}
		for _, child := range node.Content {
			if err := walk(child); err != nil {
				return err
			}
		}
		return nil
	}
	if len(document.Content) != 1 {
		return fieldError("$", "proveer exactamente un documento")
	}
	if err := walk(document.Content[0]); err != nil {
		return err
	}
	return nil
}
