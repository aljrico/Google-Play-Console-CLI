package capabilities

import "testing"

func TestListParsesEmbeddedParityMatrix(t *testing.T) {
	items, err := List(ListOptions{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected capabilities")
	}
	if items[0].Section != "Getting Started" {
		t.Fatalf("first section = %q, want Getting Started", items[0].Section)
	}
}

func TestListFiltersByStatus(t *testing.T) {
	items, err := List(ListOptions{Status: StatusTested})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected tested capabilities")
	}
	for _, item := range items {
		if item.Status != StatusTested {
			t.Fatalf("status = %q, want tested", item.Status)
		}
	}
}

func TestListFiltersBySection(t *testing.T) {
	items, err := List(ListOptions{Section: "Monetization"})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected monetization capabilities")
	}
	for _, item := range items {
		if item.Section != "Monetization" {
			t.Fatalf("section = %q, want Monetization", item.Section)
		}
	}
}

func TestListRejectsUnsupportedStatus(t *testing.T) {
	_, err := List(ListOptions{Status: "done-ish"})
	if err == nil {
		t.Fatal("expected unsupported status error")
	}
}
