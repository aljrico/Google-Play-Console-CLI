package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const DefaultDiscoveryURL = "https://androidpublisher.googleapis.com/$discovery/rest?version=v3"

type FetchOptions struct {
	DiscoveryURL string
	Resource     string
	Method       string
}

type Document struct {
	Name         string     `json:"name"`
	Version      string     `json:"version"`
	Revision     string     `json:"revision,omitempty"`
	Title        string     `json:"title,omitempty"`
	Description  string     `json:"description,omitempty"`
	DiscoveryURL string     `json:"discoveryUrl"`
	RootURL      string     `json:"rootUrl,omitempty"`
	ServicePath  string     `json:"servicePath,omitempty"`
	BaseURL      string     `json:"baseUrl,omitempty"`
	Resources    []Resource `json:"resources"`
}

type Resource struct {
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	Methods   []Method   `json:"methods,omitempty"`
	Resources []Resource `json:"resources,omitempty"`
}

type Method struct {
	Name        string      `json:"name"`
	ID          string      `json:"id,omitempty"`
	Path        string      `json:"path,omitempty"`
	HTTPMethod  string      `json:"httpMethod,omitempty"`
	Description string      `json:"description,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty"`
}

type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
}

type MethodSummary struct {
	Resource    string `json:"resource"`
	Method      string `json:"method"`
	ID          string `json:"id,omitempty"`
	HTTPMethod  string `json:"httpMethod,omitempty"`
	Path        string `json:"path,omitempty"`
	Description string `json:"description,omitempty"`
}

type discoveryDocument struct {
	Name        string                       `json:"name"`
	Version     string                       `json:"version"`
	Revision    string                       `json:"revision"`
	Title       string                       `json:"title"`
	Description string                       `json:"description"`
	RootURL     string                       `json:"rootUrl"`
	ServicePath string                       `json:"servicePath"`
	BaseURL     string                       `json:"baseUrl"`
	Resources   map[string]discoveryResource `json:"resources"`
}

type discoveryResource struct {
	Methods   map[string]discoveryMethod   `json:"methods"`
	Resources map[string]discoveryResource `json:"resources"`
}

type discoveryMethod struct {
	ID          string                        `json:"id"`
	Path        string                        `json:"path"`
	HTTPMethod  string                        `json:"httpMethod"`
	Description string                        `json:"description"`
	Parameters  map[string]discoveryParameter `json:"parameters"`
}

type discoveryParameter struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

func Fetch(ctx context.Context, client *http.Client, options FetchOptions) (Document, error) {
	discoveryURL, err := normalizedDiscoveryURL(options.DiscoveryURL)
	if err != nil {
		return Document{}, err
	}
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return Document{}, fmt.Errorf("build schema request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return Document{}, fmt.Errorf("fetch schema: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Document{}, fmt.Errorf("fetch schema: unexpected HTTP status %s", resp.Status)
	}

	var raw discoveryDocument
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&raw); err != nil {
		return Document{}, fmt.Errorf("decode schema: %w", err)
	}
	document := convertDocument(raw, discoveryURL)
	document.Resources = filterResources(document.Resources, options.Resource, options.Method)
	return document, nil
}

