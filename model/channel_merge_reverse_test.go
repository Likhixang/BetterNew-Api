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
// LongCat-2.0(-preview|-free)? to longcat-2.0.
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
	channelsIDM = map[int]*Channel{18: channel}
	group2model2channels = map[string]map[string][]int{
		"default": {
			"longcat-2.0-free": {18},
		},
	}
	channelSyncLock.Unlock()

	modelMergeMutex.Lock()
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
