package codexlifecycle

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/resultcontract"
)

const maxResultContractMessageBytes = 64 * 1024

func extractResultContract(message string) (resultcontract.Contract, error) {
	if len(message) == 0 || len(message) > maxResultContractMessageBytes {
		return resultcontract.Contract{}, fmt.Errorf("result contract ausente o fuera de límite")
	}
	if contract, err := resultcontract.Decode(bytes.NewBufferString(message)); err == nil {
		return contract, nil
	}

	candidate, err := singleCompatibleFence(message)
	if err != nil {
		return resultcontract.Contract{}, err
	}
	contract, err := resultcontract.Decode(bytes.NewBufferString(candidate))
	if err != nil {
		return resultcontract.Contract{}, fmt.Errorf("result contract fenced inválido")
	}
	return contract, nil
}

func singleCompatibleFence(message string) (string, error) {
	normalized := strings.ReplaceAll(strings.ReplaceAll(message, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")
	var candidate strings.Builder
	found := 0
	inside := false
	closed := false
	for _, line := range lines {
		marker := strings.TrimSpace(line)
		if !inside {
			switch marker {
			case "```yaml", "```json":
				if found != 0 {
					return "", fmt.Errorf("result contract ambiguo")
				}
				found++
				inside = true
				closed = false
			default:
				if strings.HasPrefix(marker, "```") {
					return "", fmt.Errorf("fence no compatible")
				}
			}
			continue
		}
		if marker == "```" {
			inside = false
			closed = true
			continue
		}
		if strings.HasPrefix(marker, "```") {
			return "", fmt.Errorf("fence anidado no permitido")
		}
		candidate.WriteString(line)
		candidate.WriteByte('\n')
	}
	if found != 1 || inside || !closed {
		return "", fmt.Errorf("se requiere un único fence yaml o json cerrado")
	}
	return candidate.String(), nil
}
