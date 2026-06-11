package managedcontent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
)

const implementerAgentRel = ".opencode/agents/implementer.md"

func CatalogWithRenderedHashes(catalog assets.Catalog, targetRoot string) (assets.Catalog, error) {
	updated := catalog
	updated.Assets = append([]assets.Asset{}, catalog.Assets...)
	for i, asset := range updated.Assets {
		if asset.Kind != assets.KindFile {
			continue
		}
		hash, err := SourceHash(catalog.SourceRoot, asset.SourceRel, targetRoot, asset.TargetRel)
		if err != nil {
			return assets.Catalog{}, err
		}
		updated.Assets[i].SourceSHA256 = hash
	}
	return updated, nil
}

func SourceHash(sourceRoot, sourceRel, targetRoot, targetRel string) (string, error) {
	body, err := Render(sourceRoot, sourceRel, targetRoot, targetRel)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func Render(sourceRoot, sourceRel, targetRoot, targetRel string) ([]byte, error) {
	body, err := readSourceContent(sourceRoot, sourceRel)
	if err != nil {
		return nil, err
	}
	if filepath.ToSlash(targetRel) != implementerAgentRel {
		return body, nil
	}
	configPath, err := projectconfig.ExistingPath(targetRoot)
	if err != nil {
		return nil, err
	}
	cfg, err := projectconfig.Load(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return body, nil
		}
		return nil, err
	}
	return renderImplementerPermissions(body, cfg.Validation.AllowedCommands.Implementer), nil
}

func readSourceContent(sourceRoot, sourceRel string) ([]byte, error) {
	if sourceRoot == assets.EmbeddedSourceRoot {
		return assets.ReadSourceFile(sourceRoot, sourceRel)
	}
	src := filepath.Join(sourceRoot, sourceRel)
	info, err := os.Lstat(src)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("source no es archivo regular seguro: %s", src)
	}
	return os.ReadFile(src)
}

func renderImplementerPermissions(body []byte, commands []string) []byte {
	cleanCommands := sanitizeCommands(commands)
	if len(cleanCommands) == 0 {
		return body
	}
	lines := strings.SplitAfter(string(body), "\n")
	insertAt := implementerEditInsertIndex(lines)
	if insertAt < 0 {
		return body
	}
	existing := map[string]bool{}
	for _, line := range lines {
		for _, command := range cleanCommands {
			if strings.Contains(line, `"`+command+`": allow`) {
				existing[command] = true
			}
		}
	}
	var inserted []string
	for _, command := range cleanCommands {
		if existing[command] {
			continue
		}
		inserted = append(inserted, `    "`+command+`": allow`+"\n")
	}
	if len(inserted) == 0 {
		return body
	}
	out := append([]string{}, lines[:insertAt]...)
	out = append(out, inserted...)
	out = append(out, lines[insertAt:]...)
	return []byte(strings.Join(out, ""))
}

func implementerEditInsertIndex(lines []string) int {
	frontmatterEnd := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			frontmatterEnd = i
			break
		}
	}
	if frontmatterEnd < 0 {
		return -1
	}
	permissionSeen := false
	editSeen := false
	for i := 1; i < frontmatterEnd; i++ {
		trimmed := strings.TrimSpace(lines[i])
		switch {
		case trimmed == "permission:" && !strings.HasPrefix(lines[i], " "):
			permissionSeen = true
		case permissionSeen && trimmed == "edit:" && strings.HasPrefix(lines[i], "  "):
			editSeen = true
		case editSeen && strings.HasPrefix(lines[i], "  ") && !strings.HasPrefix(lines[i], "    "):
			return i
		}
	}
	if editSeen {
		return frontmatterEnd
	}
	return -1
}

func sanitizeCommands(commands []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, command := range commands {
		command = strings.TrimSpace(command)
		if command == "" || strings.ContainsAny(command, "\r\n\"") || seen[command] {
			continue
		}
		seen[command] = true
		out = append(out, command)
	}
	return out
}
