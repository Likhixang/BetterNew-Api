/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
package model

import (
	"regexp"
	"testing"
)

func TestMergeModelNameExact(t *testing.T) {
	modelMergeMutex.Lock()
	modelMergeExact = map[string]string{"deepseek_ai/deepseek-v4-flash": "deepseek-v4-flash"}
	modelMergeRegex = nil
	modelMergeReady = true
	modelMergeMutex.Unlock()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"exact alias maps to canonical", "deepseek_ai/deepseek-v4-flash", "deepseek-v4-flash"},
		{"canonical name passes through", "deepseek-v4-flash", "deepseek-v4-flash"},
		{"unrelated model unchanged", "claude-opus-4-8", "claude-opus-4-8"},
		{"empty input unchanged", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeModelName(tt.input); got != tt.expected {
				t.Fatalf("MergeModelName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMergeModelNameRegex(t *testing.T) {
	// Simulate what InitModelMergeCache builds for the AxonHub-style rule:
	// (?i)(?:deepseek(?:-ai)?/)?deepseek-v4-flash(?:-[a-z0-9]+)?
	pattern, err := regexp.Compile(`(?i)(?:deepseek(?:-ai)?/)?deepseek-v4-flash(?:-[a-z0-9]+)?`)
	if err != nil {
		t.Fatalf("compile pattern: %v", err)
	}
	modelMergeMutex.Lock()
	modelMergeExact = map[string]string{}
	modelMergeRegex = []modelMergeRegexRule{
		{priority: 1, pattern: pattern, target: "deepseek-v4-flash"},
	}
	modelMergeReady = true
	modelMergeMutex.Unlock()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"bare name", "deepseek-v4-flash", "deepseek-v4-flash"},
		{"deepseek prefix", "deepseek/deepseek-v4-flash", "deepseek-v4-flash"},
		{"deepseek-ai prefix", "deepseek-ai/deepseek-v4-flash", "deepseek-v4-flash"},
		{"variant suffix", "deepseek-ai/deepseek-v4-flash-lite", "deepseek-v4-flash"},
		{"case insensitive", "DeepSeek-ai/DeepSeek-v4-Flash", "deepseek-v4-flash"},
		{"no match unchanged", "deepseek-v4-pro", "deepseek-v4-pro"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeModelName(tt.input); got != tt.expected {
				t.Fatalf("MergeModelName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMergeModelNameExactWinsOverRegex(t *testing.T) {
	regexPattern, err := regexp.Compile(`deepseek.*`)
	if err != nil {
		t.Fatalf("compile pattern: %v", err)
	}
	modelMergeMutex.Lock()
	// Exact alias for a name that would also match the regex
	modelMergeExact = map[string]string{"deepseek-v4-flash": "deepseek-v4-flash"}
	modelMergeRegex = []modelMergeRegexRule{
		{priority: 1, pattern: regexPattern, target: "deepseek-canonical"},
	}
	modelMergeReady = true
	modelMergeMutex.Unlock()

	// "deepseek-v4-flash" hits the exact map first
	if got := MergeModelName("deepseek-v4-flash"); got != "deepseek-v4-flash" {
		t.Fatalf("exact should win, got %q", got)
	}
	// "deepseek-anything-else" falls through to regex
	if got := MergeModelName("deepseek-x1"); got != "deepseek-canonical" {
		t.Fatalf("regex fallback should hit, got %q", got)
	}
}

func TestMergeModelNameSingleHop(t *testing.T) {
	// A->B, B->C rules must NOT chain: A goes to B, not C.
	modelMergeMutex.Lock()
	modelMergeExact = map[string]string{
		"model-a": "model-b",
		"model-b": "model-c",
	}
	modelMergeRegex = nil
	modelMergeReady = true
	modelMergeMutex.Unlock()

	if got := MergeModelName("model-a"); got != "model-b" {
		t.Fatalf("single-hop expected model-b, got %q", got)
	}
}

func TestMergeModelNameNotReady(t *testing.T) {
	modelMergeMutex.Lock()
	modelMergeReady = false
	modelMergeMutex.Unlock()

	if got := MergeModelName("deepseek_ai/deepseek-v4-flash"); got != "deepseek_ai/deepseek-v4-flash" {
		t.Fatalf("cache not ready should pass through, got %q", got)
	}

	modelMergeMutex.Lock()
	modelMergeReady = true
	modelMergeMutex.Unlock()
}

func TestPreviewModelMergeMatchesExact(t *testing.T) {
	channels := []*Channel{
		{Id: 1, Name: "ds-direct", Type: 1, Status: 1, Models: "deepseek-v4-flash,claude-opus-4-8"},
		{Id: 2, Name: "ds-cx", Type: 1, Status: 1, Models: "cx/deepseek-v4-flash,gpt-5.5"},
		{Id: 3, Name: "misc", Type: 1, Status: 1, Models: "gemini-2.5-pro"},
	}

	result, err := PreviewModelMergeMatches(channels, "deepseek-v4-flash", ModelMergeMatchExact)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(result))
	}
	if result[0].Id != 1 || result[0].Name != "ds-direct" {
		t.Fatalf("unexpected channel: %+v", result[0])
	}
	if len(result[0].Models) != 1 || result[0].Models[0] != "deepseek-v4-flash" {
		t.Fatalf("unexpected matched models: %v", result[0].Models)
	}
}

func TestPreviewModelMergeMatchesRegex(t *testing.T) {
	channels := []*Channel{
		{Id: 1, Name: "ds-direct", Type: 1, Status: 1, Models: "deepseek-v4-flash,deepseek-v4-pro,claude-opus-4-8"},
		{Id: 2, Name: "ds-cx", Type: 1, Status: 1, Models: "cx/deepseek-v4-flash,gpt-5.5"},
		{Id: 3, Name: "misc", Type: 1, Status: 1, Models: "gemini-2.5-pro"},
	}

	result, err := PreviewModelMergeMatches(channels, `(?i)(?:deepseek(?:-ai)?/)?deepseek-v4-flash`, ModelMergeMatchRegex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(result))
	}
	// Both channels matched, each with exactly one model
	for _, ch := range result {
		if len(ch.Models) != 1 {
			t.Fatalf("channel %s expected 1 matched model, got %v", ch.Name, ch.Models)
		}
	}
}

func TestPreviewModelMergeMatchesNoHit(t *testing.T) {
	channels := []*Channel{
		{Id: 1, Name: "misc", Type: 1, Status: 1, Models: "gemini-2.5-pro,gpt-5.5"},
	}

	result, err := PreviewModelMergeMatches(channels, "deepseek-v4-flash", ModelMergeMatchExact)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected no channels, got %d", len(result))
	}
}

func TestPreviewModelMergeMatchesSkipsDisabledChannelsModels(t *testing.T) {
	// A channel with no models or nil must not break the preview; models with
	// blank entries are skipped rather than compared.
	channels := []*Channel{
		{Id: 1, Name: "empty", Type: 1, Status: 1, Models: ""},
		nil,
		{Id: 3, Name: "trailing", Type: 1, Status: 1, Models: "deepseek-v4-flash,"},
	}

	result, err := PreviewModelMergeMatches(channels, "deepseek-v4-flash", ModelMergeMatchExact)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].Id != 3 {
		t.Fatalf("expected only channel 3 to match, got %+v", result)
	}
}

func TestPreviewModelMergeMatchesInvalidRegex(t *testing.T) {
	if _, err := PreviewModelMergeMatches(nil, "([unclosed", ModelMergeMatchRegex); err == nil {
		t.Fatal("expected invalid regex error")
	}
}

func TestPreviewModelMergeMatchesEmptyAlias(t *testing.T) {
	if _, err := PreviewModelMergeMatches(nil, "  ", ModelMergeMatchExact); err == nil {
		t.Fatal("expected empty alias error")
	}
}
