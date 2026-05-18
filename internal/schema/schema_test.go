package schema

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testDiscoveryDocument = `{
  "name": "androidpublisher",
  "version": "v3",
  "revision": "20260518",
  "title": "Google Play Android Developer API",
  "description": "Lets Android application developers access their Google Play accounts.",
  "rootUrl": "https://androidpublisher.googleapis.com/",
  "servicePath": "",
  "baseUrl": "https://androidpublisher.googleapis.com/",
  "resources": {
    "edits": {
      "resources": {
        "tracks": {
          "methods": {
            "list": {
              "id": "androidpublisher.edits.tracks.list",
              "path": "androidpublisher/v3/applications/{packageName}/edits/{editId}/tracks",
              "httpMethod": "GET",
              "description": "Lists all tracks.",
              "parameters": {
                "editId": {
                  "type": "string",
                  "required": true,
                  "location": "path",
                  "description": "Identifier of the edit."
                },
                "packageName": {
                  "type": "string",
                  "required": true,
                  "location": "path",
                  "description": "Package name of the app."
                }
              }
            }
          }
        }
      }
    },
    "reviews": {
      "methods": {
        "get": {
          "id": "androidpublisher.reviews.get",
          "path": "androidpublisher/v3/applications/{packageName}/reviews/{reviewId}",
          "httpMethod": "GET"
        }
      }
    }
  }
}`

func TestFetchReturnsSortedSchemaSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testDiscoveryDocument))
	}))
	defer server.Close()

	document, err := Fetch(context.Background(), server.Client(), FetchOptions{DiscoveryURL: server.URL})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if document.Name != "androidpublisher" {
		t.Fatalf("Name = %q, want androidpublisher", document.Name)
	}
	if len(document.Resources) != 2 {
		t.Fatalf("Resources len = %d, want 2", len(document.Resources))
	}
	if document.Resources[0].Path != "edits" || document.Resources[1].Path != "reviews" {
		t.Fatalf("resource paths = %q, %q; want sorted paths", document.Resources[0].Path, document.Resources[1].Path)
	}
	tracks := document.Resources[0].Resources[0]
	if tracks.Path != "edits.tracks" {
		t.Fatalf("tracks path = %q, want edits.tracks", tracks.Path)
	}
	if tracks.Methods[0].Parameters[0].Name != "editId" {
		t.Fatalf("first parameter = %q, want editId", tracks.Methods[0].Parameters[0].Name)
	}
}

func TestFetchFiltersResourceAndMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testDiscoveryDocument))
	}))
	defer server.Close()

	document, err := Fetch(context.Background(), server.Client(), FetchOptions{
		DiscoveryURL: server.URL,
		Resource:     "edits.tracks",
		Method:       "androidpublisher.edits.tracks.list",
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if len(document.Resources) != 1 {
		t.Fatalf("Resources len = %d, want 1", len(document.Resources))
	}
	if document.Resources[0].Path != "edits" {
		t.Fatalf("root resource = %q, want edits", document.Resources[0].Path)
	}
	tracks := document.Resources[0].Resources[0]
	if tracks.Path != "edits.tracks" {
		t.Fatalf("child resource = %q, want edits.tracks", tracks.Path)
	}
	if len(tracks.Methods) != 1 || tracks.Methods[0].Name != "list" {
		t.Fatalf("methods = %#v, want list only", tracks.Methods)
	}
}

func TestFetchParentResourceFilterKeepsMatchingDescendants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testDiscoveryDocument))
	}))
	defer server.Close()

	document, err := Fetch(context.Background(), server.Client(), FetchOptions{
		DiscoveryURL: server.URL,
		Resource:     "edits",
		Method:       "list",
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	if len(document.Resources) != 1 {
		t.Fatalf("Resources len = %d, want 1", len(document.Resources))
	}
	edits := document.Resources[0]
	if edits.Path != "edits" {
		t.Fatalf("resource path = %q, want edits", edits.Path)
	}
	if len(edits.Resources) != 1 {
		t.Fatalf("edits children len = %d, want tracks child", len(edits.Resources))
	}
	tracks := edits.Resources[0]
	if tracks.Path != "edits.tracks" {
		t.Fatalf("child path = %q, want edits.tracks", tracks.Path)
	}
	if len(tracks.Methods) != 1 || tracks.Methods[0].Name != "list" {
		t.Fatalf("tracks methods = %#v, want list only", tracks.Methods)
	}
}

func TestMethodSummariesFlattensFilteredDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testDiscoveryDocument))
	}))
	defer server.Close()

	document, err := Fetch(context.Background(), server.Client(), FetchOptions{
		DiscoveryURL: server.URL,
		Resource:     "edits",
		Method:       "list",
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	summaries := MethodSummaries(document)
	if len(summaries) != 1 {
		t.Fatalf("summaries len = %d, want 1", len(summaries))
	}
	if summaries[0].Resource != "edits.tracks" || summaries[0].Method != "list" {
		t.Fatalf("summary = %#v, want edits.tracks list", summaries[0])
	}
}

func TestFetchRejectsInvalidDiscoveryURL(t *testing.T) {
	_, err := Fetch(context.Background(), nil, FetchOptions{DiscoveryURL: "ftp://example.com/schema.json"})
	if err == nil {
		t.Fatal("Fetch() error = nil, want invalid URL error")
	}
}
