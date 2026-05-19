package gcsreports

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aljrico/Google-Play-Console-CLI/internal/googleclient"
	"google.golang.org/api/option"
	"google.golang.org/api/storage/v1"
)

type DownloadOptions struct {
	Bucket     string `json:"bucket"`
	Object     string `json:"object"`
	OutputPath string `json:"outputPath"`
	Force      bool   `json:"force"`
	DryRun     bool   `json:"dryRun"`
}

func (o DownloadOptions) Validate() error {
	if strings.TrimSpace(o.Bucket) == "" {
		return fmt.Errorf("bucket is required")
	}
	if strings.TrimSpace(o.Object) == "" {
		return fmt.Errorf("object is required")
	}
	if strings.TrimSpace(o.OutputPath) == "" {
		return fmt.Errorf("output path is required")
	}
	if err := validateReportExtension(o.OutputPath); err != nil {
		return err
	}
	return nil
}

type DownloadPlan struct {
	Bucket     string   `json:"bucket"`
	Object     string   `json:"object"`
	OutputPath string   `json:"outputPath"`
	Force      bool     `json:"force"`
	Steps      []string `json:"steps"`
}

type DownloadResult struct {
	Bucket       string       `json:"bucket"`
	Object       string       `json:"object"`
	OutputPath   string       `json:"outputPath"`
	DryRun       bool         `json:"dryRun"`
	Downloaded   bool         `json:"downloaded"`
	BytesWritten int64        `json:"bytesWritten,omitempty"`
	Plan         DownloadPlan `json:"plan"`
}

type ObjectDownloader interface {
	DownloadObject(ctx context.Context, bucket string, object string, writer io.Writer) (int64, error)
}

type Client struct {
	service *storage.Service
}

func NewClientFromActiveProfile(ctx context.Context) (*Client, error) {
	httpClient, err := googleclient.ActiveProfileHTTPClient(ctx, storage.DevstorageReadOnlyScope)
	if err != nil {
		return nil, err
	}
	service, err := storage.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create Cloud Storage service: %w", err)
	}
	return &Client{service: service}, nil
}

func (c *Client) DownloadObject(ctx context.Context, bucket string, object string, writer io.Writer) (int64, error) {
	response, err := c.service.Objects.Get(bucket, object).Context(ctx).Download()
	if err != nil {
		return 0, fmt.Errorf("download gs://%s/%s: %w", bucket, object, err)
	}
	defer response.Body.Close()
	bytesWritten, err := io.Copy(writer, response.Body)
	if err != nil {
		return 0, fmt.Errorf("write gs://%s/%s: %w", bucket, object, err)
	}
	return bytesWritten, nil
}

func NewDownloadPlan(options DownloadOptions) (DownloadPlan, error) {
	if err := options.Validate(); err != nil {
		return DownloadPlan{}, err
	}
	steps := []string{fmt.Sprintf("download gs://%s/%s", options.Bucket, options.Object)}
	if options.Force {
		steps = append(steps, "overwrite output file")
	} else {
		steps = append(steps, "create output file")
	}
	return DownloadPlan{
		Bucket:     options.Bucket,
		Object:     options.Object,
		OutputPath: options.OutputPath,
		Force:      options.Force,
		Steps:      steps,
	}, nil
}

func Download(ctx context.Context, downloader ObjectDownloader, options DownloadOptions) (DownloadResult, error) {
	plan, err := NewDownloadPlan(options)
	if err != nil {
		return DownloadResult{}, err
	}
	result := DownloadResult{
		Bucket:     options.Bucket,
		Object:     options.Object,
		OutputPath: options.OutputPath,
		DryRun:     options.DryRun,
		Downloaded: false,
		Plan:       plan,
	}
	if err := ValidateOutputPath(options.OutputPath, options.Force); err != nil {
		return DownloadResult{}, err
	}
	if options.DryRun {
		return result, nil
	}
	if downloader == nil {
		return DownloadResult{}, fmt.Errorf("report downloader is required")
	}
	bytesWritten, err := downloadToFile(ctx, downloader, options)
	if err != nil {
		return DownloadResult{}, err
	}
	result.Downloaded = true
	result.BytesWritten = bytesWritten
	return result, nil
}

func ValidateOutputPath(path string, force bool) error {
	if path == "" {
		return fmt.Errorf("output path is required")
	}
	if err := validateReportExtension(path); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("output path %s is a symlink", path)
		}
		if info.IsDir() {
			return fmt.Errorf("output path %s is a directory", path)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("output path %s is not a regular file", path)
		}
		if !force {
			return fmt.Errorf("output path %s already exists; pass --force to overwrite", path)
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect output path %s: %w", path, err)
	}
	parent := filepath.Dir(path)
	parentInfo, err := os.Stat(parent)
	if err != nil {
		return fmt.Errorf("inspect output directory %s: %w", parent, err)
	}
	if !parentInfo.IsDir() {
		return fmt.Errorf("output directory %s is not a directory", parent)
	}
	return nil
}

func downloadToFile(ctx context.Context, downloader ObjectDownloader, options DownloadOptions) (int64, error) {
	directory := filepath.Dir(options.OutputPath)
	tempFile, err := os.CreateTemp(directory, ".playpub-report-*")
	if err != nil {
		return 0, fmt.Errorf("create temporary report file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	bytesWritten, downloadErr := downloader.DownloadObject(ctx, options.Bucket, options.Object, tempFile)
	closeErr := tempFile.Close()
	if downloadErr != nil {
		return 0, downloadErr
	}
	if closeErr != nil {
		return 0, fmt.Errorf("close temporary report file: %w", closeErr)
	}
	if options.Force {
		if err := os.Remove(options.OutputPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return 0, fmt.Errorf("remove existing report file %s: %w", options.OutputPath, err)
		}
		if err := os.Rename(tempPath, options.OutputPath); err != nil {
			return 0, fmt.Errorf("move report file to %s: %w", options.OutputPath, err)
		}
		return bytesWritten, nil
	}
	if err := os.Link(tempPath, options.OutputPath); err != nil {
		if errors.Is(err, os.ErrExist) {
			return 0, fmt.Errorf("output path %s already exists; pass --force to overwrite", options.OutputPath)
		}
		return 0, fmt.Errorf("move report file to %s: %w", options.OutputPath, err)
	}
	return bytesWritten, nil
}

func validateReportExtension(path string) error {
	switch filepath.Ext(path) {
	case ".csv", ".zip":
		return nil
	default:
		return fmt.Errorf("output path must end with .csv or .zip")
	}
}
