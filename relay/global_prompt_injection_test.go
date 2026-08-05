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
*/
package relay

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/setting/model_setting"
)

func setGlobalPrompt(enabled bool, action string, content string) {
	cfg := model_setting.GetGlobalSettings()
	cfg.GlobalPromptInjection = model_setting.GlobalPromptInjection{
		Enabled: enabled,
		Action:  action,
		Content: content,
	}
}

func TestInjectGlobalPromptOpenAIDisabled(t *testing.T) {
	setGlobalPrompt(false, "prepend", "GLOBAL")
	defer setGlobalPrompt(false, "prepend", "")

	req := &dto.GeneralOpenAIRequest{
		Messages: []dto.Message{{Role: "user", Content: "hi"}},
	}
	InjectGlobalPromptOpenAI(req)
	if len(req.Messages) != 1 || req.Messages[0].Content != "hi" {
		t.Fatalf("disabled injection must not touch messages: %+v", req.Messages)
	}
}

func TestInjectGlobalPromptOpenAIPrepend(t *testing.T) {
	setGlobalPrompt(true, "prepend", "GLOBAL")
	defer setGlobalPrompt(false, "prepend", "")

	req := &dto.GeneralOpenAIRequest{
		Messages: []dto.Message{
			{Role: "system", Content: "original"},
			{Role: "user", Content: "hi"},
		},
	}
	InjectGlobalPromptOpenAI(req)
	if req.Messages[0].Role != "system" || req.Messages[0].Content != "GLOBAL\noriginal" {
		t.Fatalf("prepend failed: %+v", req.Messages[0])
	}
}

func TestInjectGlobalPromptOpenAIAppend(t *testing.T) {
	setGlobalPrompt(true, "append", "GLOBAL")
	defer setGlobalPrompt(false, "prepend", "")

	req := &dto.GeneralOpenAIRequest{
		Messages: []dto.Message{
			{Role: "system", Content: "original"},
			{Role: "user", Content: "hi"},
		},
	}
	InjectGlobalPromptOpenAI(req)
	if req.Messages[0].Content != "original\nGLOBAL" {
		t.Fatalf("append failed: %+v", req.Messages[0])
	}
}

func TestInjectGlobalPromptOpenAIOverride(t *testing.T) {
	setGlobalPrompt(true, "override", "GLOBAL")
	defer setGlobalPrompt(false, "prepend", "")

	req := &dto.GeneralOpenAIRequest{
		Messages: []dto.Message{
			{Role: "system", Content: "original"},
			{Role: "user", Content: "hi"},
		},
	}
	InjectGlobalPromptOpenAI(req)
	if req.Messages[0].Content != "GLOBAL" {
		t.Fatalf("override failed: %+v", req.Messages[0])
	}
	if req.Messages[1].Content != "hi" {
		t.Fatalf("user message must be preserved: %+v", req.Messages[1])
	}
}

func TestInjectGlobalPromptOpenAINoSystem(t *testing.T) {
	setGlobalPrompt(true, "prepend", "GLOBAL")
	defer setGlobalPrompt(false, "prepend", "")

	req := &dto.GeneralOpenAIRequest{
		Messages: []dto.Message{{Role: "user", Content: "hi"}},
	}
	InjectGlobalPromptOpenAI(req)
	if len(req.Messages) != 2 || req.Messages[0].Role != "system" || req.Messages[0].Content != "GLOBAL" {
		t.Fatalf("system insert failed: %+v", req.Messages)
	}
}

func TestInjectGlobalPromptOpenAIMediaContent(t *testing.T) {
	setGlobalPrompt(true, "prepend", "GLOBAL")
	defer setGlobalPrompt(false, "prepend", "")

	req := &dto.GeneralOpenAIRequest{
		Messages: []dto.Message{
			{
				Role: "system",
				Content: []any{
					dto.MediaContent{Type: dto.ContentTypeText, Text: "original"},
				},
			},
			{Role: "user", Content: "hi"},
		},
	}
	InjectGlobalPromptOpenAI(req)
	contents := req.Messages[0].ParseContent()
	if len(contents) != 2 || contents[0].Text != "GLOBAL" || contents[1].Text != "original" {
		t.Fatalf("media prepend failed: %+v", contents)
	}
}

func TestInjectGlobalPromptClaude(t *testing.T) {
	setGlobalPrompt(true, "prepend", "GLOBAL")
	defer setGlobalPrompt(false, "prepend", "")

	req := &dto.ClaudeRequest{
		System:   "original",
		Messages: []dto.ClaudeMessage{{Role: "user", Content: "hi"}},
	}
	InjectGlobalPromptClaude(req)
	if !req.IsStringSystem() || req.GetStringSystem() != "GLOBAL\noriginal" {
		t.Fatalf("claude prepend failed: %q", req.GetStringSystem())
	}
}

func TestInjectGlobalPromptClaudeNoSystem(t *testing.T) {
	setGlobalPrompt(true, "prepend", "GLOBAL")
	defer setGlobalPrompt(false, "prepend", "")

	req := &dto.ClaudeRequest{
		Messages: []dto.ClaudeMessage{{Role: "user", Content: "hi"}},
	}
	InjectGlobalPromptClaude(req)
	if !req.IsStringSystem() || req.GetStringSystem() != "GLOBAL" {
		t.Fatalf("claude system insert failed: %q", req.GetStringSystem())
	}
}

func TestInjectGlobalPromptGemini(t *testing.T) {
	setGlobalPrompt(true, "prepend", "GLOBAL")
	defer setGlobalPrompt(false, "prepend", "")

	req := &dto.GeminiChatRequest{
		SystemInstructions: &dto.GeminiChatContent{
			Parts: []dto.GeminiPart{{Text: "original"}},
		},
		Contents: []dto.GeminiChatContent{{Role: "user", Parts: []dto.GeminiPart{{Text: "hi"}}}},
	}
	InjectGlobalPromptGemini(req)
	if req.SystemInstructions == nil || len(req.SystemInstructions.Parts) != 1 ||
		req.SystemInstructions.Parts[0].Text != "GLOBAL\noriginal" {
		t.Fatalf("gemini prepend failed: %+v", req.SystemInstructions)
	}
}

func TestInjectGlobalPromptGeminiNoSystem(t *testing.T) {
	setGlobalPrompt(true, "append", "GLOBAL")
	defer setGlobalPrompt(false, "prepend", "")

	req := &dto.GeminiChatRequest{
		Contents: []dto.GeminiChatContent{{Role: "user", Parts: []dto.GeminiPart{{Text: "hi"}}}},
	}
	InjectGlobalPromptGemini(req)
	if req.SystemInstructions == nil || len(req.SystemInstructions.Parts) != 1 ||
		req.SystemInstructions.Parts[0].Text != "GLOBAL" {
		t.Fatalf("gemini system insert failed: %+v", req.SystemInstructions)
	}
}
