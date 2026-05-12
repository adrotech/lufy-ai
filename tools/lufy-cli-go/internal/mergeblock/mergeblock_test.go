package mergeblock

import (
	"strings"
	"testing"
)

func TestRenderReplacesOnlyManagedBlock(t *testing.T) {
	target := []byte("intro\n<!-- LUFY:BEGIN core -->\nold\n<!-- LUFY:END core -->\noutro\n")
	source := []byte("<!-- LUFY:BEGIN core -->\nnew\n<!-- LUFY:END core -->\n")
	got, err := Render(target, source)
	if err != nil {
		t.Fatal(err)
	}
	want := "intro\n<!-- LUFY:BEGIN core -->\nnew\n<!-- LUFY:END core -->\noutro\n"
	if string(got) != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderAppendsMissingBlock(t *testing.T) {
	target := []byte("user text\n")
	source := []byte("<!-- LUFY:BEGIN core -->\nmanaged\n<!-- LUFY:END core -->\n")
	got, err := Render(target, source)
	if err != nil {
		t.Fatal(err)
	}
	want := "user text\n\n<!-- LUFY:BEGIN core -->\nmanaged\n<!-- LUFY:END core -->\n"
	if string(got) != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderRejectsCorruptMarkers(t *testing.T) {
	source := []byte("<!-- LUFY:BEGIN core -->\nmanaged\n<!-- LUFY:END core -->\n")
	cases := map[string]string{
		"duplicate": "<!-- LUFY:BEGIN core -->\n1\n<!-- LUFY:END core -->\n<!-- LUFY:BEGIN core -->\n2\n<!-- LUFY:END core -->\n",
		"nested":    "<!-- LUFY:BEGIN core -->\n<!-- LUFY:BEGIN other -->\n<!-- LUFY:END other -->\n<!-- LUFY:END core -->\n",
		"unclosed":  "<!-- LUFY:BEGIN core -->\n",
	}
	for name, target := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Render([]byte(target), source)
			if err == nil || !strings.Contains(err.Error(), "LUFY") {
				t.Fatalf("expected marker error, got %v", err)
			}
		})
	}
}
