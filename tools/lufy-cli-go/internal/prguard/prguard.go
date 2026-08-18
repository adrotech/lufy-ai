package prguard

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

var defaultInternalPrefixes = []string{"openspec/", ".lufy/", ".lufy-ai/", "pr_review/"}

var userOwnedLufySDDPrefixes = []string{
	".lufy/sdd/",
	".lufy/workflows/sdd/changes/",
	".lufy/workflows/sdd/specs/",
	".lufy/workflows/sdd/decisions/",
	".lufy/workflows/sdd/verification/",
	".lufy/workflows/sdd/archive/",
}

type Options struct {
	Target          string
	Base            string
	JSON            bool
	IncludeWorktree bool
}

type Report struct {
	OK              bool        `json:"ok"`
	TargetRoot      string      `json:"targetRoot"`
	Base            string      `json:"base"`
	Range           string      `json:"range"`
	ChangedFiles    []string    `json:"changedFiles"`
	IgnoredMatches  []Violation `json:"ignoredMatches,omitempty"`
	InternalMatches []Violation `json:"internalMatches,omitempty"`
	Remediation     []string    `json:"remediation,omitempty"`
}

type Violation struct {
	Path    string `json:"path"`
	Kind    string `json:"kind"`
	Source  string `json:"source,omitempty"`
	Line    string `json:"line,omitempty"`
	Pattern string `json:"pattern"`
}

type changedFile struct {
	Path   string
	Status string
}

type Service struct{}

func NewService() Service { return Service{} }

func (s Service) Run(opts Options, stdout io.Writer) error {
	report, err := s.Build(opts)
	if err != nil {
		return err
	}
	if opts.JSON {
		body, jsonErr := json.MarshalIndent(report, "", "  ")
		if jsonErr != nil {
			return jsonErr
		}
		fmt.Fprintln(stdout, string(body))
	} else {
		present(report, stdout)
	}
	if !report.OK {
		return errors.New("pr guard detectó paths ignorados o internos en el diff")
	}
	return nil
}

func (s Service) Build(opts Options) (Report, error) {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return Report{}, err
	}
	base := opts.Base
	if base == "" {
		base = "origin/develop"
	}
	entries, diffRange, err := changedFiles(target, base, opts.IncludeWorktree)
	if err != nil {
		return Report{}, err
	}
	files := changedFilePaths(entries)
	ignored, err := ignoredFiles(target, files)
	if err != nil {
		return Report{}, err
	}
	internal := internalFileEntries(entries)
	report := Report{OK: len(ignored) == 0 && len(internal) == 0, TargetRoot: target, Base: base, Range: diffRange, ChangedFiles: files, IgnoredMatches: ignored, InternalMatches: internal}
	if !report.OK {
		report.Remediation = []string{
			"Revisar si el path debe versionarse; si sí, registrar override explícito en el resumen/PR.",
			"Si no debe entrar al PR, usar git rm --cached <path> para mantener el archivo local y sacarlo del índice, o git rm <path> si también debe borrarse localmente.",
			"Crear un commit correctivo que remueva solo esos paths y reejecutar lufy-ai pr guard.",
		}
	}
	return report, nil
}

func changedFiles(target, base string, includeWorktree bool) ([]changedFile, string, error) {
	diffRange := base + "...HEAD"
	args := []string{"diff", "--name-status", diffRange, "--"}
	if includeWorktree {
		diffRange = base
		args = []string{"diff", "--name-status", base, "--"}
	}
	stdout, stderr, err := runGit(target, args...)
	if err != nil {
		return nil, "", fmt.Errorf("git diff --name-status %s -- fallo: %s", diffRange, strings.TrimSpace(stderr))
	}
	files := parseNameStatus(stdout)
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, diffRange, nil
}

func ignoredFiles(target string, files []string) ([]Violation, error) {
	if len(files) == 0 {
		return nil, nil
	}
	cmd := exec.Command("git", "-C", target, "check-ignore", "-v", "--no-index", "--stdin")
	cmd.Stdin = strings.NewReader(strings.Join(files, "\n") + "\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("git check-ignore -v --no-index --stdin fallo: %s", strings.TrimSpace(stderr.String()))
	}
	return parseCheckIgnore(stdout.String()), nil
}

