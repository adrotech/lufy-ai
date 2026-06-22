package versioncheck

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/version"
)

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    int
		ok      bool
	}{
		{current: "v1.2.3", latest: "v1.2.4", want: -1, ok: true},
		{current: "v1.3.0", latest: "v1.2.4", want: 1, ok: true},
		{current: "1.2.3", latest: "v1.2.3", want: 0, ok: true},
		{current: "dev", latest: "v1.2.3", ok: false},
	}
	for _, tt := range tests {
		got, ok := Compare(tt.current, tt.latest)
		if ok != tt.ok || got != tt.want {
			t.Fatalf("Compare(%q,%q) = %d,%t want %d,%t", tt.current, tt.latest, got, ok, tt.want, tt.ok)
		}
	}
}

func TestCheckReportsUpdateAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.2.4"}`))
	}))
	defer server.Close()

	result := NewService().Check(Options{LatestReleaseURL: server.URL, Current: version.Info{Version: "v1.2.3"}})
	if !result.Checked || !result.UpdateAvailable || result.LatestVersion != "v1.2.4" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.Recommendation == "" {
		t.Fatalf("missing recommendation: %#v", result)
	}
}

func TestCheckReportsUpToDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.2.3"}`))
	}))
	defer server.Close()

	result := NewService().Check(Options{LatestReleaseURL: server.URL, Current: version.Info{Version: "v1.2.3"}})
	if !result.UpToDate || result.UpdateAvailable {
		t.Fatalf("unexpected result: %#v", result)
	}
}
