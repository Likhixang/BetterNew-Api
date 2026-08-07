package model

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func dryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	require.NoError(t, err)
	return db
}

func sqlAndVars(t *testing.T, db *gorm.DB) (string, []interface{}) {
	t.Helper()
	var rows []Log
	stmt := db.Find(&rows).Statement
	return stmt.SQL.String(), stmt.Vars
}

func TestApplyExplicitLogTextFilterDefaultsToFuzzy(t *testing.T) {
	originalLogDatabaseType := common.LogDatabaseType()
	t.Cleanup(func() {
		common.SetLogDatabaseType(originalLogDatabaseType)
	})
	common.SetLogDatabaseType(common.DatabaseTypeSQLite)

	db := dryRunDB(t)
	filtered, err := applyExplicitLogTextFilter(db, "logs.model_name", "gpt-4")
	require.NoError(t, err)

	sql, vars := sqlAndVars(t, filtered)
	assert.Contains(t, sql, "LIKE ? ESCAPE '!'")
	assert.Equal(t, []interface{}{"%gpt-4%"}, vars)
}

func TestApplyExplicitLogTextFilterKeepsExplicitWildcard(t *testing.T) {
	originalLogDatabaseType := common.LogDatabaseType()
	t.Cleanup(func() {
		common.SetLogDatabaseType(originalLogDatabaseType)
	})
	common.SetLogDatabaseType(common.DatabaseTypeSQLite)

	db := dryRunDB(t)
	filtered, err := applyExplicitLogTextFilter(db, "logs.model_name", "gpt-%")
	require.NoError(t, err)

	sql, vars := sqlAndVars(t, filtered)
	assert.Contains(t, sql, "LIKE ? ESCAPE '!'")
	// 显式通配符模式：不额外包装 %，_ 仍被转义
	assert.Equal(t, []interface{}{"gpt-%"}, vars)
}

func TestApplyExplicitLogTextFilterEscapesUnderscore(t *testing.T) {
	originalLogDatabaseType := common.LogDatabaseType()
	t.Cleanup(func() {
		common.SetLogDatabaseType(originalLogDatabaseType)
	})
	common.SetLogDatabaseType(common.DatabaseTypeSQLite)

	db := dryRunDB(t)
	filtered, err := applyExplicitLogTextFilter(db, "logs.model_name", "gpt_4")
	require.NoError(t, err)

	sql, vars := sqlAndVars(t, filtered)
	assert.Contains(t, sql, "LIKE ? ESCAPE '!'")
	// 下划线作为字面量匹配，避免被当作单字符通配符
	assert.Equal(t, []interface{}{"%gpt!_4%"}, vars)
}

func TestApplyExplicitLogTextFilterEmptyValue(t *testing.T) {
	db := dryRunDB(t)
	filtered, err := applyExplicitLogTextFilter(db, "logs.model_name", "")
	require.NoError(t, err)

	sql, _ := sqlAndVars(t, filtered)
	assert.NotContains(t, sql, "LIKE")
	assert.NotContains(t, sql, "model_name")
}

func TestApplyExplicitLogTextFilterClickHouse(t *testing.T) {
	originalLogDatabaseType := common.LogDatabaseType()
	t.Cleanup(func() {
		common.SetLogDatabaseType(originalLogDatabaseType)
	})
	common.SetLogDatabaseType(common.DatabaseTypeClickHouse)

	db := dryRunDB(t)
	filtered, err := applyExplicitLogTextFilter(db, "logs.model_name", "gpt-4")
	require.NoError(t, err)

	sql, vars := sqlAndVars(t, filtered)
	assert.Contains(t, sql, "LIKE ?")
	assert.False(t, strings.Contains(sql, "ESCAPE"), "ClickHouse 分支不使用 ESCAPE 子句")
	assert.Equal(t, []interface{}{"%gpt-4%"}, vars)

	db = dryRunDB(t)
	filtered, err = applyExplicitLogTextFilter(db, "logs.model_name", "gpt_4")
	require.NoError(t, err)

	sql, vars = sqlAndVars(t, filtered)
	assert.Contains(t, sql, "LIKE ?")
	assert.Equal(t, []interface{}{`%gpt\_4%`}, vars)
}
