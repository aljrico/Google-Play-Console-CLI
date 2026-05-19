package diff

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompareJSONFilesReportsStableChanges(t *testing.T) {
	from := writeJSON(t, "from.json", `{
	  "title": "Old",
	  "listings": {
	    "en-US": {
	      "shortDescription": "Old short",
	      "video": "https://example.com/old"
	    }
	  },
	  "screenshots": ["one.png", "two.png"]
	}`)
	to := writeJSON(t, "to.json", `{
	  "title": "New",
	  "listings": {
	    "en-US": {
	      "shortDescription": "Old short"
	    },
	    "es-ES": {
	      "shortDescription": "Nuevo"
	    }
	  },
	  "screenshots": ["one.png", "three.png", "four.png"]
	}`)

	result, err := CompareJSONFiles(JSONOptions{FromPath: from, ToPath: to})
	if err != nil {
		t.Fatalf("CompareJSONFiles() error = %v", err)
	}
	if result.Equal {
		t.Fatal("Equal = true, want false")
	}
	wantPaths := []string{
		"/listings/en-US/video",
		"/listings/es-ES",
		"/screenshots/1",
		"/screenshots/2",
		"/title",
	}
	if len(result.Changes) != len(wantPaths) {
		t.Fatalf("len(Changes) = %d, want %d: %#v", len(result.Changes), len(wantPaths), result.Changes)
	}
	for index, wantPath := range wantPaths {
		if result.Changes[index].Path != wantPath {
			t.Fatalf("Changes[%d].Path = %q, want %q", index, result.Changes[index].Path, wantPath)
		}
	}
	if result.Changes[0].Kind != ChangeRemoved {
		t.Fatalf("first kind = %q, want removed", result.Changes[0].Kind)
	}
	if result.Changes[1].Kind != ChangeAdded {
		t.Fatalf("second kind = %q, want added", result.Changes[1].Kind)
	}
	if result.Changes[4].Kind != ChangeChanged {
		t.Fatalf("last kind = %q, want changed", result.Changes[4].Kind)
	}
}

func TestCompareJSONFilesReportsEqual(t *testing.T) {
	from := writeJSON(t, "from.json", `{"title":"Same"}`)
	to := writeJSON(t, "to.json", `{"title":"Same"}`)

	result, err := CompareJSONFiles(JSONOptions{FromPath: from, ToPath: to})
	if err != nil {
		t.Fatalf("CompareJSONFiles() error = %v", err)
	}
	if !result.Equal {
		t.Fatalf("Equal = false, changes = %#v", result.Changes)
	}
}

func TestCompareJSONFilesReportsEqualScalarRoots(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "number", value: `1`},
		{name: "string", value: `"same"`},
		{name: "null", value: `null`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			from := writeJSON(t, "from.json", tc.value)
			to := writeJSON(t, "to.json", tc.value)

			result, err := CompareJSONFiles(JSONOptions{FromPath: from, ToPath: to})
			if err != nil {
				t.Fatalf("CompareJSONFiles() error = %v", err)
			}
			if !result.Equal {
				t.Fatalf("Equal = false, changes = %#v", result.Changes)
			}
			if result.Changes == nil {
				t.Fatal("Changes = nil, want empty slice")
			}
			if len(result.Changes) != 0 {
				t.Fatalf("len(Changes) = %d, want 0", len(result.Changes))
			}
		})
	}
}

func TestCompareJSONFilesTreatsNumericSpellingsAsEqual(t *testing.T) {
	from := writeJSON(t, "from.json", `{"price":1,"zero":-0}`)
	to := writeJSON(t, "to.json", `{"price":1.00,"zero":0}`)

	result, err := CompareJSONFiles(JSONOptions{FromPath: from, ToPath: to})
	if err != nil {
		t.Fatalf("CompareJSONFiles() error = %v", err)
	}
	if !result.Equal {
		t.Fatalf("Equal = false, changes = %#v", result.Changes)
	}
}

func TestCompareJSONFilesUsesEmptyRootPointer(t *testing.T) {
	from := writeJSON(t, "from.json", `null`)
	to := writeJSON(t, "to.json", `{"x":1}`)

	result, err := CompareJSONFiles(JSONOptions{FromPath: from, ToPath: to})
	if err != nil {
		t.Fatalf("CompareJSONFiles() error = %v", err)
	}
	if len(result.Changes) != 1 {
		t.Fatalf("len(Changes) = %d, want 1", len(result.Changes))
	}
	if result.Changes[0].Path != "" {
		t.Fatalf("Path = %q, want empty root JSON Pointer", result.Changes[0].Path)
	}
}

func TestCompareJSONFilesDistinguishesEmptyKeyFromRoot(t *testing.T) {
	from := writeJSON(t, "from.json", `{"":"old"}`)
	to := writeJSON(t, "to.json", `{"":"new"}`)

	result, err := CompareJSONFiles(JSONOptions{FromPath: from, ToPath: to})
	if err != nil {
		t.Fatalf("CompareJSONFiles() error = %v", err)
	}
	if len(result.Changes) != 1 {
		t.Fatalf("len(Changes) = %d, want 1", len(result.Changes))
	}
	if result.Changes[0].Path != "/" {
		t.Fatalf("Path = %q, want empty-key JSON Pointer", result.Changes[0].Path)
	}
}

func TestCompareJSONFilesEscapesJSONPointerTokens(t *testing.T) {
	from := writeJSON(t, "from.json", `{"a/b":{"~key":"old"}}`)
	to := writeJSON(t, "to.json", `{"a/b":{"~key":"new"}}`)

	result, err := CompareJSONFiles(JSONOptions{FromPath: from, ToPath: to})
	if err != nil {
		t.Fatalf("CompareJSONFiles() error = %v", err)
	}
	if len(result.Changes) != 1 {
		t.Fatalf("len(Changes) = %d, want 1", len(result.Changes))
	}
	if result.Changes[0].Path != "/a~1b/~0key" {
		t.Fatalf("Path = %q, want escaped JSON pointer", result.Changes[0].Path)
	}
}

func TestCompareJSONFilesRejectsTrailingJSON(t *testing.T) {
	from := writeJSON(t, "from.json", `{"title":"Old"} {"extra":true}`)
	to := writeJSON(t, "to.json", `{"title":"New"}`)

	_, err := CompareJSONFiles(JSONOptions{FromPath: from, ToPath: to})
	if err == nil {
		t.Fatal("CompareJSONFiles() error = nil, want trailing JSON error")
	}
}

func writeJSON(t *testing.T, name string, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
