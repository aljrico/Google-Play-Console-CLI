package gcsreports

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/api/option"
	"google.golang.org/api/storage/v1"
)

func TestDownloadDryRunDoesNotRequireDownloader(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "earnings.zip")
	result, err := Download(context.Background(), nil, DownloadOptions{
		Bucket:     "pubsite_prod_rev_0123456789",
		Object:     "earnings/earnings_202605.zip",
		OutputPath: outputPath,
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	if !result.DryRun || result.Downloaded {
		t.Fatalf("result = %#v, want dry-run without download", result)
	}
	if result.Plan.Steps[0] != "download gs://pubsite_prod_rev_0123456789/earnings/earnings_202605.zip" {
		t.Fatalf("steps = %#v, want GCS download step", result.Plan.Steps)
	}
}

func TestDownloadWritesObjectToFile(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "stats.csv")
	downloader := fakeDownloader{content: "Date,Package name\n2026-05-01,com.example.app\n"}

	result, err := Download(context.Background(), downloader, DownloadOptions{
		Bucket:     "pubsite_prod_rev_0123456789",
		Object:     "stats/installs.csv",
		OutputPath: outputPath,
	})
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	if !result.Downloaded || result.BytesWritten != int64(len(downloader.content)) {
		t.Fatalf("result = %#v, want downloaded bytes", result)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != downloader.content {
		t.Fatalf("content = %q, want %q", string(content), downloader.content)
	}
}

func TestDownloadRejectsExistingFileWithoutForce(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "stats.csv")
	if err := os.WriteFile(outputPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := Download(context.Background(), nil, DownloadOptions{
		Bucket:     "bucket",
		Object:     "object.csv",
		OutputPath: outputPath,
		DryRun:     true,
	})
	if err == nil {
		t.Fatalf("Download() expected error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %v, want existing file error", err)
	}
}

func TestDownloadForceOverwritesExistingFile(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "stats.csv")
	if err := os.WriteFile(outputPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	_, err := Download(context.Background(), fakeDownloader{content: "new"}, DownloadOptions{
		Bucket:     "bucket",
		Object:     "object.csv",
		OutputPath: outputPath,
		Force:      true,
	})
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != "new" {
		t.Fatalf("content = %q, want new", string(content))
	}
}

func TestDownloadWithoutForceDoesNotClobberLateExistingFile(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "stats.csv")
	downloader := fakeDownloader{
		content: "new",
		beforeReturn: func() {
			if err := os.WriteFile(outputPath, []byte("racer"), 0o644); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}
		},
	}
	_, err := Download(context.Background(), downloader, DownloadOptions{
		Bucket:     "bucket",
		Object:     "object.csv",
		OutputPath: outputPath,
	})
	if err == nil {
		t.Fatalf("Download() expected error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %v, want existing file error", err)
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatalf("ReadFile() error = %v", readErr)
	}
	if string(content) != "racer" {
		t.Fatalf("content = %q, want racer", string(content))
	}
}

func TestDownloadRejectsInvalidOptions(t *testing.T) {
	tests := []DownloadOptions{
		{},
		{Bucket: "bucket", Object: "object.csv"},
		{Bucket: "bucket", OutputPath: "report.csv"},
		{Bucket: "bucket", Object: "object.csv", OutputPath: "report.txt"},
	}
	for _, options := range tests {
		_, err := Download(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("Download(%#v) expected error", options)
		}
	}
}

func TestClientDownloadObjectUsesStorageMediaEndpoint(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.URL.Query().Get("alt") != "media" {
			t.Fatalf("query = %s, want alt=media", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte("report"))
	}))
	t.Cleanup(server.Close)

	service, err := storage.NewService(
		context.Background(),
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL+"/"),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	client := Client{service: service}
	var builder strings.Builder
	bytesWritten, err := client.DownloadObject(context.Background(), "bucket-id", "stats/report.csv", &builder)
	if err != nil {
		t.Fatalf("DownloadObject() error = %v", err)
	}
	if gotPath != "/b/bucket-id/o/stats/report.csv" {
		t.Fatalf("path = %q, want storage object media path", gotPath)
	}
	if bytesWritten != 6 || builder.String() != "report" {
		t.Fatalf("bytes/content = %d/%q, want report", bytesWritten, builder.String())
	}
}

type fakeDownloader struct {
	content      string
	beforeReturn func()
}

func (d fakeDownloader) DownloadObject(ctx context.Context, bucket string, object string, writer io.Writer) (int64, error) {
	bytesWritten, err := io.WriteString(writer, d.content)
	if err != nil {
		return 0, err
	}
	if d.beforeReturn != nil {
		d.beforeReturn()
	}
	return int64(bytesWritten), nil
}