func internalFiles(files []string) []Violation {
	entries := make([]changedFile, 0, len(files))
	for _, file := range files {
		entries = append(entries, changedFile{Path: file, Status: "M"})
	}
	return internalFileEntries(entries)
}

func internalFileEntries(files []changedFile) []Violation {
	var out []Violation
	for _, file := range files {
		slash := filepath.ToSlash(file.Path)
		if isAllowedOpenSpecPath(changedFile{Path: slash, Status: file.Status}) {
			continue
		}
		if hasAnyPrefix(slash, userOwnedLufySDDPrefixes) {
			continue
		}
		for _, prefix := range defaultInternalPrefixes {
			if strings.HasPrefix(slash, prefix) {
				out = append(out, Violation{Path: slash, Kind: "internal", Pattern: prefix})
				break
			}
		}
	}
	return out
}

func isAllowedOpenSpecPath(file changedFile) bool {
	slash := filepath.ToSlash(file.Path)
	switch {
	case strings.HasPrefix(slash, "openspec/specs/"):
		return true
	case strings.HasPrefix(slash, "openspec/changes/archive/"):
		return true
	case strings.HasPrefix(slash, "openspec/changes/") && strings.HasPrefix(file.Status, "D"):
		return true
	default:
		return false
	}
}

func changedFilePaths(files []changedFile) []string {
	out := make([]string, 0, len(files))
	for _, file := range files {
		out = append(out, file.Path)
	}
	return out
}

func parseNameStatus(output string) []changedFile {
	var out []changedFile
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		out = append(out, changedFile{
			Status: parts[0],
			Path:   filepath.ToSlash(parts[len(parts)-1]),
		})
	}
	return out
}

func hasAnyPrefix(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func parseCheckIgnore(output string) []Violation {
	var out []Violation
	for _, line := range splitLines(output) {
		meta, path, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		parts := strings.SplitN(meta, ":", 3)
		violation := Violation{Path: filepath.ToSlash(path), Kind: "ignored"}
		if len(parts) > 0 {
			violation.Source = parts[0]
		}
		if len(parts) > 1 {
			violation.Line = parts[1]
		}
		if len(parts) > 2 {
			violation.Pattern = parts[2]
		}
		out = append(out, violation)
	}
	return out
}

func runGit(target string, args ...string) (string, string, error) {
	cmd := exec.Command("git", append([]string{"-C", target}, args...)...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func splitLines(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, filepath.ToSlash(line))
		}
	}
	return out
}

func present(report Report, stdout io.Writer) {
	fmt.Fprintf(stdout, "PR guard para %s\n", report.TargetRoot)
	fmt.Fprintf(stdout, "Rango: %s\n", report.Range)
	fmt.Fprintf(stdout, "Archivos en diff: %d\n", len(report.ChangedFiles))
	if report.OK {
		fmt.Fprintln(stdout, "PR guard OK: no hay paths ignorados ni internos en el diff")
		return
	}
	fmt.Fprintln(stdout, "PR guard bloqueado: el diff incluye paths ignorados o internos")
	fmt.Fprintln(stdout, "Nota: .gitignore evita agregar archivos nuevos por accidente, pero no impide que archivos ya trackeados entren por cherry-pick, worktree o commits existentes.")
	if len(report.IgnoredMatches) > 0 {
		fmt.Fprintln(stdout, "Paths ignorados detectados con git check-ignore -v --no-index --stdin:")
		for _, v := range report.IgnoredMatches {
			where := v.Source
			if v.Line != "" {
				where += ":" + v.Line
			}
			fmt.Fprintf(stdout, "- %s (patrón %q en %s)\n", v.Path, v.Pattern, where)
		}
	}
	if len(report.InternalMatches) > 0 {
		fmt.Fprintln(stdout, "Paths internos/metadata detectados:")
		for _, v := range report.InternalMatches {
			fmt.Fprintf(stdout, "- %s (prefijo %q)\n", v.Path, v.Pattern)
		}
	}
	fmt.Fprintln(stdout, "Remediación sugerida:")
	for _, step := range report.Remediation {
		fmt.Fprintf(stdout, "- %s\n", step)
	}
}
