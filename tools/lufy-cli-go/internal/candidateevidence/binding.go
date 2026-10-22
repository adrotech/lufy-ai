// Package candidateevidence vincula evidencia declarada a un candidato Git limpio.
package candidateevidence

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const SchemaVersion = "candidate-evidence/v1"

type Binding struct {
	SchemaVersion           string `json:"schema_version"`
	RepositorySHA256        string `json:"repository_sha256"`
	BaseCommit              string `json:"base_commit"`
	BaseTree                string `json:"base_tree"`
	CandidateCommit         string `json:"candidate_commit"`
	CandidateTree           string `json:"candidate_tree"`
	ChangedPathsModesSHA256 string `json:"changed_paths_modes_sha256"`
	PolicySHA256            string `json:"policy_sha256"`
	EvidenceSHA256          string `json:"evidence_sha256"`
	Role                    string `json:"role"`
}

var hashPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var oidPattern = regexp.MustCompile(`^([a-f0-9]{40}|[a-f0-9]{64})$`)
var rolePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

func digest(value []byte) string { sum := sha256.Sum256(value); return hex.EncodeToString(sum[:]) }

func (b Binding) validate() error {
	if b.SchemaVersion != SchemaVersion || !rolePattern.MatchString(b.Role) {
		return fmt.Errorf("binding schema o rol inválido")
	}
	for _, value := range []string{b.RepositorySHA256, b.ChangedPathsModesSHA256, b.PolicySHA256, b.EvidenceSHA256} {
		if !hashPattern.MatchString(value) {
			return fmt.Errorf("binding fingerprint inválido")
		}
	}
	for _, value := range []string{b.BaseCommit, b.BaseTree, b.CandidateCommit, b.CandidateTree} {
		if !oidPattern.MatchString(value) {
			return fmt.Errorf("binding Git identity inválida")
		}
	}
	return nil
}

// Decode rechaza duplicados antes del decode tipado y nunca imprime el input.
func Decode(input io.Reader) (Binding, error) {
	data, err := io.ReadAll(io.LimitReader(input, 65537))
	if err != nil || len(data) > 65536 {
		return Binding{}, fmt.Errorf("binding fuera de límite")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	first, err := dec.Token()
	if err != nil || first != json.Delim('{') {
		return Binding{}, fmt.Errorf("binding requiere objeto JSON")
	}
	seen := map[string]bool{}
	for dec.More() {
		key, err := dec.Token()
		if err != nil {
			return Binding{}, fmt.Errorf("binding JSON inválido")
		}
		name, ok := key.(string)
		if !ok || seen[name] {
			return Binding{}, fmt.Errorf("binding clave duplicada o inválida")
		}
		switch name {
		case "schema_version", "repository_sha256", "base_commit", "base_tree", "candidate_commit", "candidate_tree", "changed_paths_modes_sha256", "policy_sha256", "evidence_sha256", "role":
		default:
			return Binding{}, fmt.Errorf("binding clave desconocida")
		}
		seen[name] = true
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return Binding{}, fmt.Errorf("binding JSON inválido")
		}
	}
	if _, err := dec.Token(); err != nil {
		return Binding{}, fmt.Errorf("binding JSON inválido")
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		return Binding{}, fmt.Errorf("binding trailing data")
	}
	var b Binding
	typed := json.NewDecoder(bytes.NewReader(data))
	typed.DisallowUnknownFields()
	if err := typed.Decode(&b); err != nil {
		return Binding{}, fmt.Errorf("binding campos inválidos")
	}
	if err := b.validate(); err != nil {
		return Binding{}, err
	}
	return b, nil
}

type boundedBuffer struct{ body bytes.Buffer }

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if b.body.Len()+len(p) > 2*1024*1024 {
		return 0, fmt.Errorf("salida Git excedida")
	}
	return b.body.Write(p)
}

func git(ctx context.Context, root string, args ...string) ([]byte, error) {
	global := []string{"--no-optional-locks", "--no-pager", "-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false", "-c", "core.hooksPath=" + os.DevNull}
	cmd := exec.CommandContext(ctx, "git", append(global, args...)...)
	cmd.WaitDelay = time.Second
	cmd.Dir = root
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(entry), "GIT_") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull, "GIT_TERMINAL_PROMPT=0", "GIT_NO_REPLACE_OBJECTS=1", "GIT_OPTIONAL_LOCKS=0")
	var out boundedBuffer
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("lectura Git no disponible, inválida o excedida")
	}
	return out.body.Bytes(), nil
}

func textGit(ctx context.Context, root string, args ...string) (string, error) {
	b, err := git(ctx, root, args...)
	return strings.TrimSuffix(strings.TrimSuffix(string(b), "\n"), "\r"), err
}

func policyHash(root string) (string, error) {
	type part struct {
		Path    string
		Present bool
		SHA256  string
	}
	parts := []part{}
	for _, path := range []string{".lufy/contracts/delivery.md", ".lufy/contracts/result-contract.md", ".lufy/config/project.yaml"} {
		full := filepath.Join(root, filepath.FromSlash(path))
		info, err := os.Lstat(full)
		if os.IsNotExist(err) && path == ".lufy/config/project.yaml" {
			parts = append(parts, part{Path: path})
			continue
		}
		if err != nil || !info.Mode().IsRegular() || info.Size() > 2*1024*1024 {
			return "", fmt.Errorf("política requerida no disponible o no regular")
		}
		// También rechazar directorios padre simbólicos que desvíen la lectura.
		resolved, err := filepath.EvalSymlinks(full)
		if err != nil || resolved != full {
			return "", fmt.Errorf("política simbólica no admitida")
		}
		f, err := os.Open(full)
		if err != nil {
			return "", fmt.Errorf("política no legible")
		}
		body, readErr := io.ReadAll(io.LimitReader(f, 2*1024*1024+1))
		f.Close()
		if readErr != nil || len(body) > 2*1024*1024 {
			return "", fmt.Errorf("política fuera de límite")
		}
		parts = append(parts, part{Path: path, Present: true, SHA256: digest(body)})
	}
	body, _ := json.Marshal(parts)
	return digest(body), nil
}

