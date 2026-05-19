package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

const defaultDirectory = ".playpub"

type InitOptions struct {
	Directory string `json:"directory,omitempty"`
	Force     bool   `json:"force"`
	DryRun    bool   `json:"dryRun"`
}

func (o InitOptions) Validate() error {
	if o.targetDirectory() == "" {
		return fmt.Errorf("directory is required")
	}
	return nil
}

func (o InitOptions) targetDirectory() string {
	if o.Directory == "" {
		return defaultDirectory
	}
	return o.Directory
}

type InitFile struct {
	Path      string `json:"path"`
	Exists    bool   `json:"exists"`
	Written   bool   `json:"written"`
	Overwrote bool   `json:"overwrote"`
}

type InitPlan struct {
	Directory string     `json:"directory"`
	Force     bool       `json:"force"`
	DryRun    bool       `json:"dryRun"`
	Files     []InitFile `json:"files"`
}

func Init(ctx context.Context, options InitOptions) (InitPlan, error) {
	if err := options.Validate(); err != nil {
		return InitPlan{}, err
	}
	if err := validateInitDirectory(options.targetDirectory()); err != nil {
		return InitPlan{}, err
	}
	plan := InitPlan{
		Directory: options.targetDirectory(),
		Force:     options.Force,
		DryRun:    options.DryRun,
		Files:     make([]InitFile, 0, len(initFiles(options.targetDirectory()))),
	}
	for _, file := range initFiles(options.targetDirectory()) {
		select {
		case <-ctx.Done():
			return InitPlan{}, ctx.Err()
		default:
		}
		result, err := writeInitFile(file, options)
		if err != nil {
			return InitPlan{}, err
		}
		plan.Files = append(plan.Files, result)
	}
	return plan, nil
}

func validateInitDirectory(directory string) error {
	info, err := os.Lstat(directory)
	switch {
	case err == nil:
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("init directory cannot be a symlink: %s", directory)
		}
		if !info.IsDir() {
			return fmt.Errorf("init directory is not a directory: %s", directory)
		}
		return nil
	case os.IsNotExist(err):
		return nil
	default:
		return fmt.Errorf("inspect %s: %w", directory, err)
	}
}

type plannedInitFile struct {
	path    string
	content string
}

func initFiles(directory string) []plannedInitFile {
	return []plannedInitFile{
		{
			path: filepath.Join(directory, "README.md"),
			content: `# playpub Workspace

This directory holds Google Play Console CLI helper files for this repository.

- Keep service account JSON outside source control.
- Prefer dry-run commands before live release or metadata changes.
- Commit workflow templates, metadata payloads, and localization files when they are meant to be shared.
`,
		},
		{
			path: filepath.Join(directory, "workflow.json"),
			content: `{
  "version": 1,
  "workflows": {
    "release-internal": [
      {
        "run": "playpub publish internal --package com.example.app --aab ./app-release.aab --dry-run"
      }
    ]
  }
}
`,
		},
	}
}

func writeInitFile(file plannedInitFile, options InitOptions) (InitFile, error) {
	result := InitFile{Path: file.path}
	info, statErr := os.Lstat(file.path)
	switch {
	case statErr == nil:
		if info.Mode()&os.ModeSymlink != 0 {
			return InitFile{}, fmt.Errorf("init file cannot be a symlink: %s", file.path)
		}
		result.Exists = true
	case os.IsNotExist(statErr):
	default:
		return InitFile{}, fmt.Errorf("inspect %s: %w", file.path, statErr)
	}

	if result.Exists && !options.Force {
		return result, nil
	}
	if options.DryRun {
		result.Written = !result.Exists || options.Force
		result.Overwrote = result.Exists && options.Force
		return result, nil
	}
	if err := os.MkdirAll(filepath.Dir(file.path), 0o755); err != nil {
		return InitFile{}, fmt.Errorf("create %s: %w", filepath.Dir(file.path), err)
	}
	if err := os.WriteFile(file.path, []byte(file.content), 0o644); err != nil {
		return InitFile{}, fmt.Errorf("write %s: %w", file.path, err)
	}
	result.Written = true
	result.Overwrote = result.Exists
	return result, nil
}
