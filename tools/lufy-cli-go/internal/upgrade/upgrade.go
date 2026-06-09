package upgrade

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

const defaultBaseURL = "https://github.com/adrotech/lufy-ai/releases/download"
const defaultFetchAttempts = 3

var defaultHTTPClient = &http.Client{Timeout: 120 * time.Second}

type Options struct {
	To       string
	BaseURL  string
	ExecPath string
	DryRun   bool
}

type Service struct{}

func NewService() Service {
	return Service{}
}

func (s Service) Run(opts Options, stdout io.Writer) error {
	if opts.To == "" {
		return fmt.Errorf("upgrade requiere --to <vX.Y.Z>")
	}
	if opts.To == "latest" {
		return fmt.Errorf("upgrade no acepta latest; usa una versión fija reproducible")
	}
	if !strings.HasPrefix(opts.To, "v") {
		return fmt.Errorf("upgrade requiere versión con prefijo v*: %s", opts.To)
	}
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	execPath := opts.ExecPath
	if execPath == "" {
		var err error
		execPath, err = os.Executable()
		if err != nil {
			return err
		}
	}
	execPath, err := filepath.Abs(execPath)
	if err != nil {
		return err
	}
	artifact := artifactName(opts.To, runtime.GOOS, runtime.GOARCH)
	checksums := fmt.Sprintf("lufy-ai_%s_checksums.txt", opts.To)
	releaseURL := strings.TrimRight(baseURL, "/") + "/" + opts.To
	artifactURL := releaseURL + "/" + artifact
	checksumsURL := releaseURL + "/" + checksums
	fmt.Fprintf(stdout, "Upgrade a %s\n", opts.To)
	fmt.Fprintf(stdout, "Artifact: %s\n", artifactURL)
	fmt.Fprintf(stdout, "Destino: %s\n", execPath)
	if opts.DryRun {
		fmt.Fprintln(stdout, "Modo dry-run: no se descargó ni reemplazó el binario")
		return nil
	}
	checksumsBody, err := fetchURL(checksumsURL)
	if err != nil {
		return err
	}
	artifactBody, err := fetchURL(artifactURL)
	if err != nil {
		return err
	}
	if err := verifyChecksum(artifact, artifactBody, checksumsBody); err != nil {
		return err
	}
	binary, err := extractBinary(artifact, artifactBody)
	if err != nil {
		return err
	}
	if err := platform.WriteFileAtomic(execPath, binary, 0o755); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "lufy-ai actualizado en %s\n", execPath)
	return nil
}

func artifactName(version, goos, goarch string) string {
	name := fmt.Sprintf("lufy-ai_%s_%s_%s", version, goos, goarch)
	if goos == "windows" {
		return name + ".zip"
	}
	return name + ".tar.gz"
}

func fetchURL(rawURL string) ([]byte, error) {
	return fetchURLWithClient(rawURL, defaultHTTPClient, defaultFetchAttempts)
}

func fetchURLWithClient(rawURL string, client *http.Client, attempts int) ([]byte, error) {
	if strings.HasPrefix(rawURL, "file://") {
		return os.ReadFile(strings.TrimPrefix(rawURL, "file://"))
	}
	if client == nil {
		client = defaultHTTPClient
	}
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		resp, err := client.Get(rawURL)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("descarga falló %s: HTTP %d", rawURL, resp.StatusCode)
			_ = resp.Body.Close()
			if !retryableHTTPStatus(resp.StatusCode) {
				break
			}
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if closeErr != nil {
			lastErr = closeErr
			continue
		}
		return body, nil
	}
	return nil, lastErr
}

func retryableHTTPStatus(status int) bool {
	return status == http.StatusRequestTimeout || status == http.StatusTooManyRequests || status >= 500
}

func verifyChecksum(artifact string, artifactBody, checksumsBody []byte) error {
	expected := ""
	for _, line := range strings.Split(string(checksumsBody), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == artifact {
			expected = fields[0]
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("checksums no contiene entrada para %s", artifact)
	}
	actualBytes := sha256.Sum256(artifactBody)
	actual := hex.EncodeToString(actualBytes[:])
	if actual != expected {
		return fmt.Errorf("checksum mismatch para %s", artifact)
	}
	return nil
}

func extractBinary(artifact string, body []byte) ([]byte, error) {
	if strings.HasSuffix(artifact, ".zip") {
		return extractZipBinary(body)
	}
	return extractTarBinary(body)
}

func extractTarBinary(body []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(header.Name) == "lufy-ai" {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("artifact no contiene binario lufy-ai")
}

func extractZipBinary(body []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, err
	}
	for _, file := range zr.File {
		if file.FileInfo().IsDir() || filepath.Base(file.Name) != "lufy-ai.exe" {
			continue
		}
		r, err := file.Open()
		if err != nil {
			return nil, err
		}
		body, readErr := io.ReadAll(r)
		closeErr := r.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		return body, nil
	}
	return nil, fmt.Errorf("artifact no contiene binario lufy-ai.exe")
}
