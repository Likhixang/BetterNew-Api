package model

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"

	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Ability struct {
	Group     string  `json:"group" gorm:"type:varchar(64);primaryKey;autoIncrement:false"`
	Model     string  `json:"model" gorm:"type:varchar(255);primaryKey;autoIncrement:false"`
	ChannelId int     `json:"channel_id" gorm:"primaryKey;autoIncrement:false;index"`
	Enabled   bool    `json:"enabled"`
	Priority  *int64  `json:"priority" gorm:"bigint;default:0;index"`
	Weight    uint    `json:"weight" gorm:"default:0;index"`
	Tag       *string `json:"tag" gorm:"index"`
}

type AbilityWithChannel struct {
	Ability
	ChannelType int `json:"channel_type"`
}

func GetAllEnableAbilityWithChannels() ([]AbilityWithChannel, error) {
	var abilities []AbilityWithChannel
	err := DB.Table("abilities").
		Select("abilities.*, channels.type as channel_type").
		Joins("left join channels on abilities.channel_id = channels.id").
		Where("abilities.enabled = ?", true).
		Scan(&abilities).Error
	return abilities, err
}

func GetGroupEnabledModels(group string) []string {
	var models []string
	// Find distinct models
	DB.Table("abilities").Where(commonGroupCol+" = ? and enabled = ?", group, true).Distinct("model").Pluck("model", &models)
	return models
}

func GetEnabledModels() []string {
	var models []string
	// Find distinct models
	DB.Table("abilities").Where("enabled = ?", true).Distinct("model").Pluck("model", &models)
	return models
}

func GetAllEnableAbilities() []Ability {
	var abilities []Ability
	DB.Find(&abilities, "enabled = ?", true)
	return abilities
}

// ErrNoEnabledAbilities is returned by getPriority when the exact model has no
// enabled abilities in the group. It is a signal that the exact model name has
// no channels (as opposed to a real database error) and callers may fall back
// to model-merge reverse matching.
var ErrNoEnabledAbilities = errors.New("数据库一致性被破坏")

// isNoEnabledAbilitiesError reports whether err is the "no enabled abilities"
// sentinel from getPriority, i.e. the exact model has no channels.
func isNoEnabledAbilitiesError(err error) bool {
	return errors.Is(err, ErrNoEnabledAbilities)
}

func getPriority(group string, model string, retry int) (int, error) {

	var priorities []int
	err := DB.Model(&Ability{}).
		Select("DISTINCT(priority)").
		Where(commonGroupCol+" = ? and model = ? and enabled = ?", group, model, true).
		Order("priority DESC").              // 按优先级降序排序
		Pluck("priority", &priorities).Error // Pluck用于将查询的结果直接扫描到一个切片中

	if err != nil {
		// 处理错误
		return 0, err
	}

	if len(priorities) == 0 {
		// 如果没有查询到优先级，则返回错误
		return 0, ErrNoEnabledAbilities
	}

	// 确定要使用的优先级
	var priorityToUse int
	if retry >= len(priorities) {
		// 如果重试次数大于优先级数，则使用最小的优先级
		priorityToUse = priorities[len(priorities)-1]
	} else {
		priorityToUse = priorities[retry]
	}
	return priorityToUse, nil
}

func getChannelQuery(group string, model string, retry int) (*gorm.DB, error) {
	maxPrioritySubQuery := DB.Model(&Ability{}).Select("MAX(priority)").Where(commonGroupCol+" = ? and model = ? and enabled = ?", group, model, true)
	channelQuery := DB.Where(commonGroupCol+" = ? and model = ? and enabled = ? and priority = (?)", group, model, true, maxPrioritySubQuery)
	if retry != 0 {
		priority, err := getPriority(group, model, retry)
		if err != nil {
			return nil, err
		} else {
			channelQuery = DB.Where(commonGroupCol+" = ? and model = ? and enabled = ? and priority = ?", group, model, true, priority)
		}
	}

	return channelQuery, nil
}

