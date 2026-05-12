package mergeblock

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	beginMarker = regexp.MustCompile(`^\s*<!--\s+LUFY:BEGIN\s+([A-Za-z0-9._-]+)\s+-->\s*$`)
	endMarker   = regexp.MustCompile(`^\s*<!--\s+LUFY:END\s+([A-Za-z0-9._-]+)\s+-->\s*$`)
)

type parsedFile struct {
	lines  []string
	blocks map[string]block
	order  []string
}

type block struct {
	id    string
	start int
	end   int
	lines []string
}

func Render(target, source []byte) ([]byte, error) {
	src, err := parse(source)
	if err != nil {
		return nil, fmt.Errorf("source merge-block inválido: %w", err)
	}
	if len(src.blocks) == 0 {
		return nil, fmt.Errorf("source merge-block no contiene bloques LUFY")
	}
	dst, err := parse(target)
	if err != nil {
		return nil, fmt.Errorf("target merge-block inválido: %w", err)
	}

	var out []string
	inserted := map[string]bool{}
	for i := 0; i < len(dst.lines); i++ {
		begin := beginMarker.FindStringSubmatch(strings.TrimSuffix(dst.lines[i], "\n"))
		if begin == nil {
			out = append(out, dst.lines[i])
			continue
		}
		id := begin[1]
		if replacement, ok := src.blocks[id]; ok {
			out = append(out, replacement.lines...)
			inserted[id] = true
		} else {
			out = append(out, dst.blocks[id].lines...)
		}
		i = dst.blocks[id].end
	}

	var missing []string
	for _, id := range src.order {
		if !inserted[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		if len(out) > 0 && !strings.HasSuffix(out[len(out)-1], "\n") {
			out[len(out)-1] += "\n"
		}
		if len(out) > 0 && strings.TrimSpace(out[len(out)-1]) != "" {
			out = append(out, "\n")
		}
		sort.SliceStable(missing, func(i, j int) bool {
			return indexOf(src.order, missing[i]) < indexOf(src.order, missing[j])
		})
		for _, id := range missing {
			out = append(out, src.blocks[id].lines...)
			if len(out) > 0 && !strings.HasSuffix(out[len(out)-1], "\n") {
				out[len(out)-1] += "\n"
			}
		}
	}
	return []byte(strings.Join(out, "")), nil
}

func parse(body []byte) (parsedFile, error) {
	lines := splitLines(string(body))
	parsed := parsedFile{lines: lines, blocks: map[string]block{}}
	openID := ""
	openStart := -1
	for i, line := range lines {
		trimmed := strings.TrimSuffix(line, "\n")
		if begin := beginMarker.FindStringSubmatch(trimmed); begin != nil {
			id := begin[1]
			if openID != "" {
				return parsedFile{}, fmt.Errorf("bloque LUFY anidado %q dentro de %q", id, openID)
			}
			if _, exists := parsed.blocks[id]; exists {
				return parsedFile{}, fmt.Errorf("bloque LUFY duplicado: %s", id)
			}
			openID = id
			openStart = i
			continue
		}
		if end := endMarker.FindStringSubmatch(trimmed); end != nil {
			id := end[1]
			if openID == "" {
				return parsedFile{}, fmt.Errorf("marcador LUFY END sin BEGIN: %s", id)
			}
			if id != openID {
				return parsedFile{}, fmt.Errorf("marcador LUFY END %q no coincide con BEGIN %q", id, openID)
			}
			parsed.blocks[id] = block{id: id, start: openStart, end: i, lines: append([]string(nil), lines[openStart:i+1]...)}
			parsed.order = append(parsed.order, id)
			openID = ""
			openStart = -1
		}
	}
	if openID != "" {
		return parsedFile{}, fmt.Errorf("bloque LUFY sin cierre: %s", openID)
	}
	return parsed, nil
}

func splitLines(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.SplitAfter(value, "\n")
	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func indexOf(values []string, value string) int {
	for i, item := range values {
		if item == value {
			return i
		}
	}
	return len(values)
}
