package model

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

func IsChannelEnabledForGroupModel(group string, modelName string, channelID int) bool {
	if group == "" || modelName == "" || channelID <= 0 {
		return false
	}
	if !common.MemoryCacheEnabled {
		return isChannelEnabledForGroupModelDB(group, modelName, channelID)
	}

	channelSyncLock.RLock()
	defer channelSyncLock.RUnlock()

	if group2model2channels == nil {
		return false
	}

	if isChannelIDInList(group2model2channels[group][modelName], channelID) {
		return true
	}
	normalized := ratio_setting.FormatMatchingModelName(modelName)
	if normalized != "" && normalized != modelName {
		if isChannelIDInList(group2model2channels[group][normalized], channelID) {
			return true
		}
	}
	// Model-merge reverse matching: the channel is usable when any of its
	// models merges to the requested name.
	return isChannelEnabledForGroupModelByMergeReverse(group, modelName, channelID)
}

// isChannelEnabledForGroupModelByMergeReverse reports whether the channel has a
// model in the group that merges to the requested name (memory-cache path).
func isChannelEnabledForGroupModelByMergeReverse(group string, modelName string, channelID int) bool {
	groupChannels := group2model2channels[group]
	requestCanonical := MergeModelNameCached(modelName)
	for channelModel, channelIds := range groupChannels {
		if channelModel == modelName {
			continue
		}
		if MergeModelNameCached(channelModel) != requestCanonical {
			continue
		}
		if isChannelIDInList(channelIds, channelID) {
			return true
		}
	}
	return false
}

func IsChannelEnabledForAnyGroupModel(groups []string, modelName string, channelID int) bool {
	if len(groups) == 0 {
		return false
	}
	for _, g := range groups {
		if IsChannelEnabledForGroupModel(g, modelName, channelID) {
			return true
		}
	}
	return false
}

func isChannelEnabledForGroupModelDB(group string, modelName string, channelID int) bool {
	var count int64
	err := DB.Model(&Ability{}).
		Where(commonGroupCol+" = ? and model = ? and channel_id = ? and enabled = ?", group, modelName, channelID, true).
		Count(&count).Error
	if err == nil && count > 0 {
		return true
	}
	normalized := ratio_setting.FormatMatchingModelName(modelName)
	if normalized != "" && normalized != modelName {
		count = 0
		err = DB.Model(&Ability{}).
			Where(commonGroupCol+" = ? and model = ? and channel_id = ? and enabled = ?", group, normalized, channelID, true).
			Count(&count).Error
		if err == nil && count > 0 {
			return true
		}
	}
	// Model-merge reverse matching: the channel is usable when any of its
	// models merges to the requested name (DB path).
	var channelModels []string
	if err := DB.Model(&Ability{}).
		Where(commonGroupCol+" = ? and channel_id = ? and enabled = ?", group, channelID, true).
		Distinct().Pluck("model", &channelModels).Error; err != nil {
		return false
	}
	requestCanonical := MergeModelNameCached(modelName)
	for _, channelModel := range channelModels {
		if channelModel == modelName {
			continue
		}
		if MergeModelNameCached(channelModel) == requestCanonical {
			return true
		}
	}
	return false
}

func isChannelIDInList(list []int, channelID int) bool {
	for _, id := range list {
		if id == channelID {
			return true
		}
	}
	return false
}

// ResolveChannelSideModelName returns the channel-side real model name for a
// request model name: if the channel lists the request model directly it
// returns the request model; otherwise it scans the channel's models and
// returns the first model that merges to the request model (model-merge
// reverse match), or the request model when nothing matches. Used after a
// channel is selected so the upstream relay sends the channel's real model.
func ResolveChannelSideModelName(channel *Channel, modelName string) string {
	if channel == nil || modelName == "" {
		return modelName
	}
	models := channel.GetModels()
	for _, m := range models {
		if m == modelName {
			return modelName
		}
	}
	requestCanonical := MergeModelNameCached(modelName)
	for _, m := range models {
		if m == modelName {
			continue
		}
		if MergeModelNameCached(m) == requestCanonical {
			return m
		}
	}
	return modelName
}
