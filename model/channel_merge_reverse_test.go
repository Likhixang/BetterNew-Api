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

	"github.com/QuantumNous/new-api/common"
)

// setupMergeReverseChannelCache installs a small in-memory channel cache with
// one channel serving longcat-2.0-free, plus the model-merge rule that maps
// LongCat-2.0(-preview|-free)? to longcat-2.0. All global state is restored on
// cleanup so other tests are not affected.
func setupMergeReverseChannelCache(t *testing.T) {
	t.Helper()
	oldMemoryCache := common.MemoryCacheEnabled
	common.MemoryCacheEnabled = true
	t.Cleanup(func() { common.MemoryCacheEnabled = oldMemoryCache })

	channel := &Channel{
		Id:     18,
		Name:   "OpencodeFree",
		Status: common.ChannelStatusEnabled,
		Models: "longcat-2.0-free",
		Type:   1,
	}

	channelSyncLock.Lock()
	oldChannelsIDM := channelsIDM
	oldGroup2model2channels := group2model2channels
	channelsIDM = map[int]*Channel{18: channel}
	group2model2channels = map[string]map[string][]int{
		"default": {
			"longcat-2.0-free": {18},
		},
	}
	channelSyncLock.Unlock()
	t.Cleanup(func() {
		channelSyncLock.Lock()
		channelsIDM = oldChannelsIDM
		group2model2channels = oldGroup2model2channels
		channelSyncLock.Unlock()
	})

	modelMergeMutex.Lock()
	oldExact := modelMergeExact
	oldRegex := modelMergeRegex
	oldReady := modelMergeReady
	modelMergeExact = map[string]string{}
	modelMergeRegex = []modelMergeRegexRule{
		{
			priority: 35,
			pattern:  regexp.MustCompile(`(?i)^(?:LongCat\-2\.0(?:-(?:preview|free))?)$`),
			target:   "longcat-2.0",
		},
	}
	modelMergeReady = true
	modelMergeMutex.Unlock()
	t.Cleanup(func() {
		modelMergeMutex.Lock()
		modelMergeExact = oldExact
		modelMergeRegex = oldRegex
		modelMergeReady = oldReady
		modelMergeMutex.Unlock()
		mergeReverseMutex.Lock()
		mergeReverseCache = nil
		mergeReverseMutex.Unlock()
	})
}

func TestGetRandomSatisfiedChannelMergeReverse(t *testing.T) {
	setupMergeReverseChannelCache(t)

	channel, channelSideModel, err := GetRandomSatisfiedChannel("default", "longcat-2.0", 0, "", nil)
	if err != nil {
		t.Fatalf("GetRandomSatisfiedChannel returned error: %v", err)
	}
	if channel == nil {
		t.Fatal("GetRandomSatisfiedChannel returned nil channel, want channel 18 via merge reverse")
	}
	if channel.Id != 18 {
		t.Fatalf("GetRandomSatisfiedChannel returned channel %d, want 18", channel.Id)
	}
	if channelSideModel != "longcat-2.0-free" {
		t.Fatalf("channelSideModel = %q, want %q (the channel's real model)", channelSideModel, "longcat-2.0-free")
	}
}

func TestGetRandomSatisfiedChannelMergeReversePreviewAlias(t *testing.T) {
	setupMergeReverseChannelCache(t)

	channel, channelSideModel, err := GetRandomSatisfiedChannel("default", "longcat-2.0-preview", 0, "", nil)
	if err != nil {
		t.Fatalf("GetRandomSatisfiedChannel returned error: %v", err)
	}
	if channel == nil || channel.Id != 18 {
		t.Fatalf("want channel 18 via merge reverse, got %v", channel)
	}
	if channelSideModel != "longcat-2.0-free" {
		t.Fatalf("channelSideModel = %q, want %q", channelSideModel, "longcat-2.0-free")
	}
}