// GetChannel returns a random channel that can serve the model, querying the
// database directly (used when the memory cache is disabled). It first looks
// for channels with the exact model name; when none exist it falls back to
// model-merge reverse matching: any channel model that merges to the requested
// name (per global merge rules) becomes a candidate. The second return value is
// the channel-side model name to send upstream (the requested name on exact
// hits, the merge alias on reverse hits).
func GetChannel(group string, model string, retry int, requestPath string, allowedChannelIds []int) (*Channel, string, error) {
	var abilities []Ability

	var err error = nil
	channelQuery, err := getChannelQuery(group, model, retry)
	if err == nil {
		if common.UsingMainDatabase(common.DatabaseTypeSQLite) || common.UsingMainDatabase(common.DatabaseTypePostgreSQL) {
			err = channelQuery.Order("weight DESC").Find(&abilities).Error
		} else {
			err = channelQuery.Order("weight DESC").Find(&abilities).Error
		}
		if err != nil {
			return nil, model, err
		}
	} else if !isNoEnabledAbilitiesError(err) {
		// Only "no enabled abilities" (from getPriority when retry>0 and the
		// exact model has no channels) means "exact name has no channel" and
		// falls through to model-merge reverse matching. Real database errors
		// must propagate, not be masked as "no channel available".
		return nil, model, err
	}
	// getChannelQuery errors (e.g. "数据库一致性被破坏" from getPriority when
	// the exact model has no enabled abilities and retry > 0) mean the exact
	// name has no channels; fall through to model-merge reverse matching.
	// Remember the channel-side model name per ability. On exact hits this is
	// the requested model; on model-merge reverse hits it is the channel's real
	// model name that merged to the requested name.
	channelSideModel := model
	if len(abilities) == 0 {
		// Model-merge reverse matching: find all enabled channel models in this
		// group that merge to the requested name, then query channels for each.
		mergeAbilities, mergeErr := getChannelByMergeReverse(group, model, retry, requestPath, allowedChannelIds)
		if mergeErr != nil {
			return nil, model, mergeErr
		}
		if len(mergeAbilities) == 0 {
			return nil, model, nil
		}
		abilities = mergeAbilities
		channelSideModel = abilities[0].Model
	}

	abilities = filterAbilitiesByRequestPathAndModel(abilities, requestPath, model)
	// Apply token-level channel whitelist (if any).
	if len(allowedChannelIds) > 0 {
		allowedSet := make(map[int]bool, len(allowedChannelIds))
		for _, id := range allowedChannelIds {
			allowedSet[id] = true
		}
		filtered := make([]Ability, 0, len(abilities))
		for _, ability := range abilities {
			if allowedSet[ability.ChannelId] {
				filtered = append(filtered, ability)
			}
		}
		abilities = filtered
	}
	channel := Channel{}
	if len(abilities) > 0 {
		// Priority layering with retry (mirrors the memory-cache path): the
		// exact query returns a single priority layer selected by retry, while
		// merge-reverse abilities span all layers, so pick the layer for this
		// retry index here.
		uniquePriorities := make(map[int64]bool)
		for _, ability_ := range abilities {
			if ability_.Priority != nil {
				uniquePriorities[*ability_.Priority] = true
			}
		}
		var sortedUniquePriorities []int64
		for p := range uniquePriorities {
			sortedUniquePriorities = append(sortedUniquePriorities, p)
		}
		sort.Slice(sortedUniquePriorities, func(i, j int) bool {
			return sortedUniquePriorities[i] > sortedUniquePriorities[j]
		})
		if retry >= len(sortedUniquePriorities) {
			retry = len(sortedUniquePriorities) - 1
		}
		targetPriority := int64(0)
		if len(sortedUniquePriorities) > 0 {
			targetPriority = sortedUniquePriorities[retry]
		}

		// Restrict to the target priority layer.
		layerAbilities := make([]Ability, 0, len(abilities))
		for _, ability_ := range abilities {
			if ability_.Priority == nil || *ability_.Priority == targetPriority {
				layerAbilities = append(layerAbilities, ability_)
			}
		}

		// Randomly choose one within the layer by weight
		weightSum := uint(0)
		for _, ability_ := range layerAbilities {
			weightSum += ability_.Weight + 10
		}
		// Randomly choose one
		weight := common.GetRandomInt(int(weightSum))
		for _, ability_ := range layerAbilities {
			weight -= int(ability_.Weight) + 10
			//log.Printf("weight: %d, ability weight: %d", weight, *ability_.Weight)
			if weight <= 0 {
				channel.Id = ability_.ChannelId
				channelSideModel = ability_.Model
				break
			}
		}
	} else {
		return nil, model, nil
	}
	err = DB.First(&channel, "id = ?", channel.Id).Error
	return &channel, channelSideModel, err
}

