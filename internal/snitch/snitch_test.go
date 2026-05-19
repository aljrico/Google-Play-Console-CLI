package snitch

import (
	"net/url"
	"strings"
	"testing"
)

func TestBuildReportCreatesDeterministicIssueURL(t *testing.T) {
	report, err := BuildReport(ReportOptions{
		Title:   "Confusing release output",
		Body:    "The track summary was hard to read.",
		Command: "playpub releases list --package com.example.app",
		Labels:  []string{"ux", "snitch", "bug"},
	})
	if err != nil {
		t.Fatalf("BuildReport() error = %v", err)
	}
	if report.Repository != DefaultRepository {
		t.Fatalf("Repository = %q, want default", report.Repository)
	}
	if strings.Join(report.Labels, ",") != "bug,snitch,ux" {
		t.Fatalf("Labels = %#v, want sorted unique labels", report.Labels)
	}
	parsedURL, err := url.Parse(report.IssueURL)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	query := parsedURL.Query()
	if query.Get("title") != "Confusing release output" {
		t.Fatalf("title query = %q", query.Get("title"))
	}
	if !strings.Contains(query.Get("body"), "playpub releases list") {
		t.Fatalf("body query = %q, want command", query.Get("body"))
	}
	if query.Get("labels") != "bug,snitch,ux" {
		t.Fatalf("labels query = %q", query.Get("labels"))
	}
}

func TestBuildReportValidatesTitle(t *testing.T) {
	_, err := BuildReport(ReportOptions{Title: " "})
	if err == nil {
		t.Fatal("BuildReport() error = nil, want title validation")
	}
}

func TestBuildReportValidatesRepository(t *testing.T) {
	for _, repository := range []string{"bad", "../owner/repo", "owner/repo/issues", "./repo", "owner/.", "owner/..", "./."} {
		_, err := BuildReport(ReportOptions{Repository: repository, Title: "Friction"})
		if err == nil {
			t.Fatalf("BuildReport(%q) error = nil, want repository validation", repository)
		}
	}
}

func TestBuildReportAllowsDottedRepositoryName(t *testing.T) {
	report, err := BuildReport(ReportOptions{Repository: "owner/foo.bar", Title: "Friction"})
	if err != nil {
		t.Fatalf("BuildReport() error = %v", err)
	}
	if report.Repository != "owner/foo.bar" {
		t.Fatalf("Repository = %q, want dotted repo", report.Repository)
	}
}

func TestBuildReportRejectsCommaLabels(t *testing.T) {
	_, err := BuildReport(ReportOptions{Title: "Friction", Labels: []string{"bug,security"}})
	if err == nil {
		t.Fatal("BuildReport() error = nil, want label validation")
	}
}
