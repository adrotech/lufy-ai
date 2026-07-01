package memory

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"gopkg.in/yaml.v3"
)

type Options struct {
	Target string
	Query  string
	JSON   bool
}

type CaptureOptions struct {
	Target string
	Title  string
	Type   string
	Body   string
	Links  []string
	DryRun bool
	JSON   bool
}

type ConnectOptions struct {
	Target        string
	From          string
	To            string
	Bidirectional bool
	DryRun        bool
	JSON          bool
}

type IndexOptions struct {
	Target string
	DryRun bool
	JSON   bool
}

type Report struct {
	OK         bool           `json:"ok"`
	TargetRoot string         `json:"targetRoot"`
	Provider   string         `json:"provider"`
	Root       string         `json:"root"`
	Status     Status         `json:"status"`
	Checks     []Check        `json:"checks"`
	Results    []SearchResult `json:"results,omitempty"`
}

type Status struct {
	Configured       bool     `json:"configured"`
	Initialized      bool     `json:"initialized"`
	Availability     string   `json:"availability"`
	Reason           string   `json:"reason,omitempty"`
	Recovery         string   `json:"recovery,omitempty"`
	SchemaVersion    int      `json:"schemaVersion"`
	Notes            int      `json:"notes"`
	Drafts           int      `json:"drafts"`
	Deprecated       int      `json:"deprecated"`
	Superseded       int      `json:"superseded"`
	BrokenBacklinks  int      `json:"brokenBacklinks"`
	PendingDrafts    []string `json:"pendingDrafts,omitempty"`
	BacklinksIndexed bool     `json:"backlinksIndexed"`
}