// getChannelByMergeReverse finds abilities whose channel-side model name merges
// to the requested model name (per global merge rules). It queries the group's
// enabled channel models, checks each against MergeModelName, and returns all
// enabled abilities for the matching channel-side model names (single batched
// query, all priority layers). Only channels that are currently enabled
// participate (a manually disabled channel may still carry stale enabled
// abilities, which must not be routed to).
func getChannelByMergeReverse(group string, model string, retry int, requestPath string, allowedChannelIds []int) ([]Ability, error) {
	var channelModels []string
	if err := DB.Model(&Ability{}).
		Where(commonGroupCol+" = ? and enabled = ?", group, true).
		Distinct().Pluck("model", &channelModels).Error; err != nil {
		return nil, err
	}
	requestCanonical := MergeModelNameCached(model)
	matchedModels := make([]string, 0)
	for _, channelModel := range channelModels {
		if channelModel == model {
			continue
		}
		if MergeModelNameCached(channelModel) != requestCanonical {
			continue
		}
		matchedModels = append(matchedModels, channelModel)
	}
	if len(matchedModels) == 0 {
		return nil, nil
	}
	// Batch query all enabled abilities for the matched models in one round
	// trip instead of one query per model.
	var result []Ability
	if err := DB.Model(&Ability{}).
		Where(commonGroupCol+" = ? and enabled = ? and model IN ?", group, true, matchedModels).
		Order("weight DESC").
		Find(&result).Error; err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	// Filter out channels that are not currently enabled: a channel may be
	// manually disabled while its abilities rows still say enabled.
	channelIds := make([]int, 0, len(result))
	seen := make(map[int]struct{}, len(result))
	for _, ability := range result {
		if _, ok := seen[ability.ChannelId]; ok {
			continue
		}
		seen[ability.ChannelId] = struct{}{}
		channelIds = append(channelIds, ability.ChannelId)
	}
	var channels []*Channel
	if err := DB.Where("id IN ?", channelIds).Find(&channels).Error; err != nil {
		return nil, err
	}
	enabledIds := make(map[int]struct{}, len(channels))
	for _, ch := range channels {
		if ch.Status == common.ChannelStatusEnabled {
			enabledIds[ch.Id] = struct{}{}
		}
	}
	filtered := make([]Ability, 0, len(result))
	for _, ability := range result {
		if _, ok := enabledIds[ability.ChannelId]; ok {
			filtered = append(filtered, ability)
		}
	}
	return filtered, nil
}

// filterAbilitiesByRequestPathAndModel restricts candidates by request path and
// model for the DB (non-memory-cache) selection path. Only Advanced Custom
// (type 58) channels are path-checked: kept only when one of their routes matches
// requestPath and model; all other channel types always pass. When requestPath is
// empty, filtering is skipped.
func filterAbilitiesByRequestPathAndModel(abilities []Ability, requestPath string, model string) []Ability {
	if requestPath == "" || len(abilities) == 0 {
		return abilities
	}

	channelIds := make([]int, 0, len(abilities))
	seen := make(map[int]struct{}, len(abilities))
	for _, ability := range abilities {
		if _, ok := seen[ability.ChannelId]; ok {
			continue
		}
		seen[ability.ChannelId] = struct{}{}
		channelIds = append(channelIds, ability.ChannelId)
	}

	var channels []*Channel
	if err := DB.Where("id IN ?", channelIds).Find(&channels).Error; err != nil {
		// On error, fall back to unfiltered candidates to avoid blocking selection
		return abilities
	}

	advancedConfigs := make(map[int]*dto.AdvancedCustomConfig)
	for _, channel := range channels {
		if channel.Type == constant.ChannelTypeAdvancedCustom {
			advancedConfigs[channel.Id] = channel.GetOtherSettings().AdvancedCustom
		}
	}

	filtered := make([]Ability, 0, len(abilities))
	for _, ability := range abilities {
		config, isAdvancedCustom := advancedConfigs[ability.ChannelId]
		if !isAdvancedCustom {
			filtered = append(filtered, ability)
			continue
		}
		if config != nil && config.SupportsPathForModel(requestPath, model) {
			filtered = append(filtered, ability)
		}
	}
	return filtered
}

