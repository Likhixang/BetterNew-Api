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
	"errors"
	"regexp"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

const (
	ModelMergeMatchExact = 0
	ModelMergeMatchRegex = 1

	ModelMergeStatusDisabled = 0
	ModelMergeStatusEnabled   = 1
)

// ModelMerge defines an inbound model-name merge rule: a request model name
// (alias, either exact or regex) is normalized to the canonical target model
// BEFORE token whitelist validation and channel selection. This mirrors
// AxonHub model associations (e.g. deepseek_ai/deepseek-v4-flash -> deepseek-v4-flash).
type ModelMerge struct {
	Id          int    `json:"id"`
	TargetModel string `json:"target_model" gorm:"size:128;not null;index"`
	Alias       string `json:"alias" gorm:"type:text;not null"`
	MatchType   int    `json:"match_type" gorm:"default:0"`
	Status      int    `json:"status" gorm:"default:1"`
	CreatedTime int64  `json:"created_time" gorm:"bigint"`
	UpdatedTime int64  `json:"updated_time" gorm:"bigint"`
}

func (m *ModelMerge) Insert() error {
	now := common.GetTimestamp()
	m.CreatedTime = now
	m.UpdatedTime = now
	err := DB.Create(m).Error
	if err == nil {
		_ = InitModelMergeCache()
	}
	return err
}

func (m *ModelMerge) Update() error {
	now := common.GetTimestamp()
	m.UpdatedTime = now
	err := DB.Model(m).Updates(map[string]any{
		"target_model": m.TargetModel,
		"alias":        m.Alias,
		"match_type":   m.MatchType,
		"status":       m.Status,
		"updated_time": now,
	}).Error
	if err == nil {
		_ = InitModelMergeCache()
	}
	return err
}

func (m *ModelMerge) Delete() error {
	if m.Id == 0 {
		return errors.New("model merge id is empty")
	}
	err := DB.Delete(m).Error
	if err == nil {
		_ = InitModelMergeCache()
	}
	return err
}

func GetAllModelMerges() ([]*ModelMerge, error) {
	var merges []*ModelMerge
	err := DB.Order("id asc").Find(&merges).Error
	return merges, err
}

func GetModelMergeById(id int) (*ModelMerge, error) {
	if id == 0 {
		return nil, errors.New("model merge id is empty")
	}
	merge := &ModelMerge{}
	err := DB.Where("id = ?", id).First(merge).Error
	return merge, err
}

// ModelMergePreviewChannel is a channel plus the models that match a draft rule.
type ModelMergePreviewChannel struct {
	Id     int      `json:"id"`
	Name   string   `json:"name"`
	Type   int      `json:"type"`
	Status int      `json:"status"`
	Models []string `json:"models"`
}

// PreviewModelMergeMatches evaluates a draft rule (one or more aliases, one per
// line, with a match type) against every channel's configured models and
// returns only the channels that have at least one matching model. Exact rules
// compare trimmed names; regex rules use Go regexp.MatchString. All aliases
// share the same match type (per-rule). The returned channels never carry API
// keys.
func PreviewModelMergeMatches(channels []*Channel, alias string, matchType int) ([]*ModelMergePreviewChannel, error) {
	aliases := splitAliases(alias)
	if len(aliases) == 0 {
		return nil, errors.New("alias is required")
	}

	// Precompile regex rules; exact aliases stay as plain strings.
	type exactRule struct{ name string }
	exactRules := make([]exactRule, 0, len(aliases))
	regexRules := make([]*regexp.Regexp, 0, len(aliases))
	for _, a := range aliases {
		if matchType == ModelMergeMatchRegex {
			pattern, err := regexp.Compile(a)
			if err != nil {
				return nil, errors.New("invalid regex alias: " + err.Error())
			}
			regexRules = append(regexRules, pattern)
		} else {
			exactRules = append(exactRules, exactRule{name: a})
		}
	}

	result := make([]*ModelMergePreviewChannel, 0)
	for _, ch := range channels {
		if ch == nil {
			continue
		}
		matched := make([]string, 0)
		for _, m := range ch.GetModels() {
			m = strings.TrimSpace(m)
			if m == "" {
				continue
			}
			hit := false
			if len(regexRules) > 0 {
				for _, pattern := range regexRules {
					if pattern.MatchString(m) {
						hit = true
						break
					}
				}
			}
			if !hit {
				for _, rule := range exactRules {
					if m == rule.name {
						hit = true
						break
					}
				}
			}
			if hit {
				matched = append(matched, m)
			}
		}
		if len(matched) > 0 {
			result = append(result, &ModelMergePreviewChannel{
				Id:     ch.Id,
				Name:   ch.Name,
				Type:   ch.Type,
				Status: ch.Status,
				Models: matched,
			})
		}
	}
	return result, nil
}

// splitAliases splits a multi-line alias field into individual non-empty
// trimmed aliases. One alias per line.
func splitAliases(alias string) []string {
	lines := strings.Split(alias, "\n")
	aliases := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		aliases = append(aliases, line)
	}
	return aliases
}
