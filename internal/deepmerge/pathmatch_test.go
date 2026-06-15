// Copyright (c) Labyrinth Labs s.r.o.
// SPDX-License-Identifier: MPL-2.0

package deepmerge

import (
	"reflect"
	"testing"
)

func TestParseJqPath(t *testing.T) {
	cases := []struct {
		in      string
		want    []string
		wantErr bool
	}{
		{in: ".a", want: []string{"a"}},
		{in: ".a.b", want: []string{"a", "b"}},
		{in: ".a.b.c", want: []string{"a", "b", "c"}},
		{in: ".spec.containers", want: []string{"spec", "containers"}},
		{in: ".a.*.c", want: []string{"a", "*", "c"}},
		{in: ".a[]", want: []string{"a", "[]"}},
		{in: ".a[].b", want: []string{"a", "[]", "b"}},
		{in: ".a[][]", want: []string{"a", "[]", "[]"}},
		{in: ".spec.containers[].env", want: []string{"spec", "containers", "[]", "env"}},

		{in: "", wantErr: true},
		{in: "a", wantErr: true},      // missing leading dot
		{in: ".", wantErr: true},      // empty segment
		{in: ".a..b", wantErr: true},  // empty segment
		{in: ".a[0]", wantErr: true},  // indexed access unsupported
		{in: ".a.[b]", wantErr: true}, // bracket key unsupported
		{in: ".a[", wantErr: true},    // unterminated bracket
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := parseJqPath(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseJqPath(%q) err=%v wantErr=%v", tc.in, err, tc.wantErr)
			}
			if !tc.wantErr && !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("parseJqPath(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestPathRuleMatches(t *testing.T) {
	cases := []struct {
		rule string
		path []string
		want bool
	}{
		{rule: ".a", path: []string{"a"}, want: true},
		{rule: ".a", path: []string{"b"}, want: false},
		{rule: ".a", path: []string{"a", "b"}, want: false}, // length mismatch (only-at-matched-node)
		{rule: ".a.b", path: []string{"a", "b"}, want: true},
		{rule: ".a.b", path: []string{"a", "c"}, want: false},
		{rule: ".a.*", path: []string{"a", "anything"}, want: true},
		{rule: ".a.*.c", path: []string{"a", "x", "c"}, want: true},
		{rule: ".a.*.c", path: []string{"a", "x", "y"}, want: false},
		{rule: ".a[]", path: []string{"a", "anything"}, want: true},
		{rule: ".a[].b", path: []string{"a", "x", "b"}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.rule, func(t *testing.T) {
			steps, err := parseJqPath(tc.rule)
			if err != nil {
				t.Fatalf("parseJqPath(%q) unexpected error: %v", tc.rule, err)
			}
			r := pathRule{steps: steps}
			if got := r.matches(tc.path); got != tc.want {
				t.Fatalf("rule %q matches %v = %v, want %v", tc.rule, tc.path, got, tc.want)
			}
		})
	}
}

func TestCompilePathRulesSorted(t *testing.T) {
	True := true
	overrides := map[string]*PathOverrideOptions{
		".b": {AppendList: &True},
		".a": {AppendList: &True},
		".c": {AppendList: &True},
	}
	rules, err := compilePathRules(overrides)
	if err != nil {
		t.Fatalf("compilePathRules err: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}
	got := []string{rules[0].raw, rules[1].raw, rules[2].raw}
	want := []string{".a", ".b", ".c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("compilePathRules order = %v, want %v", got, want)
	}
}

func TestCompilePathRulesError(t *testing.T) {
	True := true
	overrides := map[string]*PathOverrideOptions{
		"missing_leading_dot": {AppendList: &True},
	}
	if _, err := compilePathRules(overrides); err == nil {
		t.Fatalf("expected error from malformed path")
	}
}
