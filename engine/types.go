package engine

import (
	"regexp"
	"strings"
)

// SQLType sql 类型
type SQLType string

const (
	// SELECT 查询语句
	SELECT SQLType = "SELECT"
	// DELETE 删除语句
	DELETE SQLType = "DELETE"
	// UPDATE 更新语句
	UPDATE SQLType = "UPDATE"
	// INSERT 插入语句
	INSERT SQLType = "INSERT"
	// UNKNOWN 未知语句
	UNKNOWN SQLType = "UNKNOWN"
)

var (
	selectRegex = regexp.MustCompile(`(?i)^\s*SELECT\b`)
	insertRegex = regexp.MustCompile(`(?i)^\s*INSERT\b`)
	updateRegex = regexp.MustCompile(`(?i)^\s*UPDATE\b`)
	deleteRegex = regexp.MustCompile(`(?i)^\s*DELETE\b`)

	tableRegex = regexp.MustCompile(`(?i)\b(?:from|join|into|update)\s+(\w+)\b`)
)

// ParseSQLType 解析 sql 类型
func ParseSQLType(query string) SQLType {
	query = strings.TrimSpace(query)
	if query == "" {
		return UNKNOWN
	}

	switch {
	case selectRegex.MatchString(query):
		return SELECT
	case insertRegex.MatchString(query):
		return INSERT
	case updateRegex.MatchString(query):
		return UPDATE
	case deleteRegex.MatchString(query):
		return DELETE
	default:
		return UNKNOWN
	}
}

// ParseTableName extracts the table name from a single table SQL query.
func ParseTableName(sql string) string {
	matches := tableRegex.FindStringSubmatch(strings.ToLower(sql))
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
