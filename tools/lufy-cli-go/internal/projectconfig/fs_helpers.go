package projectconfig

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func statPath(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func exists(root, rel string) bool {
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil
}

func existsAny(root string, rels ...string) bool {
	for _, rel := range rels {
		if exists(root, rel) {
			return true
		}
	}
	return false
}

func existsGlob(root, pattern string) bool {
	matches, _ := filepath.Glob(filepath.Join(root, pattern))
	return len(matches) > 0
}

func hasDep(deps map[string]string, name string) bool { _, ok := deps[name]; return ok }

func hasDepPrefix(deps map[string]string, prefix string) bool {
	for name := range deps {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func hasAny(values, candidates []string) bool {
	for _, candidate := range candidates {
		if containsString(values, candidate) {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func jsPackageManager(root string) string {
	switch {
	case exists(root, "pnpm-lock.yaml"):
		return "pnpm"
	case exists(root, "yarn.lock"):
		return "yarn"
	case exists(root, "package-lock.json"):
		return "npm"
	default:
		return "npm"
	}
}

func goLinter(root string) string {
	if existsAny(root, ".golangci.yml", ".golangci.yaml") {
		return "golangci-lint run"
	}
	return "go vet ./..."
}

func goLinterFix(root string) string {
	if existsAny(root, ".golangci.yml", ".golangci.yaml") {
		return "golangci-lint run --fix"
	}
	return ""
}

func readKnown(root string, names ...string) string {
	var buf strings.Builder
	for _, name := range names {
		if data, err := os.ReadFile(filepath.Join(root, name)); err == nil {
			buf.Write(data)
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

func containsAny(text string, candidates []string) []string {
	var out []string
	for _, candidate := range candidates {
		if strings.Contains(text, candidate) {
			out = append(out, candidate)
		}
	}
	return out
}

func detectInFiles(root string, files, candidates []string) []string {
	return containsAny(readKnown(root, files...), candidates)
}

func detectPackageLibs(deps map[string]string, candidates []string) []string {
	var out []string
	for _, candidate := range candidates {
		if hasDep(deps, candidate) {
			out = append(out, candidate)
		}
	}
	return out
}

func unique(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func upsertStack(stacks []Stack, stack Stack) []Stack {
	for i := range stacks {
		if stacks[i].ID == stack.ID {
			stacks[i].Frameworks = unique(append(stacks[i].Frameworks, stack.Frameworks...))
			stacks[i].ObservabilityLibs = unique(append(stacks[i].ObservabilityLibs, stack.ObservabilityLibs...))
			return stacks
		}
	}
	return append(stacks, stack)
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
