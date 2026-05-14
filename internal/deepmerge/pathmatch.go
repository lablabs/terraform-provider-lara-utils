// Copyright (c) Labyrinth Labs s.r.o.
// SPDX-License-Identifier: MPL-2.0

package deepmerge

import (
	"fmt"
	"slices"
	"strings"
)

// pathRule pairs a parsed jq-style path with the options to apply when the
// path matches the current merge location.
type pathRule struct {
	raw   string
	steps []string
	opts  *PathOverrideOptions
}

// parseJqPath turns a jq-style path string into a list of match steps.
//
// Supported syntax:
//
//	".a.b.c"      -> ["a","b","c"]
//	".a.*.c"      -> ["a","*","c"]   (* matches any single key)
//	".a[].c"      -> ["a","[]","c"]  ([] matches any single key — v1 treats it as *)
//	".a[][]"     -> ["a","[]","[]"]
//
// A leading "." is required. Empty segments (".."), bracket indices (".a[0]"),
// pipes, filters, and other jq operators are not supported.
func parseJqPath(s string) ([]string, error) {
	if s == "" {
		return nil, fmt.Errorf("path is empty")
	}
	if !strings.HasPrefix(s, ".") {
		return nil, fmt.Errorf("path %q must start with '.'", s)
	}
	if s == "." {
		return nil, fmt.Errorf("path %q has an empty segment", s)
	}

	steps := []string{}
	// Walk character by character to keep "[]" attached to the preceding key
	// rather than being split by '.'.
	rest := s[1:] // drop leading dot
	for len(rest) > 0 {
		// "[]" as its own step (e.g. ".a[]" -> "a" then "[]").
		if strings.HasPrefix(rest, "[]") {
			steps = append(steps, "[]")
			rest = rest[2:]
			// allow ".[]" sequences and "[][]"; consume an optional separating dot
			if strings.HasPrefix(rest, ".") {
				rest = rest[1:]
				if rest == "" || strings.HasPrefix(rest, ".") {
					return nil, fmt.Errorf("path %q has an empty segment", s)
				}
			}
			continue
		}

		// Read a key until the next '.' or '['.
		end := strings.IndexAny(rest, ".[")
		var key string
		if end < 0 {
			key = rest
			rest = ""
		} else {
			key = rest[:end]
			rest = rest[end:]
		}

		if key == "" {
			return nil, fmt.Errorf("path %q has an empty segment", s)
		}

		// Reject anything that isn't a plain key or "*". Indexed access like
		// "[0]" is unsupported in v1.
		if strings.ContainsAny(key, "[]") {
			return nil, fmt.Errorf("path %q contains unsupported bracket syntax", s)
		}
		steps = append(steps, key)

		// Consume a separator if present.
		switch {
		case rest == "":
			// done
		case strings.HasPrefix(rest, "."):
			rest = rest[1:]
			if rest == "" || strings.HasPrefix(rest, ".") {
				return nil, fmt.Errorf("path %q has an empty segment", s)
			}
		case strings.HasPrefix(rest, "[]"):
			// keep going; next iteration consumes the "[]"
		case strings.HasPrefix(rest, "["):
			return nil, fmt.Errorf("path %q contains unsupported bracket syntax", s)
		}
	}

	return steps, nil
}

// matches reports whether the rule's steps match the given concrete path.
// "*" and "[]" each match exactly one path segment regardless of its value.
func (r pathRule) matches(path []string) bool {
	if len(r.steps) != len(path) {
		return false
	}
	for i, step := range r.steps {
		switch step {
		case "*", "[]":
			// wildcard
		default:
			if step != path[i] {
				return false
			}
		}
	}
	return true
}

// compilePathRules parses each entry of the path_overrides map into a pathRule.
// The slice is returned in sorted-by-key order so that rule precedence is
// deterministic (a later rule in the sorted order wins when multiple match).
func compilePathRules(overrides map[string]*PathOverrideOptions) ([]pathRule, error) {
	if len(overrides) == 0 {
		return nil, nil
	}
	keys := make([]string, 0, len(overrides))
	for k := range overrides {
		keys = append(keys, k)
	}
	// stable order so that "last matching rule wins" is deterministic
	slices.Sort(keys)

	rules := make([]pathRule, 0, len(keys))
	for _, k := range keys {
		steps, err := parseJqPath(k)
		if err != nil {
			return nil, err
		}
		rules = append(rules, pathRule{raw: k, steps: steps, opts: overrides[k]})
	}
	return rules, nil
}
