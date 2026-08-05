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
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/QuantumNous/new-api/common"
)

// In-memory cache of model merge rules (see ModelMerge).
// Exact aliases win over regex rules; regex rules apply in ascending rule id
// order (first match wins).

var (
	modelMergeMutex sync.RWMutex
	modelMergeExact map[string]string // alias -> target model
	modelMergeRegex []modelMergeRegexRule
	modelMergeReady bool
)

type modelMergeRegexRule struct {
	priority int
	pattern  *regexp.Regexp
	target   string
}

func init() {
	modelMergeExact = make(map[string]string)
	modelMergeRegex = make([]modelMergeRegexRule, 0)
}

// InitModelMergeCache loads all enabled model merge rules into memory.
// Called at startup and after every rule change.
func InitModelMergeCache() error {
	merges, err := GetAllModelMerges()
	if err != nil {
		common.SysLog("model merge: failed to load rules: " + err.Error())
		return err
	}

	newExact := make(map[string]string)
	newRegex := make([]modelMergeRegexRule, 0, len(merges))
	for _, m := range merges {
		if m.Status != ModelMergeStatusEnabled {
			continue
		}
		alias := strings.TrimSpace(m.Alias)
		target := strings.TrimSpace(m.TargetModel)
		if alias == "" || target == "" {
			continue
		}
		if m.MatchType == ModelMergeMatchRegex {
			pattern, err := regexp.Compile(alias)
			if err != nil {
				common.SysLog("model merge: invalid regex alias " + alias + ": " + err.Error())
				continue
			}
			newRegex = append(newRegex, modelMergeRegexRule{
				priority: m.Id,
				pattern:  pattern,
				target:   target,
			})
		} else {
			newExact[alias] = target
		}
	}

	sort.Slice(newRegex, func(i, j int) bool {
		return newRegex[i].priority < newRegex[j].priority
	})

	modelMergeMutex.Lock()
	modelMergeExact = newExact
	modelMergeRegex = newRegex
	modelMergeReady = true
	modelMergeMutex.Unlock()

	common.SysLog("model merge: loaded exact=" + strconv.Itoa(len(newExact)) + " regex=" + strconv.Itoa(len(newRegex)))
	return nil
}

// MergeModelName normalizes a requested model name to its canonical target
// model. Exact alias match is checked first, then regex rules in ascending id
// order. Returns the original name when no rule matches. Single-hop only —
// the result is not re-merged, so cyclic rules cannot loop.
func MergeModelName(modelName string) string {
	if modelName == "" {
		return modelName
	}
	modelMergeMutex.RLock()
	defer modelMergeMutex.RUnlock()
	if !modelMergeReady {
		return modelName
	}
	if target, ok := modelMergeExact[modelName]; ok {
		return target
	}
	for _, rule := range modelMergeRegex {
		if rule.pattern.MatchString(modelName) {
			return rule.target
		}
	}
	return modelName
}
