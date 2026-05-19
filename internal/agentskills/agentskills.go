package agentskills

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed bundled/*/SKILL.md
var bundledSkills embed.FS

const defaultRelativeDirectory = ".agents/skills"

type InstallOptions struct {
	Directory string   `json:"directory,omitempty"`
	Skills    []string `json:"skills,omitempty"`
	Force     bool     `json:"force"`
	DryRun    bool     `json:"dryRun"`
}

func (o InstallOptions) targetDirectory() (string, error) {
	if o.Directory != "" {
		return o.Directory, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, defaultRelativeDirectory), nil
}

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type InstalledSkill struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Exists    bool   `json:"exists"`
	Written   bool   `json:"written"`
	Overwrote bool   `json:"overwrote"`
}

type InstallResult struct {
	Directory string           `json:"directory"`
	Force     bool             `json:"force"`
	DryRun    bool             `json:"dryRun"`
	Skills    []InstalledSkill `json:"skills"`
}

type bundledSkill struct {
	name        string
	description string
	content     string
}

func List() []Skill {
	skills := allBundledSkills()
	result := make([]Skill, 0, len(skills))
	for _, skill := range skills {
		result = append(result, Skill{Name: skill.name, Description: skill.description})
	}
	return result
}

func Install(ctx context.Context, options InstallOptions) (InstallResult, error) {
	directory, err := options.targetDirectory()
	if err != nil {
		return InstallResult{}, err
	}
	if err := validateDirectory(directory); err != nil {
		return InstallResult{}, err
	}
	skills, err := selectedBundledSkills(options.Skills)
	if err != nil {
		return InstallResult{}, err
	}
	result := InstallResult{
		Directory: directory,
		Force:     options.Force,
		DryRun:    options.DryRun,
		Skills:    make([]InstalledSkill, 0, len(skills)),
	}
	for _, skill := range skills {
		select {
		case <-ctx.Done():
			return InstallResult{}, ctx.Err()
		default:
		}
		installedSkill, err := installSkill(directory, skill, options)
		if err != nil {
			return InstallResult{}, err
		}
		result.Skills = append(result.Skills, installedSkill)
	}
	return result, nil
}

func validateDirectory(directory string) error {
	info, err := os.Lstat(directory)
	switch {
	case err == nil:
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("skills directory cannot be a symlink: %s", directory)
		}
		if !info.IsDir() {
			return fmt.Errorf("skills directory is not a directory: %s", directory)
		}
		return nil
	case os.IsNotExist(err):
		return nil
	default:
		return fmt.Errorf("inspect %s: %w", directory, err)
	}
}

func installSkill(directory string, skill bundledSkill, options InstallOptions) (InstalledSkill, error) {
	path := filepath.Join(directory, skill.name, "SKILL.md")
	result := InstalledSkill{Name: skill.name, Path: path}
	info, err := os.Lstat(path)
	switch {
	case err == nil:
		if info.Mode()&os.ModeSymlink != 0 {
			return InstalledSkill{}, fmt.Errorf("skill file cannot be a symlink: %s", path)
		}
		if info.IsDir() {
			return InstalledSkill{}, fmt.Errorf("skill file is a directory: %s", path)
		}
		result.Exists = true
	case os.IsNotExist(err):
	default:
		return InstalledSkill{}, fmt.Errorf("inspect %s: %w", path, err)
	}

	if result.Exists && !options.Force {
		return result, nil
	}
	if options.DryRun {
		result.Written = !result.Exists || options.Force
		result.Overwrote = result.Exists && options.Force
		return result, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return InstalledSkill{}, fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(skill.content), 0o644); err != nil {
		return InstalledSkill{}, fmt.Errorf("write %s: %w", path, err)
	}
	result.Written = true
	result.Overwrote = result.Exists
	return result, nil
}

func selectedBundledSkills(names []string) ([]bundledSkill, error) {
	skills := allBundledSkills()
	if len(names) == 0 {
		return skills, nil
	}
	byName := make(map[string]bundledSkill, len(skills))
	for _, skill := range skills {
		byName[skill.name] = skill
	}
	selected := make([]bundledSkill, 0, len(names))
	for _, name := range names {
		skill, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("unknown skill %q", name)
		}
		selected = append(selected, skill)
	}
	return selected, nil
}

func allBundledSkills() []bundledSkill {
	entries, err := bundledSkills.ReadDir("bundled")
	if err != nil {
		panic(fmt.Sprintf("read bundled skills: %v", err))
	}
	skills := make([]bundledSkill, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		content, err := bundledSkills.ReadFile(filepath.ToSlash(filepath.Join("bundled", name, "SKILL.md")))
		if err != nil {
			panic(fmt.Sprintf("read bundled skill %s: %v", name, err))
		}
		skills = append(skills, bundledSkill{
			name:        name,
			description: firstHeading(string(content)),
			content:     string(content),
		})
	}
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].name < skills[j].name
	})
	return skills
}

func firstHeading(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if len(line) > 2 && line[0] == '#' && line[1] == ' ' {
			return line[2:]
		}
	}
	return ""
}
