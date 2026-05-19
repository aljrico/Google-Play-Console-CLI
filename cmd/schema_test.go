package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSchemaOutputsDiscoverySummaryWithoutAuth(t *testing.T) {
	t.Setenv("PLAYPUB_CONFIG", t.TempDir()+"/missing-config.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
		  "name": "androidpublisher",
		  "version": "v3",
		  "resources": {
		    "edits": {
		      "resources": {
		        "tracks": {
		          "methods": {
		            "list": {
		              "id": "androidpublisher.edits.tracks.list",
		              "path": "androidpublisher/v3/applications/{packageName}/edits/{editId}/tracks",
		              "httpMethod": "GET"
		            }
		          }
		        }
		      }
		    }
		  }
		}`))
	}))
	defer server.Close()
	previousDiscoveryURL := schemaDiscoveryURL
	schemaDiscoveryURL = server.URL
	defer func() {
		schemaDiscoveryURL = previousDiscoveryURL
	}()

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"schema",
		"--resource",
		"edits.tracks",
		"--method",
		"list",
		"--output",
		"json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		`"name":"androidpublisher"`,
		`"path":"edits.tracks"`,
		`"id":"androidpublisher.edits.tracks.list"`,
		`"httpMethod":"GET"`,
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
	if strings.Contains(output, "no active auth profile") {
		t.Fatalf("output = %s, did not expect auth", output)
	}
}

func TestSchemaOutputsFlatMarkdownSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
		  "name": "androidpublisher",
		  "version": "v3",
		  "resources": {
		    "edits": {
		      "resources": {
		        "tracks": {
		          "methods": {
		            "list": {
		              "id": "androidpublisher.edits.tracks.list",
		              "path": "androidpublisher/v3/applications/{packageName}/edits/{editId}/tracks",
		              "httpMethod": "GET"
		            }
		          }
		        }
		      }
		    }
		  }
		}`))
	}))
	defer server.Close()
	previousDiscoveryURL := schemaDiscoveryURL
	schemaDiscoveryURL = server.URL
	defer func() {
		schemaDiscoveryURL = previousDiscoveryURL
	}()

	var buf bytes.Buffer
	cmd := newRootCommand(&buf)
	cmd.SetArgs([]string{
		"schema",
		"--resource",
		"edits",
		"--method",
		"list",
		"--output",
		"markdown",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := buf.String()
	for _, want := range []string{
		"| resource | method | id | httpMethod | path | description |",
		"edits.tracks",
		"androidpublisher.edits.tracks.list",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output = %s, want %s", output, want)
		}
	}
}
