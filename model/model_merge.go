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

// PreviewModelMergeMatches evaluates a draft rule (alias + match type) against
// every channel's configured models and returns only the channels that have at
// least one matching model. Exact rules compare trimmed names; regex rules use
// Go regexp.MatchString. The returned channels never carry API keys.
func PreviewModelMergeMatches(channels []*Channel, alias string, matchType int) ([]*ModelMergePreviewChannel, error) {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return nil, errors.New("alias is required")
	}

	var pattern *regexp.Regexp
	if matchType == ModelMergeMatchRegex {
		var err error
		pattern, err = regexp.Compile(alias)
		if err != nil {
			return nil, errors.New("invalid regex alias: " + err.Error())
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
			if pattern != nil {
				hit = pattern.MatchString(m)
			} else {
				hit = m == alias
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
