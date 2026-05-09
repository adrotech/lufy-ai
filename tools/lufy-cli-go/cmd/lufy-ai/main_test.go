package main

import (
	"bytes"
	"testing"
)

func TestRunDispatchesVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(version) code=%d stderr=%s", code, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("lufy-ai")) || !bytes.Contains(stdout.Bytes(), []byte("commit:")) {
		t.Fatalf("version output missing runtime metadata: %s", stdout.String())
	}
}
