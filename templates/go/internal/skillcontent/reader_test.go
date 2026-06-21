package skillcontent

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestListReadsSkillFrontmatter(t *testing.T) {
	reader := New(fstest.MapFS{
		"example/SKILL.md": {Data: []byte("---\nname: example\ndescription: >-\n  Example skill\n---\nbody\n")},
	})
	items, err := reader.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Name != "example" || items[0].Description != "Example skill" {
		t.Fatalf("unexpected items: %#v", items)
	}
}

func TestReadRejectsTraversal(t *testing.T) {
	reader := New(fstest.MapFS{"example/SKILL.md": {Data: []byte("body\n")}})
	for _, path := range []string{"../x", "../../etc/passwd", `..\x`, "/tmp/x"} {
		t.Run(path, func(t *testing.T) {
			_, err := reader.Read("example", path)
			if err == nil || !strings.Contains(err.Error(), "invalid path") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestReadDefaultSkillFileIncludesGuidance(t *testing.T) {
	reader := New(fstest.MapFS{"example/SKILL.md": {Data: []byte("body\n")}})
	result, err := reader.Read("example", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Skill != "example" || result.Path != "SKILL.md" || result.Content != "body\n" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if !strings.Contains(result.Guidance, "skills read example --json") {
		t.Fatalf("missing guidance: %q", result.Guidance)
	}
}

func TestReadSubPath(t *testing.T) {
	reader := New(fstest.MapFS{
		"example/SKILL.md":          {Data: []byte("body\n")},
		"example/test-prompts.json": {Data: []byte("[]\n")},
	})
	result, err := reader.Read("example", "test-prompts.json")
	if err != nil {
		t.Fatal(err)
	}
	if result.Path != "test-prompts.json" || result.Content != "[]\n" || result.Guidance != "" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestReadUnknownSkill(t *testing.T) {
	reader := New(fstest.MapFS{"example/SKILL.md": {Data: []byte("body\n")}})
	_, err := reader.Read("missing", "")
	if err == nil || !strings.Contains(err.Error(), "unknown skill") {
		t.Fatalf("error = %v", err)
	}
}

func TestSplitTarget(t *testing.T) {
	name, relpath := SplitTarget("example/test-prompts.json")
	if name != "example" || relpath != "test-prompts.json" {
		t.Fatalf("name=%q relpath=%q", name, relpath)
	}
}
