package websurface

type Status struct {
	Surface      string   `json:"surface"`
	Status       string   `json:"status"`
	Reason       string   `json:"reason"`
	Alternatives []string `json:"alternatives"`
}

func CurrentStatus() Status {
	return Status{
		Surface: "Play Console browser workflows",
		Status:  "blocked",
		Reason:  "Play Console browser automation is intentionally kept out of the Go CLI until a stable, testable automation contract exists.",
		Alternatives: []string{
			"use API-backed playpub commands where Android Publisher or Play Developer Reporting APIs cover the workflow",
			"use explicit operator-driven browser automation outside playpub for Console-only surfaces",
			"document missing public API coverage in docs/PARITY.md instead of adding brittle selectors",
		},
	}
}
