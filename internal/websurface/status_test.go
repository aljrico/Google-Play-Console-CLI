package websurface

import "testing"

func TestCurrentStatusDocumentsBlockedBrowserSurface(t *testing.T) {
	status := CurrentStatus()
	if status.Status != "blocked" {
		t.Fatalf("Status = %q, want blocked", status.Status)
	}
	if status.Surface == "" || status.Reason == "" || len(status.Alternatives) == 0 {
		t.Fatalf("status is incomplete: %#v", status)
	}
}
