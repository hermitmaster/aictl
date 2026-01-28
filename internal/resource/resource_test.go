package resource

import (
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantVersion string
		wantType    Type
		wantContent string
		wantErr     bool
	}{
		{
			name: "valid frontmatter",
			input: `---
name: test-resource
version: 1.0.0
type: workflow
description: A test resource
---

# Test Content

This is the content.`,
			wantName:    "test-resource",
			wantVersion: "1.0.0",
			wantType:    TypeWorkflow,
			wantContent: "# Test Content\n\nThis is the content.",
			wantErr:     false,
		},
		{
			name: "with tags and tools",
			input: `---
name: tagged-resource
version: 2.0.0
type: rules
description: Resource with tags
tags:
  - golang
  - testing
tools:
  - windsurf
  - cursor
---

Content here.`,
			wantName:    "tagged-resource",
			wantVersion: "2.0.0",
			wantType:    TypeRules,
			wantContent: "Content here.",
			wantErr:     false,
		},
		{
			name:        "no frontmatter",
			input:       "# Just content\n\nNo frontmatter here.",
			wantName:    "",
			wantVersion: "",
			wantType:    "",
			wantContent: "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, content, err := ParseFrontmatter(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if res.Name != tt.wantName {
				t.Errorf("ParseFrontmatter() name = %v, want %v", res.Name, tt.wantName)
			}
			if res.Version != tt.wantVersion {
				t.Errorf("ParseFrontmatter() version = %v, want %v", res.Version, tt.wantVersion)
			}
			if res.Type != tt.wantType {
				t.Errorf("ParseFrontmatter() type = %v, want %v", res.Type, tt.wantType)
			}
			if content != tt.wantContent {
				t.Errorf("ParseFrontmatter() content = %v, want %v", content, tt.wantContent)
			}
		})
	}
}

func TestResourceValidate(t *testing.T) {
	tests := []struct {
		name    string
		res     Resource
		wantErr bool
	}{
		{
			name: "valid resource",
			res: Resource{
				Name:        "test",
				Version:     "1.0.0",
				Type:        TypeWorkflow,
				Description: "A test",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			res: Resource{
				Version:     "1.0.0",
				Type:        TypeWorkflow,
				Description: "A test",
			},
			wantErr: true,
		},
		{
			name: "missing version",
			res: Resource{
				Name:        "test",
				Type:        TypeWorkflow,
				Description: "A test",
			},
			wantErr: true,
		},
		{
			name: "missing type",
			res: Resource{
				Name:        "test",
				Version:     "1.0.0",
				Description: "A test",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			res: Resource{
				Name:        "test",
				Version:     "1.0.0",
				Type:        "invalid",
				Description: "A test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceIsMultiFile(t *testing.T) {
	t.Run("single file resource", func(t *testing.T) {
		res := Resource{Name: "test"}
		if res.IsMultiFile() {
			t.Error("IsMultiFile() should return false for resource without files")
		}
	})

	t.Run("multi file resource", func(t *testing.T) {
		res := Resource{
			Name: "test",
			Files: []FileMapping{
				{Src: "a.sh", Dest: "bin/a"},
			},
		}
		if !res.IsMultiFile() {
			t.Error("IsMultiFile() should return true for resource with files")
		}
	})
}
