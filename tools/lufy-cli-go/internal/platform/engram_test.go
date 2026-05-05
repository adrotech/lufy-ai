package platform

import (
	"errors"
	"testing"
)

type fakeResolver struct {
	path string
	err  error
}

func (f fakeResolver) LookPath(string) (string, error) {
	return f.path, f.err
}

func TestResolveEngramFound(t *testing.T) {
	path, ok := ResolveEngram(false, fakeResolver{path: "/usr/local/bin/engram"})
	if !ok || path == "" {
		t.Fatalf("expected engram to be found")
	}
}

func TestResolveEngramMissing(t *testing.T) {
	path, ok := ResolveEngram(false, fakeResolver{err: errors.New("missing")})
	if ok || path != "" {
		t.Fatalf("expected engram to be missing")
	}
}

func TestResolveEngramNoEngramFlag(t *testing.T) {
	path, ok := ResolveEngram(true, fakeResolver{path: "/usr/local/bin/engram"})
	if ok || path != "" {
		t.Fatalf("expected no engram resolution when flag is set")
	}
}
