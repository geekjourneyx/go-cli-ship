package skillcontent

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

type Reader struct {
	fsys fs.FS
}

type SkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ReadResult struct {
	Skill    string `json:"skill"`
	Path     string `json:"path"`
	Content  string `json:"content"`
	Guidance string `json:"guidance,omitempty"`
}

func New(fsys fs.FS) *Reader {
	return &Reader{fsys: fsys}
}

func (r *Reader) List() ([]SkillInfo, error) {
	if r == nil || r.fsys == nil {
		return nil, fmt.Errorf("skill content not available")
	}
	entries, err := fs.ReadDir(r.fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read skill content: %w", err)
	}
	items := make([]SkillInfo, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, err := fs.ReadFile(r.fsys, entry.Name()+"/SKILL.md")
		if err != nil {
			continue
		}
		items = append(items, SkillInfo{Name: entry.Name(), Description: parseDescription(data)})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, nil
}

func (r *Reader) Read(name, relpath string) (ReadResult, error) {
	if relpath == "" {
		relpath = "SKILL.md"
	}
	if err := r.ensureSkill(name); err != nil {
		return ReadResult{}, err
	}
	cleaned, err := cleanSubPath(relpath)
	if err != nil {
		return ReadResult{}, err
	}
	fullPath := name + "/" + cleaned
	info, err := fs.Stat(r.fsys, fullPath)
	if err != nil {
		return ReadResult{}, fmt.Errorf("skill file %q not found in %q", cleaned, name)
	}
	if info.IsDir() {
		return ReadResult{}, fmt.Errorf("skill file %q is a directory", cleaned)
	}
	data, err := fs.ReadFile(r.fsys, fullPath)
	if err != nil {
		return ReadResult{}, fmt.Errorf("read skill file %q: %w", cleaned, err)
	}
	result := ReadResult{Skill: name, Path: cleaned, Content: string(data)}
	if cleaned == "SKILL.md" {
		result.Guidance = fmt.Sprintf("Read this skill from the installed CLI package with `%s skills read %s --json` so the SOP stays in sync with this CLI version.", "__CLI_NAME__", name)
	}
	return result, nil
}

func SplitTarget(arg string) (name, relpath string) {
	name, relpath, _ = strings.Cut(arg, "/")
	return name, relpath
}

func (r *Reader) ensureSkill(name string) error {
	if r == nil || r.fsys == nil {
		return fmt.Errorf("skill content not available")
	}
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return unknownSkill(name)
	}
	info, err := fs.Stat(r.fsys, name)
	if err != nil || !info.IsDir() {
		return unknownSkill(name)
	}
	if _, err := fs.Stat(r.fsys, name+"/SKILL.md"); err != nil {
		return unknownSkill(name)
	}
	return nil
}

func unknownSkill(name string) error {
	return fmt.Errorf("unknown skill %q; run '__CLI_NAME__ skills list --json' to see available skills", name)
}

func cleanSubPath(relpath string) (string, error) {
	cleaned := path.Clean(relpath)
	if relpath == "" || path.IsAbs(relpath) || cleaned == "." || cleaned == ".." ||
		strings.Contains(relpath, `\`) || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", fmt.Errorf("invalid path %q: must be a relative path without '..'", relpath)
	}
	return cleaned, nil
}

func parseDescription(skillMD []byte) string {
	lines := strings.Split(string(skillMD), "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != "---" {
		return ""
	}
	inDescription := false
	parts := []string{}
	for _, line := range lines[1:] {
		trimmed := strings.TrimRight(line, "\r")
		if trimmed == "---" {
			break
		}
		if strings.HasPrefix(trimmed, "description:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
			value = strings.Trim(value, "\"")
			if value == ">-" || value == "|" || value == ">" {
				inDescription = true
				continue
			}
			return value
		}
		if inDescription {
			if strings.HasPrefix(trimmed, "  ") || strings.HasPrefix(trimmed, "\t") {
				parts = append(parts, strings.TrimSpace(trimmed))
				continue
			}
			break
		}
	}
	return strings.Join(parts, " ")
}
