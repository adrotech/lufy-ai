package opsx

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var versionPattern = regexp.MustCompile(`\d+(?:\.\d+){0,2}`)

func ParseVersion(text string) (string, error) {
	match := versionPattern.FindString(text)
	if match == "" {
		return "", fmt.Errorf("no se pudo detectar versión OpenSpec en %q", strings.TrimSpace(text))
	}
	return normalizeVersion(match)
}

func CompatibleVersion(got, minimum string) bool {
	gotNorm, err := normalizeVersion(got)
	if err != nil {
		return false
	}
	minNorm, err := normalizeVersion(minimum)
	if err != nil {
		return false
	}
	return compareVersion(gotNorm, minNorm) >= 0
}

func normalizeVersion(value string) (string, error) {
	parts := strings.Split(strings.TrimSpace(value), ".")
	if len(parts) == 0 || len(parts) > 3 {
		return "", fmt.Errorf("versión OpenSpec inválida: %q", value)
	}
	for len(parts) < 3 {
		parts = append(parts, "0")
	}
	for _, part := range parts {
		if part == "" {
			return "", fmt.Errorf("versión OpenSpec inválida: %q", value)
		}
		if _, err := strconv.Atoi(part); err != nil {
			return "", fmt.Errorf("versión OpenSpec inválida: %q", value)
		}
	}
	return strings.Join(parts, "."), nil
}

func compareVersion(left, right string) int {
	l := strings.Split(left, ".")
	r := strings.Split(right, ".")
	for i := 0; i < 3; i++ {
		li, _ := strconv.Atoi(l[i])
		ri, _ := strconv.Atoi(r[i])
		if li > ri {
			return 1
		}
		if li < ri {
			return -1
		}
	}
	return 0
}
