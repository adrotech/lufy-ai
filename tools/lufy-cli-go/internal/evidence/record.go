// Package evidence interpreta comprobaciones declaradas para integraciones opcionales.
// No ejecuta comprobaciones, concede permisos ni mantiene estados de workflow.
package evidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

type Check struct {
	Name     string `json:"name"`
	Result   string `json:"result"`
	Evidence string `json:"evidence,omitempty"`
	Reason   string `json:"reason,omitempty"`
}
type Finding struct {
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
}
type Record struct {
	Checks   []Check   `json:"checks"`
	Findings []Finding `json:"findings,omitempty"`
}
type Assessment struct {
	Conclusion       string `json:"conclusion"`
	Passed           int    `json:"passed"`
	Failed           int    `json:"failed"`
	Inconclusive     int    `json:"inconclusive"`
	NotApplicable    int    `json:"not_applicable"`
	BlockingFindings int    `json:"blocking_findings"`
}

func Decode(input io.Reader) (Record, error) {
	body, err := io.ReadAll(io.LimitReader(input, 128*1024+1))
	if err != nil || len(body) > 128*1024 {
		return Record{}, fmt.Errorf("evidencia no legible o fuera de límite")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := uniqueValue(decoder); err != nil {
		return Record{}, fmt.Errorf("evidencia JSON inválida o ambigua")
	}
	if _, err := decoder.Token(); err != io.EOF {
		return Record{}, fmt.Errorf("evidencia requiere un único objeto JSON")
	}
	decoder = json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var record Record
	if err := decoder.Decode(&record); err != nil {
		return Record{}, fmt.Errorf("campos de evidencia inválidos; usa checks y findings")
	}
	if len(record.Checks) == 0 || len(record.Checks) > 128 || len(record.Findings) > 128 {
		return Record{}, fmt.Errorf("evidencia requiere comprobaciones y listas acotadas")
	}
	seen := map[string]bool{}
	for i := range record.Checks {
		c := &record.Checks[i]
		c.Name = strings.TrimSpace(c.Name)
		c.Evidence = strings.TrimSpace(c.Evidence)
		c.Reason = strings.TrimSpace(c.Reason)
		key := strings.ToLower(c.Name)
		if key == "" || len(c.Name) > 256 || seen[key] || len(c.Evidence) > 4096 || len(c.Reason) > 4096 {
			return Record{}, fmt.Errorf("comprobación vacía, duplicada o fuera de límite")
		}
		seen[key] = true
		switch c.Result {
		case "passed", "failed", "inconclusive":
			if c.Evidence == "" {
				return Record{}, fmt.Errorf("la comprobación necesita evidencia observada")
			}
		case "not_applicable":
			if c.Reason == "" {
				return Record{}, fmt.Errorf("una comprobación no aplicable necesita motivo")
			}
		default:
			return Record{}, fmt.Errorf("resultado de comprobación no soportado")
		}
	}
	for i := range record.Findings {
		f := &record.Findings[i]
		f.Summary = strings.TrimSpace(f.Summary)
		if f.Summary == "" || len(f.Summary) > 4096 {
			return Record{}, fmt.Errorf("hallazgo vacío o fuera de límite")
		}
		switch f.Severity {
		case "critical", "high", "medium", "low":
		default:
			return Record{}, fmt.Errorf("severidad de hallazgo no soportada")
		}
	}
	return record, nil
}

// uniqueValue rechaza keys duplicadas también dentro de checks y findings.
func uniqueValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || name != strings.ToLower(name) || seen[name] {
				return fmt.Errorf("key duplicada")
			}
			seen[name] = true
			if err := uniqueValue(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := uniqueValue(decoder); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("delimitador inválido")
	}
	_, err = decoder.Token()
	return err
}

func Assess(record Record) Assessment {
	result := Assessment{Conclusion: "not_applicable"}
	for _, check := range record.Checks {
		switch check.Result {
		case "passed":
			result.Passed++
		case "failed":
			result.Failed++
		case "inconclusive":
			result.Inconclusive++
		case "not_applicable":
			result.NotApplicable++
		}
	}
	for _, finding := range record.Findings {
		if finding.Severity == "critical" || finding.Severity == "high" {
			result.BlockingFindings++
		}
	}
	switch {
	case result.Failed > 0 || result.BlockingFindings > 0:
		result.Conclusion = "failed"
	case result.Inconclusive > 0:
		result.Conclusion = "inconclusive"
	case result.Passed > 0:
		result.Conclusion = "supported"
	}
	return result
}

// Fingerprint incluye observaciones y hallazgos; formato y orden no alteran identidad.
func Fingerprint(record Record) string {
	record.Checks = append([]Check(nil), record.Checks...)
	record.Findings = append([]Finding(nil), record.Findings...)
	sort.Slice(record.Checks, func(i, j int) bool { return record.Checks[i].Name < record.Checks[j].Name })
	sort.Slice(record.Findings, func(i, j int) bool {
		a, b := record.Findings[i], record.Findings[j]
		if a.Severity != b.Severity {
			return a.Severity < b.Severity
		}
		return a.Summary < b.Summary
	})
	body, _ := json.Marshal(record)
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
