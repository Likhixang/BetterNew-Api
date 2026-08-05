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
	"strings"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/setting/model_setting"
)

// Global prompt injection (system setting "Global Model Configuration").
// Mirrors AxonHub's global prompt prepend/append: a single system prompt can
// be prepended, appended, or override the existing system content on every
// relayed chat request, regardless of the upstream protocol.

const (
	globalPromptActionPrepend = "prepend"
	globalPromptActionAppend  = "append"
	globalPromptActionOverride = "override"
)

// globalPromptInjection returns the enabled injection content and action.
func globalPromptInjection() (enabled bool, action string, content string) {
	cfg := model_setting.GetGlobalSettings().GlobalPromptInjection
	if !cfg.Enabled {
		return false, "", ""
	}
	content = strings.TrimSpace(cfg.Content)
	if content == "" {
		return false, "", ""
	}
	action = cfg.Action
	if action != globalPromptActionPrepend && action != globalPromptActionAppend && action != globalPromptActionOverride {
		action = globalPromptActionPrepend
	}
	return true, action, content
}

// InjectGlobalPromptOpenAI applies the global prompt to an OpenAI-format
// request by rewriting its system messages.
func InjectGlobalPromptOpenAI(request *dto.GeneralOpenAIRequest) {
	enabled, action, content := globalPromptInjection()
	if !enabled || request == nil {
		return
	}
	systemRole := request.GetSystemRoleName()
	// Locate the first system message.
	index := -1
	for i := range request.Messages {
		if request.Messages[i].Role == systemRole {
			index = i
			break
		}
	}

	systemMessage := dto.Message{Role: systemRole, Content: content}
	if index < 0 {
		// No system message: insert one at the front (prepend/override) or
		// keep append semantics identical — a leading system message is the
		// only sensible position for all actions.
		request.Messages = append([]dto.Message{systemMessage}, request.Messages...)
		return
	}

	switch action {
	case globalPromptActionOverride:
		request.Messages[index].Content = content
	case globalPromptActionAppend:
		msg := &request.Messages[index]
		if msg.IsStringContent() {
			msg.SetStringContent(msg.StringContent() + "\n" + content)
		} else {
			contents := msg.ParseContent()
			contents = append(contents, dto.MediaContent{Type: dto.ContentTypeText, Text: content})
			msg.SetMediaContent(contents)
		}
	default: // prepend
		msg := &request.Messages[index]
		if msg.IsStringContent() {
			msg.SetStringContent(content + "\n" + msg.StringContent())
		} else {
			contents := msg.ParseContent()
			contents = append([]dto.MediaContent{{Type: dto.ContentTypeText, Text: content}}, contents...)
			msg.SetMediaContent(contents)
		}
	}
}

// InjectGlobalPromptClaude applies the global prompt to a Claude-format
// request by rewriting its system block.
func InjectGlobalPromptClaude(request *dto.ClaudeRequest) {
	enabled, action, content := globalPromptInjection()
	if !enabled || request == nil {
		return
	}

	if request.System == nil {
		request.SetStringSystem(content)
		return
	}

	switch action {
	case globalPromptActionOverride:
		request.SetStringSystem(content)
	case globalPromptActionAppend:
		if request.IsStringSystem() {
			existing := strings.TrimSpace(request.GetStringSystem())
			if existing == "" {
				request.SetStringSystem(content)
			} else {
				request.SetStringSystem(existing + "\n" + content)
			}
		} else {
			systemContents := request.ParseSystem()
			extra := dto.ClaudeMediaMessage{Type: dto.ContentTypeText}
			extra.SetText(content)
			request.System = append(systemContents, extra)
		}
	default: // prepend
		if request.IsStringSystem() {
			existing := strings.TrimSpace(request.GetStringSystem())
			if existing == "" {
				request.SetStringSystem(content)
			} else {
				request.SetStringSystem(content + "\n" + existing)
			}
		} else {
			systemContents := request.ParseSystem()
			extra := dto.ClaudeMediaMessage{Type: dto.ContentTypeText}
			extra.SetText(content)
			request.System = append([]dto.ClaudeMediaMessage{extra}, systemContents...)
		}
	}
}

// InjectGlobalPromptGemini applies the global prompt to a Gemini-format
// request by rewriting its systemInstruction block.
func InjectGlobalPromptGemini(request *dto.GeminiChatRequest) {
	enabled, action, content := globalPromptInjection()
	if !enabled || request == nil {
		return
	}

	if request.SystemInstructions == nil || len(request.SystemInstructions.Parts) == 0 {
		request.SystemInstructions = &dto.GeminiChatContent{
			Parts: []dto.GeminiPart{{Text: content}},
		}
		return
	}

	switch action {
	case globalPromptActionOverride:
		request.SystemInstructions.Parts = []dto.GeminiPart{{Text: content}}
	case globalPromptActionAppend:
		// Append to the first non-empty text part, else add a new part.
		merged := false
		for i := range request.SystemInstructions.Parts {
			if request.SystemInstructions.Parts[i].Text == "" {
				continue
			}
			request.SystemInstructions.Parts[i].Text = request.SystemInstructions.Parts[i].Text + "\n" + content
			merged = true
			break
		}
		if !merged {
			request.SystemInstructions.Parts = append(request.SystemInstructions.Parts, dto.GeminiPart{Text: content})
		}
	default: // prepend
		merged := false
		for i := range request.SystemInstructions.Parts {
			if request.SystemInstructions.Parts[i].Text == "" {
				continue
			}
			request.SystemInstructions.Parts[i].Text = content + "\n" + request.SystemInstructions.Parts[i].Text
			merged = true
			break
		}
		if !merged {
			request.SystemInstructions.Parts = append([]dto.GeminiPart{{Text: content}}, request.SystemInstructions.Parts...)
		}
	}
}
