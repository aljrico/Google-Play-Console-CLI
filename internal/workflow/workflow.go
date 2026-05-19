package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const DefaultFile = ".gpc/workflow.json"

type Definition struct {
	Version   int               `json:"version"`
	Workflows map[string][]Step `json:"workflows"`
}

type Step struct {
	Name string `json:"name,omitempty"`
	Run  string `json:"run"`
}

type ListOptions struct {
	File string `json:"file"`
}

type RunOptions struct {
	File    string `json:"file"`
	Name    string `json:"name"`
	WorkDir string `json:"workDir,omitempty"`
	DryRun  bool   `json:"dryRun"`
	Confirm bool   `json:"confirm"`
}

type Summary struct {
	Name  string `json:"name"`
	Steps int    `json:"steps"`
}

type RunResult struct {
	Name    string       `json:"name"`
	File    string       `json:"file"`
	WorkDir string       `json:"workDir"`
	DryRun  bool         `json:"dryRun"`
	Confirm bool         `json:"confirm"`
	Success bool         `json:"success"`
	Steps   []StepResult `json:"steps"`
}

type StepResult struct {
	Index    int    `json:"index"`
	Name     string `json:"name,omitempty"`
	Run      string `json:"run"`
	Skipped  bool   `json:"skipped"`
	Success  bool   `json:"success"`
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	Error    string `json:"error,omitempty"`
}

type Runner interface {
	Run(ctx context.Context, step Step, directory string) StepResult
}

type ShellRunner struct{}

func List(options ListOptions) ([]Summary, error) {
	definition, err := Load(options.file())
	if err != nil {
		return nil, err
	}
	summaries := make([]Summary, 0, len(definition.Workflows))
	for name, steps := range definition.Workflows {
		summaries = append(summaries, Summary{Name: name, Steps: len(steps)})
	}
	sort.Slice(summaries, func(i int, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})
	return summaries, nil
}

func Run(ctx context.Context, runner Runner, options RunOptions) (RunResult, error) {
	if err := options.Validate(); err != nil {
		return RunResult{}, err
	}
	definition, err := Load(options.file())
	if err != nil {
		return RunResult{}, err
	}
	steps, ok := definition.Workflows[options.Name]
	if !ok {
		return RunResult{}, fmt.Errorf("workflow %q not found", options.Name)
	}
	if runner == nil {
		runner = ShellRunner{}
	}
	workDir, err := options.workDir()
	if err != nil {
		return RunResult{}, err
	}
	result := RunResult{
		Name:    options.Name,
		File:    options.file(),
		WorkDir: workDir,
		DryRun:  options.DryRun,
		Confirm: options.Confirm,
		Success: true,
		Steps:   make([]StepResult, 0, len(steps)),
	}
	for index, step := range steps {
		stepResult := StepResult{
			Index:    index + 1,
			Name:     step.Name,
			Run:      step.Run,
			Skipped:  options.DryRun,
			Success:  true,
			ExitCode: 0,
		}
		if !options.DryRun {
			stepResult = runner.Run(ctx, step, workDir)
			stepResult.Index = index + 1
		}
		result.Steps = append(result.Steps, stepResult)
		if !stepResult.Success {
			result.Success = false
			return result, fmt.Errorf("workflow %q step %d failed: %s", options.Name, stepResult.Index, stepResult.Error)
		}
	}
	return result, nil
}

func Load(path string) (Definition, error) {
	path = normalizedFile(path)
	content, err := os.ReadFile(path)
	if err != nil {
		return Definition{}, fmt.Errorf("read workflow file %s: %w", path, err)
	}
	var definition Definition
	dec := json.NewDecoder(bytes.NewReader(content))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&definition); err != nil {
		return Definition{}, fmt.Errorf("parse workflow file %s: %w", path, err)
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return Definition{}, fmt.Errorf("parse workflow file %s: %w", path, err)
		}
		return Definition{}, fmt.Errorf("parse workflow file %s: trailing JSON value", path)
	}
	if err := definition.Validate(); err != nil {
		return Definition{}, fmt.Errorf("validate workflow file %s: %w", path, err)
	}
	return definition, nil
}

func (d Definition) Validate() error {
	if d.Version != 1 {
		return fmt.Errorf("unsupported workflow version %d", d.Version)
	}
	if len(d.Workflows) == 0 {
		return fmt.Errorf("at least one workflow is required")
	}
	for name, steps := range d.Workflows {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			return fmt.Errorf("workflow name is required")
		}
		if name != trimmedName {
			return fmt.Errorf("workflow name %q cannot have leading or trailing whitespace", name)
		}
		if len(steps) == 0 {
			return fmt.Errorf("workflow %q requires at least one step", name)
		}
		for index, step := range steps {
			if strings.TrimSpace(step.Run) == "" {
				return fmt.Errorf("workflow %q step %d run command is required", name, index+1)
			}
		}
	}
	return nil
}

func (o RunOptions) Validate() error {
	if strings.TrimSpace(o.Name) == "" {
		return fmt.Errorf("workflow name is required")
	}
	if o.DryRun && o.Confirm {
		return fmt.Errorf("--confirm and --dry-run cannot be used together")
	}
	if !o.DryRun && !o.Confirm {
		return fmt.Errorf("workflow run requires --confirm or --dry-run")
	}
	_, err := o.workDir()
	return err
}

func (o ListOptions) file() string {
	return normalizedFile(o.File)
}

func (o RunOptions) file() string {
	return normalizedFile(o.File)
}

func (o RunOptions) workDir() (string, error) {
	if strings.TrimSpace(o.WorkDir) == "" {
		return defaultWorkDir(o.file())
	}
	info, err := os.Stat(o.WorkDir)
	if err != nil {
		return "", fmt.Errorf("inspect workflow directory %s: %w", o.WorkDir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workflow directory is not a directory: %s", o.WorkDir)
	}
	return o.WorkDir, nil
}

func (ShellRunner) Run(ctx context.Context, step Step, directory string) StepResult {
	result := StepResult{
		Name: step.Name,
		Run:  step.Run,
	}
	shellPath, shellArgs, err := shellCommand(step.Run)
	if err != nil {
		result.Success = false
		result.ExitCode = -1
		result.Error = err.Error()
		return result
	}
	cmd := exec.CommandContext(ctx, shellPath, shellArgs...)
	cmd.Dir = directory
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	result.ExitCode = 0
	result.Success = true
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}
	return result
}

func defaultWorkDir(file string) (string, error) {
	absoluteFile, err := filepath.Abs(file)
	if err != nil {
		return "", fmt.Errorf("resolve workflow file %s: %w", file, err)
	}
	directory := filepath.Dir(absoluteFile)
	if filepath.Base(directory) == ".gpc" {
		return filepath.Dir(directory), nil
	}
	return directory, nil
}

func shellCommand(command string) (string, []string, error) {
	if runtime.GOOS == "windows" {
		return "", nil, fmt.Errorf("workflow shell execution is not supported on windows yet")
	}
	return "/bin/sh", []string{"-c", command}, nil
}

func normalizedFile(path string) string {
	if strings.TrimSpace(path) == "" {
		return DefaultFile
	}
	return filepath.Clean(path)
}