type Check struct {
	Level   string `json:"level"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

type SearchResult struct {
	Status string `json:"status"`
	Path   string `json:"path"`
	Line   int    `json:"line"`
	Text   string `json:"text"`
}

type CaptureResult struct {
	Path    string   `json:"path"`
	Slug    string   `json:"slug"`
	Created bool     `json:"created"`
	Updated bool     `json:"updated"`
	DryRun  bool     `json:"dryRun"`
	Links   []string `json:"links"`
}

type ConnectResult struct {
	Changed []string `json:"changed"`
	DryRun  bool     `json:"dryRun"`
}

type IndexResult struct {
	Path   string `json:"path"`
	Notes  int    `json:"notes"`
	Links  int    `json:"links"`
	DryRun bool   `json:"dryRun"`
}

type Service struct{}

func NewService() Service {
	return Service{}
}

func (s Service) Init(opts Options, stdout io.Writer) error {
	target, cfg, err := prepareTargetConfig(opts.Target)
	if err != nil {
		return err
	}
	root, err := memoryRoot(target, cfg)
	if err != nil {
		return err
	}
	for _, dir := range []string{"inbox", "knowledge", "maps", "index"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			return err
		}
	}
	files := map[string]string{
		"MEMORY.md":            memoryIndexTemplate(),
		"maps/_app-profile.md": appProfileTemplate(),
		"index/backlinks.json": "{\n  \"backlinks\": {}\n}\n",
		".gitignore":           gitignoreTemplate(),
	}
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := writeIfMissing(path, content); err != nil {
			return err
		}
	}
	report, err := s.BuildStatus(Options{Target: target})
	if err != nil {
		return err
	}
	if opts.JSON {
		return writeJSON(stdout, report)
	}
	fmt.Fprintf(stdout, "Memoria Obsidian inicializada: %s\n", filepath.ToSlash(cfg.Root))
	fmt.Fprintf(stdout, "Notas privadas ignoradas por Git: inbox/, knowledge/\n")
	return nil
}

func (s Service) Status(opts Options, stdout io.Writer) error {
	report, err := s.BuildStatus(opts)
	if err != nil {
		return err
	}
	if opts.JSON {
		return writeJSON(stdout, report)
	}
	fmt.Fprintf(stdout, "Memoria para %s\n", report.TargetRoot)
	fmt.Fprintf(stdout, "Provider: %s\n", report.Provider)
	fmt.Fprintf(stdout, "Root: %s\n", report.Root)
	fmt.Fprintf(stdout, "Inicializada: %s\n", yesNo(report.Status.Initialized))
	fmt.Fprintf(stdout, "Notas: %d draft=%d deprecated=%d superseded=%d backlinks_rotos=%d\n", report.Status.Notes, report.Status.Drafts, report.Status.Deprecated, report.Status.Superseded, report.Status.BrokenBacklinks)
	for _, check := range report.Checks {
		if check.Path == "" {
			fmt.Fprintf(stdout, "- [%s] %s\n", check.Level, check.Message)
			continue
		}
		fmt.Fprintf(stdout, "- [%s] %s %s\n", check.Level, check.Path, check.Message)
	}
	return nil
}

func (s Service) Validate(opts Options, stdout io.Writer) error {
	report, err := s.BuildValidate(opts)
	if err != nil {
		return err
	}
	if opts.JSON {
		if err := writeJSON(stdout, report); err != nil {
			return err
		}
		if !report.OK {
			return fmt.Errorf("memory validate detectó problemas")
		}
		return nil
	}
	for _, check := range report.Checks {
		if check.Path == "" {
			fmt.Fprintf(stdout, "%s: %s\n", check.Level, check.Message)
			continue
		}
		fmt.Fprintf(stdout, "%s: %s: %s\n", check.Level, check.Message, check.Path)
	}
	if !report.OK {
		return fmt.Errorf("memory validate detectó problemas")
	}
	fmt.Fprintln(stdout, "Memoria OK")
	return nil
}

func (s Service) Search(opts Options, stdout io.Writer) error {
	if strings.TrimSpace(opts.Query) == "" {
		return fmt.Errorf("memory search requiere query")
	}
	report, err := s.BuildSearch(opts)
	if err != nil {
		return err
	}
	if opts.JSON {
		return writeJSON(stdout, report)
	}
	for _, result := range report.Results {
		fmt.Fprintf(stdout, "[%s] %s:%d %s\n", result.Status, result.Path, result.Line, result.Text)
	}
	return nil
}

func (s Service) Capture(opts CaptureOptions, stdout io.Writer) error {
	target, root, _, err := s.requireInitialized(opts.Target)
	if err != nil {
		return err
	}
	title := strings.TrimSpace(opts.Title)
	body := strings.TrimSpace(opts.Body)
	noteType := strings.TrimSpace(opts.Type)
	if title == "" {
		return fmt.Errorf("memory capture requiere --title")
	}
	if body == "" {
		return fmt.Errorf("memory capture requiere texto")
	}
	if noteType == "" {
		noteType = "lesson"
	}
	if !contains([]string{"decision", "rule", "flow", "lesson", "concept"}, noteType) {
		return fmt.Errorf("type inválido: %s", noteType)
	}
	known, err := knownNoteSlugs(root)
	if err != nil {
		return err
	}
	links, err := normalizeExistingLinks(opts.Links, known)
	if err != nil {
		return err
	}
	slug := slugify(title)
	if slug == "" {
		return fmt.Errorf("title no genera slug válido")
	}
	path := filepath.Join(root, "knowledge", slug+".md")
	created := false
	updated := false
	var content string
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		created = true
		content = newNoteContent(slug, title, noteType, body, links)
	} else if err != nil {
		return err
	} else {
		existing, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content = string(existing)
		content, updated = appendCaptureUpdate(content, body, links)
	}
	rel, _ := filepath.Rel(target, path)
	result := CaptureResult{Path: filepath.ToSlash(rel), Slug: slug, Created: created, Updated: updated || created, DryRun: opts.DryRun, Links: links}
	if !opts.DryRun && (created || updated) {
		if err := platform.WriteFileAtomic(path, []byte(content), 0o644); err != nil {
			return err
		}
		if _, err := s.BuildIndex(IndexOptions{Target: target}); err != nil {
			return err
		}
		validateReport, err := s.BuildValidate(Options{Target: target})
		if err != nil {
			return err
		}
		if !validateReport.OK {
			return fmt.Errorf("memory validate detectó problemas: %s", firstCheckMessage(validateReport.Checks))
		}
	}
	if opts.JSON {
		return writeJSON(stdout, result)
	}
	verb := "actualizada"
	if created {
		verb = "creada"
	} else if opts.DryRun {
		verb = "propuesta"
	}
	fmt.Fprintf(stdout, "Nota %s: %s\n", verb, result.Path)
	return nil
}

func (s Service) Connect(opts ConnectOptions, stdout io.Writer) error {
	target, root, _, err := s.requireInitialized(opts.Target)
	if err != nil {
		return err
	}
	known, err := knownNoteSlugs(root)
	if err != nil {
		return err
	}
	from := slugify(opts.From)
	to := slugify(opts.To)
	if from == "" || to == "" {
		return fmt.Errorf("memory connect requiere slugs from y to")
	}
	if from == to {
		return fmt.Errorf("memory connect requiere notas distintas")
	}
	fromPath, ok := known[from]
	if !ok {
		return fmt.Errorf("nota no encontrada: %s", from)
	}
	if _, ok := known[to]; !ok {
		return fmt.Errorf("nota no encontrada: %s", to)
	}
	changed := []string{}
	if did, err := addLinkToNote(root, fromPath, to, opts.DryRun); err != nil {
		return err
	} else if did {
		changed = append(changed, filepath.ToSlash(fromPath))
	}
	if opts.Bidirectional {
		toPath := known[to]
		if did, err := addLinkToNote(root, toPath, from, opts.DryRun); err != nil {
			return err
		} else if did {
			changed = append(changed, filepath.ToSlash(toPath))
		}
	}
	if !opts.DryRun {
		if _, err := s.BuildIndex(IndexOptions{Target: target}); err != nil {
			return err
		}
		validateReport, err := s.BuildValidate(Options{Target: target})
		if err != nil {
			return err
		}
		if !validateReport.OK {
			return fmt.Errorf("memory validate detectó problemas: %s", firstCheckMessage(validateReport.Checks))
		}
	}
	result := ConnectResult{Changed: changed, DryRun: opts.DryRun}
	if opts.JSON {
		return writeJSON(stdout, result)
	}
	fmt.Fprintf(stdout, "Conexiones actualizadas: %d\n", len(changed))
	return nil
}

func (s Service) Index(opts IndexOptions, stdout io.Writer) error {
	result, err := s.BuildIndex(opts)
	if err != nil {
		return err
	}
	if opts.JSON {
		return writeJSON(stdout, result)
	}
	fmt.Fprintf(stdout, "Índice de backlinks actualizado: %s notas=%d links=%d\n", result.Path, result.Notes, result.Links)
	return nil
}

func (s Service) BuildIndex(opts IndexOptions) (IndexResult, error) {
	target, root, cfg, err := s.requireInitialized(opts.Target)
	if err != nil {
		return IndexResult{}, err
	}
	known, err := knownNoteSlugs(root)
	if err != nil {
		return IndexResult{}, err
	}
	backlinks := map[string][]string{}
	linkCount := 0
	seenRel := map[string]bool{}
	noteCount := 0
	for _, rel := range known {
		if seenRel[rel] {
			continue
		}
		seenRel[rel] = true
		noteCount++
		path := filepath.Join(root, filepath.FromSlash(rel))
		n, err := readNote(path)
		if err != nil {
			return IndexResult{}, err
		}
		sourceSlug := canonicalNoteSlug(path, n)
		for _, link := range wikiLinks(n.Body) {
			linked := slugify(link)
			if _, ok := known[linked]; !ok {
				continue
			}
			backlinks[linked] = appendUnique(backlinks[linked], sourceSlug)
			linkCount++
		}
	}
	for slug := range backlinks {
		sort.Strings(backlinks[slug])
	}
	indexPath, err := platform.SafeJoin(target, cfg.BacklinksIndex)
	if err != nil {
		return IndexResult{}, err
	}
	payload := struct {
		Backlinks map[string][]string `json:"backlinks"`
	}{Backlinks: backlinks}
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return IndexResult{}, err
	}
	if !opts.DryRun {
		if err := platform.WriteFileAtomic(indexPath, append(body, '\n'), 0o644); err != nil {
			return IndexResult{}, err
		}
	}
	rel, _ := filepath.Rel(target, indexPath)
	return IndexResult{Path: filepath.ToSlash(rel), Notes: noteCount, Links: linkCount, DryRun: opts.DryRun}, nil
}

func (s Service) BuildStatus(opts Options) (Report, error) {
	target, cfg, exists, err := loadTargetConfig(opts.Target)
	if err != nil {
		return Report{}, err
	}
	report := baseReport(target, cfg, exists)
	root, err := memoryRoot(target, cfg)
	if err != nil {
		report.Checks = append(report.Checks, Check{Level: "fail", Path: cfg.Root, Message: err.Error()})
		return finish(report), nil
	}
	report.Status = collectStatus(target, root, cfg, false, &report.Checks)
	return finish(report), nil
}

func (s Service) BuildValidate(opts Options) (Report, error) {
	target, cfg, exists, err := loadTargetConfig(opts.Target)
	if err != nil {
		return Report{}, err
	}
	report := baseReport(target, cfg, exists)
	root, err := memoryRoot(target, cfg)
	if err != nil {
		report.Checks = append(report.Checks, Check{Level: "fail", Path: cfg.Root, Message: err.Error()})
		return finish(report), nil
	}
	report.Status = collectStatus(target, root, cfg, true, &report.Checks)
	return finish(report), nil
}

func (s Service) BuildSearch(opts Options) (Report, error) {
	report, err := s.BuildStatus(opts)
	if err != nil {
		return Report{}, err
	}
	if !report.Status.Initialized {
		return report, fmt.Errorf("memoria no inicializada; ejecuta lufy-ai memory init --target %s", report.TargetRoot)
	}
	root, err := platform.SafeJoin(report.TargetRoot, report.Root)
	if err != nil {
		return Report{}, err
	}
	results, err := searchMemory(root, opts.Query)
	if err != nil {
		return Report{}, err
	}
	report.Results = results
	return report, nil
}

func (s Service) requireInitialized(targetValue string) (string, string, projectconfig.MemoryConfig, error) {
	target, cfg, _, err := loadTargetConfig(targetValue)
	if err != nil {
		return "", "", projectconfig.MemoryConfig{}, err
	}
	root, err := memoryRoot(target, cfg)
	if err != nil {
		return "", "", projectconfig.MemoryConfig{}, err
	}
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		return "", "", projectconfig.MemoryConfig{}, fmt.Errorf("memoria no inicializada; ejecuta lufy-ai memory init --target %s", target)
	}
	return target, root, cfg, nil
}

func prepareTargetConfig(targetValue string) (string, projectconfig.MemoryConfig, error) {
	target, err := platform.ResolveTargetPath(targetValue)
	if err != nil {
		return "", projectconfig.MemoryConfig{}, err
	}
	if _, err := projectconfig.NewService().Ensure(target); err != nil {
		return "", projectconfig.MemoryConfig{}, err
	}
	path := projectconfig.Path(target)
	cfg, err := projectconfig.Load(path)
	if err != nil {
		return "", projectconfig.MemoryConfig{}, err
	}
	changed := false
	if isZeroMemoryConfig(cfg.Memory) {
		cfg.Memory = projectconfig.DefaultMemoryConfig()
		changed = true
	}
	if isZeroParallelExecutionConfig(cfg.ParallelExecution) {
		cfg.ParallelExecution = projectconfig.DefaultParallelExecutionConfig()
		changed = true
	}
	if changed {
		data, err := projectconfig.Marshal(cfg)
		if err != nil {
			return "", projectconfig.MemoryConfig{}, err
		}
		if err := platform.WriteFileAtomic(path, data, 0o644); err != nil {
			return "", projectconfig.MemoryConfig{}, err
		}
	}
	return target, cfg.Memory, nil
}

func loadTargetConfig(targetValue string) (string, projectconfig.MemoryConfig, bool, error) {
	target, err := platform.ResolveTargetPath(targetValue)
	if err != nil {
		return "", projectconfig.MemoryConfig{}, false, err
	}
	path, err := projectconfig.ExistingPath(target)
	if err != nil {
		return "", projectconfig.MemoryConfig{}, false, err
	}
	cfg, err := projectconfig.Load(path)
	if errors.Is(err, os.ErrNotExist) {
		return target, projectconfig.DefaultMemoryConfig(), false, nil
	}
	if err != nil {
		return "", projectconfig.MemoryConfig{}, false, err
	}
	if isZeroMemoryConfig(cfg.Memory) {
		cfg.Memory = projectconfig.DefaultMemoryConfig()
	}
	return target, cfg.Memory, true, nil
}

func baseReport(target string, cfg projectconfig.MemoryConfig, configExists bool) Report {
	checks := []Check{}
	if !configExists {
		checks = append(checks, Check{Level: "warn", Path: projectconfig.ProjectConfigPath, Message: "project.yaml ausente; se usan defaults de memoria"})
	}
	return Report{OK: true, TargetRoot: target, Provider: cfg.Provider, Root: cfg.Root, Checks: checks}
}

func finish(report Report) Report {
	report.OK = true
	for _, check := range report.Checks {
		if check.Level == "fail" {
			report.OK = false
			break
		}
	}
	return report
}

func memoryRoot(target string, cfg projectconfig.MemoryConfig) (string, error) {
	if cfg.Provider != "obsidian" {
		return "", fmt.Errorf("provider de memoria no soportado: %s", cfg.Provider)
	}
	if cfg.Root == "" {
		return "", fmt.Errorf("memory.root vacío")
	}
	return platform.SafeJoin(target, cfg.Root)
}

func collectStatus(target, root string, cfg projectconfig.MemoryConfig, strict bool, checks *[]Check) Status {
	status := Status{Configured: true, Availability: "ready", SchemaVersion: cfg.SchemaVersion}
	if cfg.SchemaVersion != 1 {
		*checks = append(*checks, Check{Level: "fail", Path: projectconfig.ProjectConfigPath, Message: "memory.schema_version debe ser 1"})
	}
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		level := "warn"
		if strict {
			level = "fail"
		}
		*checks = append(*checks, Check{Level: level, Path: cfg.Root, Message: "memoria no inicializada"})
		status.Availability = "not_available"
		status.Reason = "obsidian memory root is not initialized"
		status.Recovery = "lufy-ai memory init --target <repo>"
		return status
	}
	status.Initialized = true
	required := []string{"MEMORY.md", "inbox", "knowledge", "maps/_app-profile.md", "index/backlinks.json", ".gitignore"}
	for _, rel := range required {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err != nil {
			*checks = append(*checks, Check{Level: "fail", Path: filepath.ToSlash(filepath.Join(cfg.Root, rel)), Message: "entrada requerida ausente"})
		}
	}
	if _, err := os.Stat(filepath.Join(root, "index", "backlinks.json")); err == nil {
		status.BacklinksIndexed = true
	}
	validateNotes(target, root, cfg, strict, &status, checks)
	return status
}

func validateNotes(target, root string, cfg projectconfig.MemoryConfig, strict bool, status *Status, checks *[]Check) {
	knownNotes := map[string]string{}
	noteFiles := []string{}
	for _, dir := range []string{"knowledge", "maps"} {
		base := filepath.Join(root, dir)
		_ = filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			slug := strings.TrimSuffix(filepath.Base(path), ".md")
			knownNotes[strings.ToLower(slug)] = filepath.ToSlash(rel)
			if n, err := readNote(path); err == nil && strings.TrimSpace(n.Name) != "" {
				knownNotes[strings.ToLower(slugify(n.Name))] = filepath.ToSlash(rel)
			}
			noteFiles = append(noteFiles, path)
			return nil
		})
	}
	sort.Strings(noteFiles)
	for _, path := range noteFiles {
		rel, _ := filepath.Rel(target, path)
		note, err := readNote(path)
		if err != nil {
			*checks = append(*checks, Check{Level: "fail", Path: filepath.ToSlash(rel), Message: err.Error()})
			continue
		}
		if strings.HasPrefix(filepath.ToSlash(rel), filepath.ToSlash(filepath.Join(cfg.Root, "knowledge"))+"/") {
			status.Notes++
			validateNoteSchema(filepath.ToSlash(rel), note, checks)
		}
		switch note.Status {
		case "draft":
			status.Drafts++
			status.PendingDrafts = append(status.PendingDrafts, filepath.ToSlash(rel))
		case "deprecated":
			status.Deprecated++
		case "superseded":
			status.Superseded++
		}
		for _, link := range wikiLinks(note.Body) {
			if _, ok := knownNotes[strings.ToLower(slugify(link))]; ok {
				continue
			}
			status.BrokenBacklinks++
			level := "warn"
			if strict {
				level = "fail"
			}
			*checks = append(*checks, Check{Level: level, Path: filepath.ToSlash(rel), Message: fmt.Sprintf("backlink roto: [[%s]]", link)})
		}
	}
}

type note struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Status      string `yaml:"status"`
	Body        string `yaml:"-"`
}

func readNote(path string) (note, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return note{}, err
	}
	text := string(body)
	if !strings.HasPrefix(text, "---\n") {
		return note{Body: text}, fmt.Errorf("frontmatter requerido")
	}
	rest := text[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return note{}, fmt.Errorf("frontmatter sin cierre")
	}
	var decoded note
	if err := yaml.Unmarshal([]byte(rest[:end]), &decoded); err != nil {
		return note{}, fmt.Errorf("frontmatter YAML inválido: %w", err)
	}
	decoded.Body = rest[end+len("\n---"):]
	return decoded, nil
}

func validateNoteSchema(rel string, n note, checks *[]Check) {
	required := map[string]string{"name": n.Name, "description": n.Description, "type": n.Type, "status": n.Status}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			*checks = append(*checks, Check{Level: "fail", Path: rel, Message: "frontmatter requiere " + field})
		}
	}
	if strings.EqualFold(strings.TrimSpace(n.Name), strings.TrimSpace(n.Description)) && n.Name != "" {
		*checks = append(*checks, Check{Level: "fail", Path: rel, Message: "description debe aportar más contexto que name"})
	}
	if !contains([]string{"decision", "rule", "flow", "lesson", "concept"}, n.Type) && n.Type != "" {
		*checks = append(*checks, Check{Level: "fail", Path: rel, Message: "type inválido: " + n.Type})
	}
	if !contains([]string{"active", "draft", "deprecated", "superseded"}, n.Status) && n.Status != "" {
		*checks = append(*checks, Check{Level: "fail", Path: rel, Message: "status inválido: " + n.Status})
	}
	if n.Type == "decision" && !strings.Contains(n.Body, "**Why:**") {
		*checks = append(*checks, Check{Level: "fail", Path: rel, Message: "decision requiere sección **Why:**"})
	}
}

func searchMemory(root, query string) ([]SearchResult, error) {
	searchRoots := []string{filepath.Join(root, "knowledge"), filepath.Join(root, "maps")}
	if rg, err := exec.LookPath("rg"); err == nil {
		args := []string{"-n", "--ignore-case", "--fixed-strings", query}
		args = append(args, searchRoots...)
		cmd := exec.Command(rg, args...)
		output, err := cmd.Output()
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
				return nil, nil
			}
			return nil, err
		}
		return parseRGOutput(root, string(output)), nil
	}
	return searchMemoryFallback(root, searchRoots, query)
}

func parseRGOutput(root, output string) []SearchResult {
	results := []SearchResult{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		path, line, text, ok := parseRGLine(scanner.Text())
		if !ok {
			continue
		}
		rel, _ := filepath.Rel(root, path)
		results = append(results, SearchResult{Status: noteStatus(path), Path: filepath.ToSlash(rel), Line: line, Text: strings.TrimSpace(text)})
	}
	sortSearchResults(results)
	return results
}

func parseRGLine(lineText string) (string, int, string, bool) {
	textColon := strings.LastIndex(lineText, ":")
	if textColon < 0 {
		return "", 0, "", false
	}
	beforeText := lineText[:textColon]
	lineColon := strings.LastIndex(beforeText, ":")
	if lineColon < 0 {
		return "", 0, "", false
	}
	line := 0
	if _, err := fmt.Sscanf(beforeText[lineColon+1:], "%d", &line); err != nil || line <= 0 {
		return "", 0, "", false
	}
	return beforeText[:lineColon], line, lineText[textColon+1:], true
}

func searchMemoryFallback(root string, roots []string, query string) ([]SearchResult, error) {
	results := []SearchResult{}
	needle := strings.ToLower(query)
	for _, base := range roots {
		_ = filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				return nil
			}
			file, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			line := 0
			for scanner.Scan() {
				line++
				text := scanner.Text()
				if !strings.Contains(strings.ToLower(text), needle) {
					continue
				}
				rel, _ := filepath.Rel(root, path)
				results = append(results, SearchResult{Status: noteStatus(path), Path: filepath.ToSlash(rel), Line: line, Text: strings.TrimSpace(text)})
			}
			return nil
		})
	}
	sortSearchResults(results)
	return results, nil
}

func noteStatus(path string) string {
	n, err := readNote(path)
	if err != nil || n.Status == "" {
		return "unknown"
	}
	return n.Status
}

func sortSearchResults(results []SearchResult) {
	sort.SliceStable(results, func(i, j int) bool {
		if statusRank(results[i].Status) != statusRank(results[j].Status) {
			return statusRank(results[i].Status) < statusRank(results[j].Status)
		}
		if results[i].Path != results[j].Path {
			return results[i].Path < results[j].Path
		}
		return results[i].Line < results[j].Line
	})
}

func knownNoteSlugs(root string) (map[string]string, error) {
	known := map[string]string{}
	for _, dir := range []string{"knowledge", "maps"} {
		base := filepath.Join(root, dir)
		_ = filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				return nil
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			slug := slugify(strings.TrimSuffix(filepath.Base(path), ".md"))
			known[slug] = filepath.ToSlash(rel)
			if n, readErr := readNote(path); readErr == nil && strings.TrimSpace(n.Name) != "" {
				known[slugify(n.Name)] = filepath.ToSlash(rel)
			}
			return nil
		})
	}
	return known, nil
}

func canonicalNoteSlug(path string, n note) string {
	if strings.TrimSpace(n.Name) != "" {
		return slugify(n.Name)
	}
	return slugify(strings.TrimSuffix(filepath.Base(path), ".md"))
}

func normalizeExistingLinks(values []string, known map[string]string) ([]string, error) {
	links := []string{}
	for _, value := range values {
		slug := slugify(value)
		if slug == "" {
			continue
		}
		if _, ok := known[slug]; !ok {
			return nil, fmt.Errorf("nota enlazada no encontrada: %s", slug)
		}
		links = appendUnique(links, slug)
	}
	sort.Strings(links)
	return links, nil
}

func newNoteContent(slug, title, noteType, body string, links []string) string {
	var out strings.Builder
	fmt.Fprintf(&out, "---\nname: %s\ndescription: %s\ntype: %s\nstatus: active\n---\n\n", slug, yamlScalar("Memoria durable: "+title), noteType)
	out.WriteString("## Summary\n\n")
	out.WriteString(body)
	out.WriteString("\n")
	if noteType == "decision" {
		out.WriteString("\n**Why:** ")
		out.WriteString(body)
		out.WriteString("\n")
	}
	if len(links) > 0 {
		out.WriteString("\n## Links\n\n")
		for _, link := range links {
			fmt.Fprintf(&out, "- [[%s]]\n", link)
		}
	}
	return out.String()
}

func appendCaptureUpdate(content, body string, links []string) (string, bool) {
	changed := false
	if !strings.Contains(content, body) {
		content = strings.TrimRight(content, "\n") + "\n\n## Updates\n\n" + body + "\n"
		changed = true
	}
	updated, linkChanged := appendLinks(content, links)
	return updated, changed || linkChanged
}

func addLinkToNote(root, rel, link string, dryRun bool) (bool, error) {
	path := filepath.Join(root, filepath.FromSlash(rel))
	body, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	updated, changed := appendLinks(string(body), []string{link})
	if !changed || dryRun {
		return changed, nil
	}
	return true, platform.WriteFileAtomic(path, []byte(updated), 0o644)
}

func appendLinks(content string, links []string) (string, bool) {
	changed := false
	for _, link := range links {
		needle := "[[" + link + "]]"
		if strings.Contains(content, needle) {
			continue
		}
		if !strings.Contains(content, "\n## Links\n") {
			content = strings.TrimRight(content, "\n") + "\n\n## Links\n\n"
		}
		content = strings.TrimRight(content, "\n") + "\n- " + needle + "\n"
		changed = true
	}
	return content, changed
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func firstCheckMessage(checks []Check) string {
	for _, check := range checks {
		if check.Level != "fail" {
			continue
		}
		if check.Path == "" {
			return check.Message
		}
		return check.Path + ": " + check.Message
	}
	return "sin detalle"
}

func yamlScalar(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "'", "''")
	return "'" + value + "'"
}

func statusRank(status string) int {
	switch status {
	case "active":
		return 0
	case "draft":
		return 1
	case "superseded":
		return 2
	case "deprecated":
		return 3
	default:
		return 4
	}
}

func wikiLinks(body string) []string {
	links := []string{}
	start := 0
	for {
		open := strings.Index(body[start:], "[[")
		if open < 0 {
			return links
		}
		open += start + 2
		close := strings.Index(body[open:], "]]")
		if close < 0 {
			return links
		}
		value := strings.TrimSpace(body[open : open+close])
		if pipe := strings.Index(value, "|"); pipe >= 0 {
			value = strings.TrimSpace(value[:pipe])
		}
		if value != "" {
			links = append(links, value)
		}
		start = open + close + 2
	}
}

func slugify(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	var out strings.Builder
	previousDash := false
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			out.WriteRune(r)
			previousDash = false
			continue
		}
		if !previousDash {
			out.WriteByte('-')
			previousDash = true
		}
	}
	return strings.Trim(out.String(), "-")
}

func writeIfMissing(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return platform.WriteFileAtomic(path, []byte(content), 0o644)
}

func writeJSON(stdout io.Writer, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "%s\n", body)
	return nil
}

func memoryIndexTemplate() string {
	return `# Lufy Memory

Memoria portable del proyecto. Las notas privadas viven en inbox/ y knowledge/ y quedan ignoradas por Git por defecto.

`
}

func appProfileTemplate() string {
	return `---
name: app-profile
description: Perfil navegable del proyecto para orientar sesiones Lufy.
type: concept
status: active
---

## Scope

Completar con dominios, flujos críticos, reglas y decisiones durables.

`
}

func gitignoreTemplate() string {
	return `*
!.gitignore
!MEMORY.md
!maps/
!maps/_app-profile.md
!index/
!index/backlinks.json
`
}

func isZeroMemoryConfig(config projectconfig.MemoryConfig) bool {
	return config.Provider == "" && config.Root == "" && config.Vault == "" && config.GitPolicy == "" && config.SchemaVersion == 0 && config.Search == "" && config.BacklinksIndex == "" && len(config.Extra) == 0
}

func isZeroParallelExecutionConfig(config projectconfig.ParallelExecutionConfig) bool {
	return config.Strategy == "" && config.MaxParallelAgents == 0 && config.ValidationMode == "" && !config.Enabled && !config.RequiresIndependentFiles && !config.RequiresMergePlan && len(config.Extra) == 0
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func yesNo(value bool) string {
	if value {
		return "sí"
	}
	return "no"
}
