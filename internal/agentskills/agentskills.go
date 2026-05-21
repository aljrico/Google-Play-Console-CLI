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
	Name           string `json:"name"`
	Path           string `json:"path"`
	Exists         bool   `json:"exists"`
	Written        bool   `json:"written"`
	Overwrote      bool   `json:"overwrote"`
	WouldWrite     bool   `json:"wouldWrite"`
	WouldOverwrite bool   `json:"wouldOverwrite"`
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

type skillInstallPlan struct {
	skill  bundledSkill
	result InstalledSkill
}

type skillMetadata struct {
	Name        string
	Description string
}

func List() []Skill {
	skills := allBundledSkills()
	result := make([]Skill, 0, len(skills))
	for _, skill := range skills {
		result = append(result, Skill{Name: skill.name, Description: skill.description})
	}
	return result
}

func BundledNames() []string {
	skills := allBundledSkills()
	names := make([]string, 0, len(skills))
	for _, skill := range skills {
		names = append(names, skill.name)
	}
	return names
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
	plans := make([]skillInstallPlan, 0, len(skills))
	for _, skill := range skills {
		plan, err := planSkillInstall(directory, skill, options)
		if err != nil {
			return InstallResult{}, err
		}
		plans = append(plans, plan)
	}

	result := InstallResult{
		Directory: directory,
		Force:     options.Force,
		DryRun:    options.DryRun,
		Skills:    make([]InstalledSkill, 0, len(plans)),
	}
	for _, plan := range plans {
		select {
		case <-ctx.Done():
			return InstallResult{}, ctx.Err()
		default:
		}
		installedSkill := plan.result
		if !options.DryRun && installedSkill.WouldWrite {
			writtenSkill, err := writeSkill(plan)
			if err != nil {
				return InstallResult{}, err
			}
			installedSkill = writtenSkill
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

func planSkillInstall(directory string, skill bundledSkill, options InstallOptions) (skillInstallPlan, error) {
	path := filepath.Join(directory, skill.name, "SKILL.md")
	if err := validateSkillDirectory(filepath.Dir(path)); err != nil {
		return skillInstallPlan{}, err
	}

	result := InstalledSkill{Name: skill.name, Path: path}
	info, err := os.Lstat(path)
	switch {
	case err == nil:
		if info.Mode()&os.ModeSymlink != 0 {
			return skillInstallPlan{}, fmt.Errorf("skill file cannot be a symlink: %s", path)
		}
		if info.IsDir() {
			return skillInstallPlan{}, fmt.Errorf("skill file is a directory: %s", path)
		}
		result.Exists = true
	case os.IsNotExist(err):
	default:
		return skillInstallPlan{}, fmt.Errorf("inspect %s: %w", path, err)
	}

	if result.Exists && !options.Force {
		return skillInstallPlan{skill: skill, result: result}, nil
	}
	result.WouldWrite = true
	result.WouldOverwrite = result.Exists && options.Force
	return skillInstallPlan{skill: skill, result: result}, nil
}

func validateSkillDirectory(directory string) error {
	info, err := os.Lstat(directory)
	switch {
	case err == nil:
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("skill directory cannot be a symlink: %s", directory)
		}
		if !info.IsDir() {
			return fmt.Errorf("skill directory is not a directory: %s", directory)
		}
		return nil
	case os.IsNotExist(err):
		return nil
	default:
		return fmt.Errorf("inspect %s: %w", directory, err)
	}
}

func writeSkill(plan skillInstallPlan) (InstalledSkill, error) {
	path := plan.result.Path
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return InstalledSkill{}, fmt.Errorf("create %s: %w", directory, err)
	}
	if err := validateSkillDirectory(directory); err != nil {
		return InstalledSkill{}, err
	}

	tempFile, err := os.CreateTemp(directory, ".SKILL.md.*")
	if err != nil {
		return InstalledSkill{}, fmt.Errorf("create temp skill file in %s: %w", directory, err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)
	if _, err := tempFile.WriteString(plan.skill.content); err != nil {
		tempFile.Close()
		return InstalledSkill{}, fmt.Errorf("write %s: %w", tempPath, err)
	}
	if err := tempFile.Chmod(0o644); err != nil {
		tempFile.Close()
		return InstalledSkill{}, fmt.Errorf("chmod %s: %w", tempPath, err)
	}
	if err := tempFile.Close(); err != nil {
		return InstalledSkill{}, fmt.Errorf("close %s: %w", tempPath, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return InstalledSkill{}, fmt.Errorf("write %s: %w", path, err)
	}

	result := plan.result
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
			return nil, fmt.Errorf("unknown skill %q; valid skills: %s", name, strings.Join(skillNames(skills), ", "))
		}
		selected = append(selected, skill)
	}
	return selected, nil
}

func skillNames(skills []bundledSkill) []string {
	names := make([]string, 0, len(skills))
	for _, skill := range skills {
		names = append(names, skill.name)
	}
	sort.Strings(names)
	return names
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
		metadata, err := parseSkillMetadata(string(content))
		if err != nil {
			panic(fmt.Sprintf("parse bundled skill %s: %v", name, err))
		}
		if metadata.Name != name {
			panic(fmt.Sprintf("bundled skill %s declares name %s", name, metadata.Name))
		}
		skills = append(skills, bundledSkill{
			name:        name,
			description: metadata.Description,
			content:     string(content),
		})
	}
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].name < skills[j].name
	})
	return skills
}

func parseSkillMetadata(content string) (skillMetadata, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 4 || lines[0] != "---" {
		return skillMetadata{}, fmt.Errorf("missing skill frontmatter")
	}

	metadata := skillMetadata{}
	closed := false
	for _, line := range lines[1:] {
		if line == "---" {
			closed = true
			break
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "name":
			metadata.Name = value
		case "description":
			metadata.Description = value
		}
	}
	if !closed {
		return skillMetadata{}, fmt.Errorf("unterminated skill frontmatter")
	}
	if metadata.Name == "" {
		return skillMetadata{}, fmt.Errorf("skill frontmatter name is required")
	}
	if metadata.Description == "" {
		return skillMetadata{}, fmt.Errorf("skill frontmatter description is required")
	}
	return metadata, nil
}
