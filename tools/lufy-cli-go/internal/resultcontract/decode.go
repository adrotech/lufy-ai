package resultcontract

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

const maxInputBytes = 64 * 1024

var (
	safePathSegment  = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
	duplicateYAMLKey = regexp.MustCompile(`mapping key "([a-z][a-z0-9_]{0,63})" already defined`)
)

func Decode(input io.Reader) (Contract, error) {
	root, err := decodeDocument(input)
	if err != nil {
		return Contract{}, err
	}
	if err := validateNodeShape(root, reflect.TypeOf(Contract{}), ""); err != nil {
		return Contract{}, err
	}

	var contract Contract
	if err := root.Decode(&contract); err != nil {
		if typed, ok := err.(*DiagnosticError); ok {
			return Contract{}, typed
		}
		return Contract{}, diagnostic("decode_failed", "$", "corregir tipos escalares y listas según result-contract/v1")
	}
	if err := Validate(contract); err != nil {
		return Contract{}, err
	}
	return contract, nil
}

func decodeDocument(input io.Reader) (*yaml.Node, error) {
	if input == nil {
		return nil, diagnostic("read_failed", "$", "proveer una entrada YAML o JSON")
	}
	body, err := io.ReadAll(io.LimitReader(input, maxInputBytes+1))
	if err != nil {
		return nil, diagnostic("read_failed", "$", "reintentar con una entrada local legible")
	}
	if len(body) > maxInputBytes {
		return nil, diagnostic("input_too_large", "$", "reducir el documento a 64 KiB o menos")
	}
	if !utf8.Valid(body) {
		return nil, diagnostic("invalid_utf8", "$", "codificar el documento como UTF-8 válido")
	}

	decoder := yaml.NewDecoder(bytes.NewReader(body))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return nil, syntaxDiagnostic(err)
	}
	if len(document.Content) != 1 {
		return nil, diagnostic("invalid_document", "$", "proveer un único objeto YAML o JSON")
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, diagnostic("multiple_documents", "$", "proveer exactamente un documento")
	}

	root := document.Content[0]
	if err := rejectUnsafeYAML(root, "$"); err != nil {
		return nil, err
	}
	return root, nil
}

func syntaxDiagnostic(err error) error {
	match := duplicateYAMLKey.FindStringSubmatch(err.Error())
	if len(match) == 2 {
		return diagnostic("duplicate_key", match[1], "eliminar la clave duplicada")
	}
	return diagnostic("invalid_syntax", "$", "corregir la sintaxis YAML o JSON")
}

func rejectUnsafeYAML(node *yaml.Node, path string) error {
	if node == nil {
		return diagnostic("invalid_document", path, "proveer un objeto YAML o JSON")
	}
	if node.Kind == yaml.AliasNode || node.Anchor != "" {
		return diagnostic("yaml_reference_forbidden", path, "repetir el valor explícitamente sin anchors ni aliases")
	}
	if !standardYAMLTag(node.Tag) {
		return diagnostic("yaml_tag_forbidden", path, "usar solo tipos escalares YAML/JSON estándar")
	}
	for _, child := range node.Content {
		if err := rejectUnsafeYAML(child, path); err != nil {
			return err
		}
	}
	return nil
}

func validateNodeShape(node *yaml.Node, typ reflect.Type, path string) error {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	switch typ.Kind() {
	case reflect.Struct:
		if typ == reflect.TypeOf(time.Time{}) {
			if node.Kind != yaml.ScalarNode || node.Tag != "!!str" && node.Tag != "!!timestamp" {
				return diagnostic("invalid_type", displayPath(path), "usar un timestamp RFC3339")
			}
			return nil
		}
		if node.Kind != yaml.MappingNode {
			return diagnostic("invalid_type", displayPath(path), "usar un objeto con campos permitidos")
		}
		fields := allowedFields(typ)
		seen := make(map[string]bool, len(node.Content)/2)
		for index := 0; index+1 < len(node.Content); index += 2 {
			keyNode, valueNode := node.Content[index], node.Content[index+1]
			if keyNode.Kind != yaml.ScalarNode || !safePathSegment.MatchString(keyNode.Value) {
				return diagnostic("invalid_key", displayPath(path), "usar claves documentadas en minúsculas")
			}
			key := keyNode.Value
			childPath := joinPath(path, key)
			if seen[key] {
				return diagnostic("duplicate_key", childPath, "eliminar la clave duplicada")
			}
			seen[key] = true
			fieldType, ok := fields[key]
			if !ok {
				return diagnostic("unknown_field", childPath, "eliminar el campo no documentado")
			}
			if err := validateNodeShape(valueNode, fieldType, childPath); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		if node.Kind != yaml.SequenceNode {
			return diagnostic("invalid_type", displayPath(path), "usar una lista")
		}
		for _, child := range node.Content {
			if err := validateNodeShape(child, typ.Elem(), path); err != nil {
				return err
			}
		}
		return nil
	case reflect.String:
		if node.Kind != yaml.ScalarNode {
			return diagnostic("invalid_type", displayPath(path), "usar un valor escalar")
		}
		if typ == reflect.TypeOf(BoolValue("")) {
			if node.Tag != "!!bool" && node.Tag != "!!str" {
				return diagnostic("invalid_type", displayPath(path), "usar true, false o un marcador permitido")
			}
			return nil
		}
		if typ == reflect.TypeOf(CountValue("")) {
			if node.Tag != "!!int" && node.Tag != "!!str" {
				return diagnostic("invalid_type", displayPath(path), "usar 1, 2 o not_applicable")
			}
			return nil
		}
		if node.Tag != "!!str" {
			return diagnostic("invalid_type", displayPath(path), "usar texto")
		}
		return nil
	case reflect.Bool:
		if node.Kind != yaml.ScalarNode || node.Tag != "!!bool" {
			return diagnostic("invalid_type", displayPath(path), "usar true o false")
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if node.Kind != yaml.ScalarNode || node.Tag != "!!int" {
			return diagnostic("invalid_type", displayPath(path), "usar un entero no negativo")
		}
		return nil
	default:
		return diagnostic("unsupported_type", displayPath(path), "usar un campo documentado")
	}
}

func standardYAMLTag(tag string) bool {
	switch tag {
	case "", "!!map", "!!seq", "!!str", "!!bool", "!!int", "!!null", "!!timestamp":
		return true
	default:
		return false
	}
}

func allowedFields(typ reflect.Type) map[string]reflect.Type {
	fields := make(map[string]reflect.Type, typ.NumField())
	for index := 0; index < typ.NumField(); index++ {
		field := typ.Field(index)
		name := strings.Split(field.Tag.Get("yaml"), ",")[0]
		if name != "" && name != "-" {
			fields[name] = field.Type
		}
	}
	return fields
}

func joinPath(parent, child string) string {
	if parent == "" || parent == "$" {
		return child
	}
	return parent + "." + child
}

func displayPath(path string) string {
	if path == "" {
		return "$"
	}
	return path
}

func required(path string) error {
	return diagnostic("required_field", path, fmt.Sprintf("agregar el campo obligatorio %s", path))
}
