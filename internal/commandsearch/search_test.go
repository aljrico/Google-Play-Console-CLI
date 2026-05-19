package commandsearch

import "testing"

func TestSearchRanksPathMatches(t *testing.T) {
	result, err := Search([]Document{
		{Path: "playpub releases upload", Use: "playpub releases upload [flags]", Short: "Upload an APK or Android App Bundle"},
		{Path: "playpub images upload", Use: "playpub images upload [flags]", Short: "Upload one store image"},
	}, Options{Query: "releases upload"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(result.Matches) != 1 {
		t.Fatalf("matches = %#v, want one release match", result.Matches)
	}
	if result.Matches[0].Path != "playpub releases upload" {
		t.Fatalf("first match = %#v", result.Matches[0])
	}
}

func TestSearchMatchesFlags(t *testing.T) {
	result, err := Search([]Document{
		{Path: "playpub notify send", Use: "playpub notify send [flags]", Short: "Send a webhook", Flags: []string{"--webhook-url-env", "--message"}},
	}, Options{Query: "--webhook-url-env"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(result.Matches) != 1 {
		t.Fatalf("matches = %#v, want flag match", result.Matches)
	}
}

func TestSearchScoresExactFlagsAbovePathSubstrings(t *testing.T) {
	result, err := Search([]Document{
		{Path: "playpub one-o", Use: "playpub one-o", Short: "Hyphenated path"},
		{Path: "playpub root", Use: "playpub root", Short: "Root command", Flags: []string{"-o", "--output"}},
	}, Options{Query: "-o"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(result.Matches) == 0 || result.Matches[0].Path != "playpub root" {
		t.Fatalf("matches = %#v, want exact flag first", result.Matches)
	}
}

func TestSearchValidatesQuery(t *testing.T) {
	_, err := Search(nil, Options{Query: " "})
	if err == nil {
		t.Fatal("Search() error = nil, want validation")
	}
}

func TestSearchLimit(t *testing.T) {
	result, err := Search([]Document{
		{Path: "playpub one", Use: "playpub one", Short: "Release command"},
		{Path: "playpub two", Use: "playpub two", Short: "Release command"},
	}, Options{Query: "release", Limit: 1})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(result.Matches) != 1 {
		t.Fatalf("matches = %#v, want one match", result.Matches)
	}
}

func TestSearchZeroLimitReturnsAllMatches(t *testing.T) {
	result, err := Search([]Document{
		{Path: "playpub one", Use: "playpub one", Short: "Release command"},
		{Path: "playpub two", Use: "playpub two", Short: "Release command"},
	}, Options{Query: "release", Limit: 0})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(result.Matches) != 2 {
		t.Fatalf("matches = %#v, want all matches", result.Matches)
	}
}
