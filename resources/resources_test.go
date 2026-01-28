package resources

import (
	"testing"

	"github.com/hermitmaster/aictl/internal/resource"
)

func TestLoadBundled(t *testing.T) {
	res, content, err := LoadBundled("git-commit")
	if err != nil {
		t.Fatalf("LoadBundled() error = %v", err)
	}

	if res.Name != "git-commit" {
		t.Errorf("Name = %v, want git-commit", res.Name)
	}
	if res.Version != "1.0.0" {
		t.Errorf("Version = %v, want 1.0.0", res.Version)
	}
	if res.Type != resource.TypeWorkflow {
		t.Errorf("Type = %v, want workflow", res.Type)
	}
	if content == "" {
		t.Error("Content should not be empty")
	}
}

func TestLoadBundledNotFound(t *testing.T) {
	_, _, err := LoadBundled("nonexistent-resource")
	if err == nil {
		t.Error("LoadBundled() should error for nonexistent resource")
	}
}

func TestListBundled(t *testing.T) {
	resources, err := ListBundled()
	if err != nil {
		t.Fatalf("ListBundled() error = %v", err)
	}

	// Should have at least git-commit
	found := false
	for _, res := range resources {
		if res.Name == "git-commit" {
			found = true
			break
		}
	}

	if !found {
		t.Error("ListBundled() should include git-commit")
	}
}
