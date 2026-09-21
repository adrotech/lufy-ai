package resultcontract

import "fmt"

type DiagnosticError struct {
	Code     string `json:"code"`
	Path     string `json:"path"`
	Recovery string `json:"recovery"`
}

func (e *DiagnosticError) Error() string {
	if e == nil {
		return "result contract inválido"
	}
	return fmt.Sprintf("result contract inválido: code=%s path=%s recovery=%s", e.Code, e.Path, e.Recovery)
}

func diagnostic(code, path, recovery string) error {
	if path == "" {
		path = "$"
	}
	return &DiagnosticError{Code: code, Path: path, Recovery: recovery}
}