func (channel *Channel) AddAbilities(tx *gorm.DB) error {
	models_ := strings.Split(channel.Models, ",")
	groups_ := strings.Split(channel.Group, ",")
	abilitySet := make(map[string]struct{})
	abilities := make([]Ability, 0, len(models_))
	for _, model := range models_ {
		for _, group := range groups_ {
			key := group + "|" + model
			if _, exists := abilitySet[key]; exists {
				continue
			}
			abilitySet[key] = struct{}{}
			ability := Ability{
				Group:     group,
				Model:     model,
				ChannelId: channel.Id,
				Enabled:   channel.Status == common.ChannelStatusEnabled,
				Priority:  channel.Priority,
				Weight:    uint(channel.GetWeight()),
				Tag:       channel.Tag,
			}
			abilities = append(abilities, ability)
		}
	}
	if len(abilities) == 0 {
		return nil
	}
	// choose DB or provided tx
	useDB := DB
	if tx != nil {
		useDB = tx
	}
	for _, chunk := range lo.Chunk(abilities, 50) {
		err := useDB.Clauses(clause.OnConflict{DoNothing: true}).Create(&chunk).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (channel *Channel) DeleteAbilities() error {
	return DB.Where("channel_id = ?", channel.Id).Delete(&Ability{}).Error
}

// UpdateAbilities updates abilities of this channel.
// Make sure the channel is completed before calling this function.
func (channel *Channel) UpdateAbilities(tx *gorm.DB) error {
	isNewTx := false
	// 如果没有传入事务，创建新的事务
	if tx == nil {
		tx = DB.Begin()
		if tx.Error != nil {
			return tx.Error
		}
		isNewTx = true
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
	}

	// First delete all abilities of this channel
	err := tx.Where("channel_id = ?", channel.Id).Delete(&Ability{}).Error
	if err != nil {
		if isNewTx {
			tx.Rollback()
		}
		return err
	}

	// Then add new abilities
	models_ := strings.Split(channel.Models, ",")
	groups_ := strings.Split(channel.Group, ",")
	abilitySet := make(map[string]struct{})
	abilities := make([]Ability, 0, len(models_))
	for _, model := range models_ {
		for _, group := range groups_ {
			key := group + "|" + model
			if _, exists := abilitySet[key]; exists {
				continue
			}
			abilitySet[key] = struct{}{}
			ability := Ability{
				Group:     group,
				Model:     model,
				ChannelId: channel.Id,
				Enabled:   channel.Status == common.ChannelStatusEnabled,
				Priority:  channel.Priority,
				Weight:    uint(channel.GetWeight()),
				Tag:       channel.Tag,
			}
			abilities = append(abilities, ability)
		}
	}

	if len(abilities) > 0 {
		for _, chunk := range lo.Chunk(abilities, 50) {
			err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&chunk).Error
			if err != nil {
				if isNewTx {
					tx.Rollback()
				}
				return err
			}
		}
	}

	// 如果是新创建的事务，需要提交
	if isNewTx {
		return tx.Commit().Error
	}

	return nil
}

func UpdateAbilityStatus(channelId int, status bool) error {
	return DB.Model(&Ability{}).Where("channel_id = ?", channelId).Select("enabled").Update("enabled", status).Error
}

func UpdateAbilityStatusByTag(tag string, status bool) error {
	return DB.Model(&Ability{}).Where("tag = ?", tag).Select("enabled").Update("enabled", status).Error
}

func UpdateAbilityByTag(tag string, newTag *string, priority *int64, weight *uint) error {
	ability := Ability{}
	if newTag != nil {
		ability.Tag = newTag
	}
	if priority != nil {
		ability.Priority = priority
	}
	if weight != nil {
		ability.Weight = *weight
	}
	return DB.Model(&Ability{}).Where("tag = ?", tag).Updates(ability).Error
}

var fixLock = sync.Mutex{}

func FixAbility() (int, int, error) {
	lock := fixLock.TryLock()
	if !lock {
		return 0, 0, errors.New("已经有一个修复任务在运行中，请稍后再试")
	}
	defer fixLock.Unlock()

	// truncate abilities table
	if common.UsingMainDatabase(common.DatabaseTypeSQLite) {
		err := DB.Exec("DELETE FROM abilities").Error
		if err != nil {
			common.SysLog(fmt.Sprintf("Delete abilities failed: %s", err.Error()))
			return 0, 0, err
		}
	} else {
		err := DB.Exec("TRUNCATE TABLE abilities").Error
		if err != nil {
			common.SysLog(fmt.Sprintf("Truncate abilities failed: %s", err.Error()))
			return 0, 0, err
		}
	}
	var channels []*Channel
	// Find all channels
	err := DB.Model(&Channel{}).Find(&channels).Error
	if err != nil {
		return 0, 0, err
	}
	if len(channels) == 0 {
		return 0, 0, nil
	}
	successCount := 0
	failCount := 0
	for _, chunk := range lo.Chunk(channels, 50) {
		ids := lo.Map(chunk, func(c *Channel, _ int) int { return c.Id })
		// Delete all abilities of this channel
		err = DB.Where("channel_id IN ?", ids).Delete(&Ability{}).Error
		if err != nil {
			common.SysLog(fmt.Sprintf("Delete abilities failed: %s", err.Error()))
			failCount += len(chunk)
			continue
		}
		// Then add new abilities
		for _, channel := range chunk {
			err = channel.AddAbilities(nil)
			if err != nil {
				common.SysLog(fmt.Sprintf("Add abilities for channel %d failed: %s", channel.Id, err.Error()))
				failCount++
			} else {
				successCount++
			}
		}
	}
	InitChannelCache()
	return successCount, failCount, nil
}
