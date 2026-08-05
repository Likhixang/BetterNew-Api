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
package controller

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"

	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func GetAllModelMerges(c *gin.Context) {
	merges, err := model.GetAllModelMerges()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    merges,
	})
}

func CreateModelMerge(c *gin.Context) {
	var merge model.ModelMerge
	if err := c.ShouldBindJSON(&merge); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	merge.Id = 0
	if merge.TargetModel == "" || merge.Alias == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "target_model and alias are required"})
		return
	}
	if merge.MatchType == model.ModelMergeMatchRegex {
		if err := validateModelMergeRegex(merge.Alias); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			return
		}
	}
	if err := merge.Insert(); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    merge,
	})
}

func UpdateModelMerge(c *gin.Context) {
	var merge model.ModelMerge
	if err := c.ShouldBindJSON(&merge); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if merge.Id == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "id is required"})
		return
	}
	if merge.TargetModel == "" || merge.Alias == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "target_model and alias are required"})
		return
	}
	if merge.MatchType == model.ModelMergeMatchRegex {
		if err := validateModelMergeRegex(merge.Alias); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			return
		}
	}
	if err := merge.Update(); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    merge,
	})
}

func DeleteModelMerge(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid id"})
		return
	}
	merge, err := model.GetModelMergeById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := merge.Delete(); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func validateModelMergeRegex(alias string) error {
	// Compile is handled by InitModelMergeCache as well, but validate early so
	// the API returns a clear error instead of silently skipping the rule.
	_, err := regexp.Compile(alias)
	if err != nil {
		return errors.New("invalid regex alias: " + err.Error())
	}
	return nil
}

// PreviewModelMerge returns, for a draft alias/match_type, every channel whose
// configured models match the rule. Used by the merge drawer's live preview so
// admins see exactly which channels and models a rule would affect before saving.
func PreviewModelMerge(c *gin.Context) {
	var req struct {
		Alias     string `json:"alias"`
		MatchType int    `json:"match_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	channels, err := model.GetAllChannelsOmitKey()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	result, err := model.PreviewModelMergeMatches(channels, req.Alias, req.MatchType)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    result,
	})
}