func TestGetRandomSatisfiedChannelExactHitKeepsModel(t *testing.T) {
	setupMergeReverseChannelCache(t)

	channel, channelSideModel, err := GetRandomSatisfiedChannel("default", "longcat-2.0-free", 0, "", nil)
	if err != nil {
		t.Fatalf("GetRandomSatisfiedChannel returned error: %v", err)
	}
	if channel == nil || channel.Id != 18 {
		t.Fatalf("want channel 18 on exact hit, got %v", channel)
	}
	if channelSideModel != "longcat-2.0-free" {
		t.Fatalf("exact hit channelSideModel = %q, want %q", channelSideModel, "longcat-2.0-free")
	}
}

func TestGetRandomSatisfiedChannelNoMatch(t *testing.T) {
	setupMergeReverseChannelCache(t)

	channel, channelSideModel, err := GetRandomSatisfiedChannel("default", "claude-opus-5", 0, "", nil)
	if err != nil {
		t.Fatalf("GetRandomSatisfiedChannel returned error: %v", err)
	}
	if channel != nil {
		t.Fatalf("want nil channel for unmatched model, got %v", channel)
	}
	if channelSideModel != "claude-opus-5" {
		t.Fatalf("no-match channelSideModel = %q, want request model", channelSideModel)
	}
}

func TestIsChannelEnabledForGroupModelMergeReverse(t *testing.T) {
	setupMergeReverseChannelCache(t)

	if !IsChannelEnabledForGroupModel("default", "longcat-2.0", 18) {
		t.Fatal("channel 18 should be enabled for longcat-2.0 via merge reverse")
	}
	if !IsChannelEnabledForGroupModel("default", "longcat-2.0-preview", 18) {
		t.Fatal("channel 18 should be enabled for longcat-2.0-preview via merge reverse")
	}
	if IsChannelEnabledForGroupModel("default", "claude-opus-5", 18) {
		t.Fatal("channel 18 should NOT be enabled for claude-opus-5")
	}
}

// TestGetRandomSatisfiedChannelMergeAndExactTogether verifies that merge-reverse
// channels participate in selection EVEN WHEN exact channels exist (AxonHub
// semantics): the candidate pool contains both, and repeated selections can
// hit either. Weight is 1 for both channels, so each should appear over a
// reasonable number of trials.
func TestGetRandomSatisfiedChannelMergeAndExactTogether(t *testing.T) {
	// channel 18: exact deepseek-v4-flash; channel 19: merge alias
	// deepseek-v4-flash-free -> deepseek-v4-flash. Add a merge rule mapping the
	// free alias to the canonical name.
	oldMemoryCache := common.MemoryCacheEnabled
	common.MemoryCacheEnabled = true
	t.Cleanup(func() { common.MemoryCacheEnabled = oldMemoryCache })

	w1 := uint(1)
	p0 := int64(0)
	ch18 := &Channel{Id: 18, Name: "exact", Status: common.ChannelStatusEnabled, Models: "deepseek-v4-flash", Type: 1, Priority: &p0, Weight: &w1}
	ch19 := &Channel{Id: 19, Name: "free", Status: common.ChannelStatusEnabled, Models: "deepseek-v4-flash-free", Type: 1, Priority: &p0, Weight: &w1}

	channelSyncLock.Lock()
	oldChannelsIDM := channelsIDM
	oldGroup2model2channels := group2model2channels
	channelsIDM = map[int]*Channel{18: ch18, 19: ch19}
	group2model2channels = map[string]map[string][]int{
		"default": {
			"deepseek-v4-flash":      {18},
			"deepseek-v4-flash-free": {19},
		},
	}
	channelSyncLock.Unlock()
	t.Cleanup(func() {
		channelSyncLock.Lock()
		channelsIDM = oldChannelsIDM
		group2model2channels = oldGroup2model2channels
		channelSyncLock.Unlock()
	})

	modelMergeMutex.Lock()
	oldExact := modelMergeExact
	oldRegex := modelMergeRegex
	oldReady := modelMergeReady
	modelMergeExact = map[string]string{"deepseek-v4-flash-free": "deepseek-v4-flash"}
	modelMergeRegex = nil
	modelMergeReady = true
	modelMergeMutex.Unlock()
	t.Cleanup(func() {
		modelMergeMutex.Lock()
		modelMergeExact = oldExact
		modelMergeRegex = oldRegex
		modelMergeReady = oldReady
		modelMergeMutex.Unlock()
		mergeReverseMutex.Lock()
		mergeReverseCache = nil
		mergeReverseMutex.Unlock()
	})

	seen := make(map[int]int)
	for i := 0; i < 200; i++ {
		channel, channelSideModel, err := GetRandomSatisfiedChannel("default", "deepseek-v4-flash", 0, "", nil)
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
		if channel == nil {
			t.Fatalf("iteration %d: nil channel", i)
		}
		seen[channel.Id]++
		if channel.Id == 19 && channelSideModel != "deepseek-v4-flash-free" {
			t.Fatalf("channel 19 must carry its real model name, got %q", channelSideModel)
		}
		if channel.Id == 18 && channelSideModel != "deepseek-v4-flash" {
			t.Fatalf("channel 18 must carry exact model name, got %q", channelSideModel)
		}
	}
	if seen[18] == 0 || seen[19] == 0 {
		t.Fatalf("both channels must be selectable, seen=%v", seen)
	}
	if seen[18]+seen[19] != 200 {
		t.Fatalf("total selections = %d, want 200", seen[18]+seen[19])
	}
}

