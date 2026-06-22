package versioncheck

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/version"
)

const defaultLatestReleaseURL = "https://api.github.com/repos/adrotech/lufy-ai/releases/latest"

type Options struct {
	LatestReleaseURL string
	HTTPClient       *http.Client
	Current          version.Info
}

type Result struct {
	Checked         bool   `json:"checked"`
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion,omitempty"`
	UpdateAvailable bool   `json:"updateAvailable"`
	UpToDate        bool   `json:"upToDate"`
	DevBuild        bool   `json:"devBuild"`
	Error           string `json:"error,omitempty"`
	Recommendation  string `json:"recommendation,omitempty"`
}

type Service struct{}

func NewService() Service {
	return Service{}
}

func (s Service) Check(opts Options) Result {
	current := opts.Current
	if current.Version == "" {
		current = version.Current()
	}
	result := Result{Checked: true, CurrentVersion: current.Version, DevBuild: current.DevBuild}
	latest, err := latestRelease(opts)
	if err != nil {
		result.Error = err.Error()
		result.Recommendation = "No se pudo verificar la ultima version; continua con setup o reintenta con red disponible."
		return result
	}
	result.LatestVersion = latest
	cmp, comparable := Compare(current.Version, latest)
	if !comparable || current.DevBuild {
		result.Recommendation = fmt.Sprintf("Build local no comparable; ultima version estable disponible: %s", latest)
		return result
	}
	if cmp < 0 {
		result.UpdateAvailable = true
		result.Recommendation = fmt.Sprintf("Ejecuta lufy-ai upgrade --to %s antes de continuar si quieres usar la ultima version.", latest)
		return result
	}
	result.UpToDate = true
	result.Recommendation = "Lufy AI esta al dia."
	return result
}

func latestRelease(opts Options) (string, error) {
	url := opts.LatestReleaseURL
	if url == "" {
		url = defaultLatestReleaseURL
	}
	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("consulta de release fallo: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var decoded struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return "", err
	}
	if strings.TrimSpace(decoded.TagName) == "" {
		return "", fmt.Errorf("release latest no contiene tag_name")
	}
	return strings.TrimSpace(decoded.TagName), nil
}

func Compare(current, latest string) (int, bool) {
	a, okA := parseSemver(current)
	b, okB := parseSemver(latest)
	if !okA || !okB {
		return 0, false
	}
	for i := 0; i < 3; i++ {
		if a[i] < b[i] {
			return -1, true
		}
		if a[i] > b[i] {
			return 1, true
		}
	}
	return 0, true
}

func parseSemver(value string) ([3]int, bool) {
	var out [3]int
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "v")
	if cut := strings.IndexAny(value, "+-"); cut >= 0 {
		value = value[:cut]
	}
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}