func normalizedDiscoveryURL(discoveryURL string) (string, error) {
	if strings.TrimSpace(discoveryURL) == "" {
		discoveryURL = DefaultDiscoveryURL
	}
	parsed, err := url.ParseRequestURI(discoveryURL)
	if err != nil {
		return "", fmt.Errorf("invalid discovery URL %q: %w", discoveryURL, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("invalid discovery URL %q: scheme must be http or https", discoveryURL)
	}
	return discoveryURL, nil
}

func convertDocument(raw discoveryDocument, discoveryURL string) Document {
	return Document{
		Name:         raw.Name,
		Version:      raw.Version,
		Revision:     raw.Revision,
		Title:        raw.Title,
		Description:  raw.Description,
		DiscoveryURL: discoveryURL,
		RootURL:      raw.RootURL,
		ServicePath:  raw.ServicePath,
		BaseURL:      raw.BaseURL,
		Resources:    convertResources("", raw.Resources),
	}
}

func convertResources(parentPath string, resources map[string]discoveryResource) []Resource {
	names := sortedKeys(resources)
	converted := make([]Resource, 0, len(names))
	for _, name := range names {
		path := joinPath(parentPath, name)
		raw := resources[name]
		converted = append(converted, Resource{
			Name:      name,
			Path:      path,
			Methods:   convertMethods(raw.Methods),
			Resources: convertResources(path, raw.Resources),
		})
	}
	return converted
}

func convertMethods(methods map[string]discoveryMethod) []Method {
	names := sortedKeys(methods)
	converted := make([]Method, 0, len(names))
	for _, name := range names {
		raw := methods[name]
		converted = append(converted, Method{
			Name:        name,
			ID:          raw.ID,
			Path:        raw.Path,
			HTTPMethod:  raw.HTTPMethod,
			Description: raw.Description,
			Parameters:  convertParameters(raw.Parameters),
		})
	}
	return converted
}

func convertParameters(parameters map[string]discoveryParameter) []Parameter {
	names := sortedKeys(parameters)
	converted := make([]Parameter, 0, len(names))
	for _, name := range names {
		raw := parameters[name]
		converted = append(converted, Parameter{
			Name:        name,
			Type:        raw.Type,
			Required:    raw.Required,
			Location:    raw.Location,
			Description: raw.Description,
		})
	}
	return converted
}

func filterResources(resources []Resource, resourceFilter string, methodFilter string) []Resource {
	resourceFilter = strings.TrimSpace(resourceFilter)
	methodFilter = strings.TrimSpace(methodFilter)
	if resourceFilter == "" && methodFilter == "" {
		return resources
	}
	return filterResourcesInScope(resources, resourceFilter, methodFilter, false)
}

func filterResourcesInScope(resources []Resource, resourceFilter string, methodFilter string, includeAllDescendants bool) []Resource {
	filtered := make([]Resource, 0, len(resources))
	for _, resource := range resources {
		exactResourceMatch := resourceFilter == "" || resource.Path == resourceFilter
		ancestorResourceMatch := resourceFilter != "" && strings.HasPrefix(resourceFilter, resource.Path+".")
		includeResourceSubtree := includeAllDescendants || exactResourceMatch
		childResources := filterResourcesInScope(resource.Resources, resourceFilter, methodFilter, includeResourceSubtree)
		methods := filterMethods(resource.Methods, methodFilter)
		methodMatches := methodFilter == "" || len(methods) > 0 || resourcesContainMethods(childResources)
		if includeResourceSubtree && methodMatches {
			resource.Methods = methods
			resource.Resources = childResources
			filtered = append(filtered, resource)
			continue
		}
		if ancestorResourceMatch && len(childResources) > 0 {
			resource.Methods = nil
			resource.Resources = childResources
			filtered = append(filtered, resource)
		}
	}
	return filtered
}

func filterMethods(methods []Method, methodFilter string) []Method {
	if methodFilter == "" {
		return methods
	}
	filtered := make([]Method, 0, len(methods))
	for _, method := range methods {
		if method.Name == methodFilter || method.ID == methodFilter {
			filtered = append(filtered, method)
		}
	}
	return filtered
}

func resourcesContainMethods(resources []Resource) bool {
	for _, resource := range resources {
		if len(resource.Methods) > 0 || resourcesContainMethods(resource.Resources) {
			return true
		}
	}
	return false
}

func MethodSummaries(document Document) []MethodSummary {
	summaries := []MethodSummary{}
	appendMethodSummaries(&summaries, document.Resources)
	return summaries
}

func appendMethodSummaries(summaries *[]MethodSummary, resources []Resource) {
	for _, resource := range resources {
		for _, method := range resource.Methods {
			*summaries = append(*summaries, MethodSummary{
				Resource:    resource.Path,
				Method:      method.Name,
				ID:          method.ID,
				HTTPMethod:  method.HTTPMethod,
				Path:        method.Path,
				Description: method.Description,
			})
		}
		appendMethodSummaries(summaries, resource.Resources)
	}
}

func sortedKeys[V any](items map[string]V) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func joinPath(parentPath string, name string) string {
	if parentPath == "" {
		return name
	}
	return parentPath + "." + name
}