func snapshot(ctx context.Context, root, base, evidenceSHA, role string) (Binding, error) {
	var b Binding
	b.SchemaVersion = SchemaVersion
	b.EvidenceSHA256 = evidenceSHA
	b.Role = role
	common, err := textGit(ctx, root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return b, err
	}
	common, err = filepath.EvalSymlinks(common)
	if err != nil {
		return b, fmt.Errorf("identidad de repositorio no disponible")
	}
	b.RepositorySHA256 = digest([]byte(filepath.Clean(common)))
	for _, item := range []struct {
		ref  string
		dest *string
	}{{base + "^{commit}", &b.BaseCommit}, {"HEAD^{commit}", &b.CandidateCommit}} {
		*item.dest, err = textGit(ctx, root, "rev-parse", "--verify", "--end-of-options", item.ref)
		if err != nil {
			return b, err
		}
	}
	b.BaseTree, err = textGit(ctx, root, "rev-parse", "--verify", "--end-of-options", b.BaseCommit+"^{tree}")
	if err != nil {
		return b, err
	}
	b.CandidateTree, err = textGit(ctx, root, "rev-parse", "--verify", "--end-of-options", b.CandidateCommit+"^{tree}")
	if err != nil {
		return b, err
	}
	keys, err := git(ctx, root, "config", "--null", "--list", "--name-only")
	if err != nil {
		return b, err
	}
	for _, key := range bytes.Split(keys, []byte{0}) {
		if strings.HasPrefix(strings.ToLower(string(key)), "filter.") {
			return b, fmt.Errorf("filtros Git configurados no admitidos para captura read-only")
		}
	}
	entries, err := git(ctx, root, "ls-files", "--stage", "-z")
	if err != nil {
		return b, err
	}
	for _, entry := range bytes.Split(entries, []byte{0}) {
		if bytes.HasPrefix(entry, []byte("160000 ")) {
			return b, fmt.Errorf("gitlinks y submódulos no admitidos para captura read-only")
		}
	}
	status, err := git(ctx, root, "status", "--porcelain=v1", "-z", "--untracked-files=all", "--ignore-submodules=none")
	if err != nil {
		return b, err
	}
	for _, entry := range bytes.Split(status, []byte{0}) {
		if len(entry) != 0 && string(entry) != "?? .lufy/config/project.yaml" {
			return b, fmt.Errorf("candidato no limpio; comprometer o aislar cambios antes de capturar")
		}
	}
	if _, err := git(ctx, root, "ls-files", "--error-unmatch", "--", ".lufy/contracts/delivery.md", ".lufy/contracts/result-contract.md"); err != nil {
		return b, fmt.Errorf("políticas canónicas deben pertenecer al candidato comprometido")
	}
	flags, err := git(ctx, root, "ls-files", "-v", "-z")
	if err != nil {
		return b, err
	}
	for _, entry := range bytes.Split(flags, []byte{0}) {
		if len(entry) > 0 && (entry[0] == 'S' || entry[0] >= 'a' && entry[0] <= 'z') {
			return b, fmt.Errorf("índice con assume-unchanged o skip-worktree; no se puede demostrar limpieza")
		}
	}
	diff, err := git(ctx, root, "diff", "--raw", "--no-abbrev", "--no-renames", "--no-ext-diff", "--no-textconv", "-z", b.BaseCommit, b.CandidateCommit, "--")
	if err != nil {
		return b, err
	}
	b.ChangedPathsModesSHA256 = digest(diff)
	b.PolicySHA256, err = policyHash(root)
	if err != nil {
		return b, err
	}
	return b, b.validate()
}

// Capture sólo observa commits; no crea árboles, actualiza índices ni toma locks.
func Capture(ctx context.Context, target, base, evidenceSHA, role string) (Binding, error) {
	if base == "" || !hashPattern.MatchString(evidenceSHA) || !rolePattern.MatchString(role) {
		return Binding{}, fmt.Errorf("base, evidencia o rol inválidos")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	root, err := textGit(ctx, target, "rev-parse", "--show-toplevel")
	if err != nil {
		return Binding{}, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return Binding{}, fmt.Errorf("repositorio no disponible")
	}
	first, err := snapshot(ctx, root, base, evidenceSHA, role)
	if err != nil {
		return Binding{}, err
	}
	second, err := snapshot(ctx, root, base, evidenceSHA, role)
	if err != nil {
		return Binding{}, err
	}
	if first != second {
		return Binding{}, fmt.Errorf("candidato cambió durante captura; repetir revisión")
	}
	return first, nil
}

func Verify(ctx context.Context, target, base, evidenceSHA, role string, b Binding) error {
	if err := b.validate(); err != nil {
		return err
	}
	current, err := Capture(ctx, target, base, evidenceSHA, role)
	if err != nil {
		return err
	}
	if current != b {
		return fmt.Errorf("evidencia obsoleta para candidato, base, repositorio, política o claim; repetir revisión y captura")
	}
	return nil
}
