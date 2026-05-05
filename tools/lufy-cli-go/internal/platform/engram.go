package platform

import "os/exec"

type CommandResolver interface {
	LookPath(file string) (string, error)
}

type OSResolver struct{}

func (OSResolver) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func ResolveEngram(noEngram bool, resolver CommandResolver) (string, bool) {
	if noEngram {
		return "", false
	}
	path, err := resolver.LookPath("engram")
	if err != nil {
		return "", false
	}
	return path, true
}