func TestResolveChannelSideModelName(t *testing.T) {
	setupMergeReverseChannelCache(t)

	channel := &Channel{
		Id:     18,
		Models: "longcat-2.0-free",
	}

	if got := ResolveChannelSideModelName(channel, "longcat-2.0"); got != "longcat-2.0-free" {
		t.Fatalf("ResolveChannelSideModelName(longcat-2.0) = %q, want %q", got, "longcat-2.0-free")
	}
	if got := ResolveChannelSideModelName(channel, "longcat-2.0-preview"); got != "longcat-2.0-free" {
		t.Fatalf("ResolveChannelSideModelName(longcat-2.0-preview) = %q, want %q", got, "longcat-2.0-free")
	}
	if got := ResolveChannelSideModelName(channel, "longcat-2.0-free"); got != "longcat-2.0-free" {
		t.Fatalf("exact hit should stay: got %q", got)
	}
	if got := ResolveChannelSideModelName(channel, "gpt-5"); got != "gpt-5" {
		t.Fatalf("unmatched should stay: got %q", got)
	}
	if got := ResolveChannelSideModelName(nil, "longcat-2.0"); got != "longcat-2.0" {
		t.Fatalf("nil channel should pass through: got %q", got)
	}
}

func TestMergeModelNameCached(t *testing.T) {
	setupMergeReverseChannelCache(t)

	// First call computes and memoizes; second call hits the cache. Both must
	// return the same canonical result.
	for i := 0; i < 2; i++ {
		if got := MergeModelNameCached("longcat-2.0-free"); got != "longcat-2.0" {
			t.Fatalf("MergeModelNameCached(longcat-2.0-free) = %q, want %q", got, "longcat-2.0")
		}
	}
	if got := MergeModelNameCached("unrelated-model"); got != "unrelated-model" {
		t.Fatalf("unmatched should pass through: got %q", got)
	}
	if got := MergeModelNameCached(""); got != "" {
		t.Fatalf("empty should pass through: got %q", got)
	}
}

func TestMergeModelNameCachedInvalidatedOnReload(t *testing.T) {
	setupMergeReverseChannelCache(t)

	if got := MergeModelNameCached("longcat-2.0-free"); got != "longcat-2.0" {
		t.Fatalf("before reload: %q", got)
	}
	// Reload rules with a different mapping for the same alias.
	modelMergeMutex.Lock()
	modelMergeRegex = []modelMergeRegexRule{
		{
			priority: 99,
			pattern:  regexp.MustCompile(`(?i)^(?:LongCat\-2\.0(?:-(?:preview|free))?)$`),
			target:   "longcat-2.0-v2",
		},
	}
	modelMergeMutex.Unlock()
	mergeReverseMutex.Lock()
	mergeReverseCache = nil
	mergeReverseMutex.Unlock()

	if got := MergeModelNameCached("longcat-2.0-free"); got != "longcat-2.0-v2" {
		t.Fatalf("after reload: %q, want %q", got, "longcat-2.0-v2")
	}
}
