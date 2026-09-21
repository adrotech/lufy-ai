package resultcontract

import (
	"bytes"
	"encoding/json"

	"gopkg.in/yaml.v3"
)

func (v *BoolValue) UnmarshalJSON(body []byte) error {
	if bytes.Equal(body, []byte("true")) {
		*v = BoolTrue
		return nil
	}
	if bytes.Equal(body, []byte("false")) {
		*v = BoolFalse
		return nil
	}
	var marker string
	if err := json.Unmarshal(body, &marker); err != nil {
		return diagnostic("invalid_type", "boolean", "use true, false o un marcador permitido")
	}
	*v = BoolValue(marker)
	return nil
}

func (v *BoolValue) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return diagnostic("invalid_type", "boolean", "use true, false o un marcador permitido")
	}
	*v = BoolValue(node.Value)
	return nil
}

func (v BoolValue) MarshalJSON() ([]byte, error) {
	switch v {
	case BoolTrue:
		return []byte("true"), nil
	case BoolFalse:
		return []byte("false"), nil
	default:
		return json.Marshal(string(v))
	}
}

func (v *CountValue) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return diagnostic("invalid_type", "candidate_count", "use 1, 2 o not_applicable")
	}
	*v = CountValue(node.Value)
	return nil
}

func (v CountValue) MarshalJSON() ([]byte, error) {
	switch v {
	case CountOne, CountTwo:
		return []byte(v), nil
	default:
		return json.Marshal(string(v))
	}
}

func (v *CountValue) UnmarshalJSON(body []byte) error {
	if bytes.Equal(body, []byte("1")) || bytes.Equal(body, []byte("2")) {
		*v = CountValue(string(body))
		return nil
	}
	var marker string
	if err := json.Unmarshal(body, &marker); err != nil {
		return diagnostic("invalid_type", "candidate_count", "use 1, 2 o not_applicable")
	}
	*v = CountValue(marker)
	return nil
}
