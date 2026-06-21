package skillcontent

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

const EnvSkillsDir = "__RELEASE_ENV_PREFIX___SKILLS_DIR"

func OpenFS() (fs.FS, error) {
	if dir := os.Getenv(EnvSkillsDir); dir != "" {
		return openSkillsDir(dir)
	}
	if dir := findSkillsFromWorkingDir(); dir != "" {
		return openSkillsDir(dir)
	}
	if dir := findSkillsBesideExecutable(); dir != "" {
		return openSkillsDir(dir)
	}
	if dir := findSkillsFromSourceFile(); dir != "" {
		return openSkillsDir(dir)
	}
	return nil, fmt.Errorf("set %s to the directory containing <skill-name>/SKILL.md", EnvSkillsDir)
}

func openSkillsDir(dir string) (fs.FS, error) {
	cleaned, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(cleaned); err != nil {
		return nil, err
	}
	return os.DirFS(cleaned), nil
}

func findSkillsFromWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return findSkillsInAncestors(wd)
}

func findSkillsBesideExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return findSkillsInAncestors(filepath.Dir(exe))
}

func findSkillsFromSourceFile() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return findSkillsInAncestors(filepath.Dir(file))
}

func findSkillsInAncestors(start string) string {
	dir, err := filepath.Abs(start)
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, "skills")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
