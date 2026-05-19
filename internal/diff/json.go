package diff

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type ChangeKind string

const (
	ChangeAdded   ChangeKind = "added"
	ChangeRemoved ChangeKind = "removed"
	ChangeChanged ChangeKind = "changed"
)

type JSONOptions struct {
	FromPath string `json:"fromPath"`
	ToPath   string `json:"toPath"`
}

type JSONResult struct {
	FromPath string       `json:"fromPath"`
	ToPath   string       `json:"toPath"`
	Equal    bool         `json:"equal"`
	Changes  []JSONChange `json:"changes"`
}

type JSONChange struct {
	Path   string           `json:"path"`
	Kind   ChangeKind       `json:"kind"`
	Before *json.RawMessage `json:"before,omitempty"`
	After  *json.RawMessage `json:"after,omitempty"`
}

func CompareJSONFiles(options JSONOptions) (JSONResult, error) {
	if options.FromPath == "" {
		return JSONResult{}, fmt.Errorf("from JSON path is required")
	}
	if options.ToPath == "" {
		return JSONResult{}, fmt.Errorf("to JSON path is required")
	}
	before, err := readJSONValue(options.FromPath)
	if err != nil {
		return JSONResult{}, err
	}
	after, err := readJSONValue(options.ToPath)
	if err != nil {
		return JSONResult{}, err
	}
	changes, err := compareValues("", before, after)
	if err != nil {
		return JSONResult{}, err
	}
	return JSONResult{
		FromPath: options.FromPath,
		ToPath:   options.ToPath,
		Equal:    len(changes) == 0,
		Changes:  changes,
	}, nil
}

func readJSONValue(path string) (any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read JSON file %s: %w", path, err)
	}
	dec := json.NewDecoder(bytes.NewReader(content))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, fmt.Errorf("parse JSON file %s: %w", path, err)
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err == nil {
		return nil, fmt.Errorf("parse JSON file %s: trailing JSON value", path)
	} else if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("parse JSON file %s: %w", path, err)
	}
	return value, nil
}

func compareValues(path string, before any, after any) ([]JSONChange, error) {
	switch beforeValue := before.(type) {
	case map[string]any:
		afterValue, ok := after.(map[string]any)
		if !ok {
			return changedValue(path, before, after)
		}
		return compareObjects(path, beforeValue, afterValue)
	case []any:
		afterValue, ok := after.([]any)
		if !ok {
			return changedValue(path, before, after)
		}
		return compareArrays(path, beforeValue, afterValue)
	default:
		if reflect.DeepEqual(before, after) {
			return nil, nil
		}
		return changedValue(path, before, after)
	}
}

func compareObjects(path string, before map[string]any, after map[string]any) ([]JSONChange, error) {
	keys := sortedObjectKeys(before, after)
	changes := []JSONChange{}
	for _, key := range keys {
		childPath := joinPath(path, key)
		beforeValue, beforeOK := before[key]
		afterValue, afterOK := after[key]
		switch {
		case !beforeOK:
			change, err := addedValue(childPath, afterValue)
			if err != nil {
				return nil, err
			}
			changes = append(changes, change)
		case !afterOK:
			change, err := removedValue(childPath, beforeValue)
			if err != nil {
				return nil, err
			}
			changes = append(changes, change)
		default:
			childChanges, err := compareValues(childPath, beforeValue, afterValue)
			if err != nil {
				return nil, err
			}
			changes = append(changes, childChanges...)
		}
	}
	return changes, nil
}

func compareArrays(path string, before []any, after []any) ([]JSONChange, error) {
	changes := []JSONChange{}
	maxLength := max(len(before), len(after))
	for index := range maxLength {
		childPath := joinPath(path, strconv.Itoa(index))
		switch {
		case index >= len(before):
			change, err := addedValue(childPath, after[index])
			if err != nil {
				return nil, err
			}
			changes = append(changes, change)
		case index >= len(after):
			change, err := removedValue(childPath, before[index])
			if err != nil {
				return nil, err
			}
			changes = append(changes, change)
		default:
			childChanges, err := compareValues(childPath, before[index], after[index])
			if err != nil {
				return nil, err
			}
			changes = append(changes, childChanges...)
		}
	}
	return changes, nil
}

func addedValue(path string, value any) (JSONChange, error) {
	after, err := rawJSON(value)
	if err != nil {
		return JSONChange{}, err
	}
	return JSONChange{Path: printablePath(path), Kind: ChangeAdded, After: after}, nil
}

func removedValue(path string, value any) (JSONChange, error) {
	before, err := rawJSON(value)
	if err != nil {
		return JSONChange{}, err
	}
	return JSONChange{Path: printablePath(path), Kind: ChangeRemoved, Before: before}, nil
}

func changedValue(path string, before any, after any) ([]JSONChange, error) {
	beforeRaw, err := rawJSON(before)
	if err != nil {
		return nil, err
	}
	afterRaw, err := rawJSON(after)
	if err != nil {
		return nil, err
	}
	return []JSONChange{{Path: printablePath(path), Kind: ChangeChanged, Before: beforeRaw, After: afterRaw}}, nil
}

func rawJSON(value any) (*json.RawMessage, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal JSON diff value: %w", err)
	}
	raw := json.RawMessage(data)
	return &raw, nil
}

func sortedObjectKeys(before map[string]any, after map[string]any) []string {
	seen := map[string]struct{}{}
	for key := range before {
		seen[key] = struct{}{}
	}
	for key := range after {
		seen[key] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func joinPath(parent string, token string) string {
	escapedToken := escapePointerToken(token)
	if parent == "" {
		return "/" + escapedToken
	}
	return parent + "/" + escapedToken
}

func printablePath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

func escapePointerToken(token string) string {
	token = strings.ReplaceAll(token, "~", "~0")
	return strings.ReplaceAll(token, "/", "~1")
}
