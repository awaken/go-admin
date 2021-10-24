// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package dialect

import (
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"strings"
)

// Dialect is methods set of different driver.
type Dialect interface {
	// GetName get dialect's name
	GetName() string

	// ShowColumns show columns of specified table
	ShowColumns(table string) string

	// ShowTables show tables of database
	ShowTables() string

	// Insert
	Insert(comp *SQLComponent) string

	// Delete
	Delete(comp *SQLComponent) string

	// Update
	Update(comp *SQLComponent) string

	// Select
	Select(comp *SQLComponent) string

	// GetDelimiter return the delimiter of Dialect.
	GetDelimiter() string
}

// GetDialect return the default Dialect.
func GetDialect() Dialect {
	return GetDialectByDriver(config.GetDatabases().GetDefault().Driver)
}

// GetDialectByDriver return the Dialect of given driver.
func GetDialectByDriver(driver string) Dialect {
	switch driver {
	case "mysql":
		return mysql{ commonDialect: commonDialect{ delimiter: "`", delimiter2: "`" }}
	case "mssql":
		return mssql{ commonDialect: commonDialect{ delimiter: "[", delimiter2: "]" }}
	case "postgresql":
		return postgresql{ commonDialect: commonDialect{ delimiter: `"`, delimiter2: `"` }}
	case "sqlite":
		return sqlite{ commonDialect: commonDialect{ delimiter: "`", delimiter2: "`" }}
	default:
		return commonDialect{ delimiter: "`", delimiter2: "`" }
	}
}

// H is a shorthand of map.
type H map[string]interface{}

// SQLComponent is a sql components set.
type SQLComponent struct {
	Fields     []string
	Functions  []string
	TableName  string
	Wheres     []Where
	Leftjoins  []Join
	Args       []interface{}
	Order      string
	Offset     string
	Limit      string
	WhereRaws  string
	UpdateRaws []RawUpdate
	Group      string
	Statement  string
	Values     H
}

// Where contains the operation and field.
type Where struct {
	Operation string
	Field     string
	Qmark     string
}

// Join contains the table and field and operation.
type Join struct {
	Table     string
	FieldA    string
	Operation string
	FieldB    string
}

// RawUpdate contains the expression and arguments.
type RawUpdate struct {
	Expression string
	Args       []interface{}
}

// *******************************
// internal help function
// *******************************

func (sql *SQLComponent) getLimit() string {
	if sql.Limit == "" { return "" }
	return " LIMIT " + sql.Limit
}

func (sql *SQLComponent) getOffset() string {
	if sql.Offset == "" { return "" }
	return " OFFSET " + sql.Offset
}

func (sql *SQLComponent) getOrderBy() string {
	if sql.Order == "" { return "" }
	return " ORDER BY " + sql.Order
}

func (sql *SQLComponent) getGroupBy() string {
	if sql.Group == "" { return "" }
	return " GROUP BY " + sql.Group
}

func (sql *SQLComponent) getJoins(delimiter, delimiter2 string) string {
	if len(sql.Leftjoins) == 0 { return "" }

	var sb strings.Builder
	sb.Grow(256)

	for _, join := range sql.Leftjoins {
		sb.WriteString(" LEFT JOIN ")
		sb.WriteString(delimiter)
		sb.WriteString(join.Table)
		sb.WriteString(delimiter2)
		sb.WriteString(" ON ")
		sb.WriteString(sql.processLeftJoinField(join.FieldA, delimiter, delimiter2))
		sb.WriteByte(' ')
		sb.WriteString(join.Operation)
		sb.WriteByte(' ')
		sb.WriteString(sql.processLeftJoinField(join.FieldB, delimiter, delimiter2))
	}

	return sb.String()
}

func (sql *SQLComponent) processLeftJoinField(field, delimiter, delimiter2 string) string {
	if field == "" { return "" }
	name1, name2 := utils.StrSplitByte2(field, '.')
	if name2 == "" {
		return utils.StrConcat(delimiter, name1, delimiter2)
	}
	return utils.StrConcat(delimiter, name1, delimiter2, ".", delimiter, name2, delimiter2)
}

func (sql *SQLComponent) getFields(delimiter, delimiter2 string) string {
	if len(sql.Fields) == 0 { return "*" }

	var sb strings.Builder
	sb.Grow(128)

	if len(sql.Leftjoins) == 0 {
		for k, field := range sql.Fields {
			if sb.Len() > 0 { sb.WriteByte(',') }
			funcName := sql.Functions[k]
			if funcName != "" {
				sb.WriteString(funcName)
				sb.WriteByte('(')
			}
			if field == "*" {
				sb.WriteByte('*')
			} else {
				sb.WriteString(delimiter)
				sb.WriteString(field)
				sb.WriteString(delimiter2)
			}
			if funcName != "" {
				sb.WriteByte(')')
			}
		}
	} else {
		for _, field := range sql.Fields {
			if sb.Len() > 0 { sb.WriteByte(',') }
			v1, v2 := utils.StrSplitByte2(field, '.')
			if v1 == "*" {
				sb.WriteByte('*')
			} else {
				sb.WriteString(delimiter)
				sb.WriteString(v1)
				sb.WriteString(delimiter2)
			}
			if v2 != "" {
				sb.WriteByte('.')
				if v2 == "*" {
					sb.WriteByte('*')
				} else {
					sb.WriteString(delimiter)
					sb.WriteString(v2)
					sb.WriteString(delimiter2)
				}
			}
		}
	}

	return sb.String()
}

/*func wrap(delimiter, delimiter2, field string) string {
	if field == "*" { return "*" }
	return utils.StrConcat(delimiter, field, delimiter2)
}*/

func (sql *SQLComponent) getWheres(delimiter, delimiter2 string) string {
	const cWhere = " WHERE "

	if len(sql.Wheres) == 0 {
		if sql.WhereRaws == "" { return "" }
		return cWhere + sql.WhereRaws
	}

	var sb strings.Builder
	sb.Grow(512)
	sb.WriteString(cWhere)

	for _, where := range sql.Wheres {
		if sb.Len() > len(cWhere) { sb.WriteString(" AND ") }
		v1, v2 := utils.StrSplitByte2(where.Field, '.')
		if v2 != "" {
			sb.WriteString(v1)
			sb.WriteByte('.')
			if v2 == "*" {
				sb.WriteByte('*')
			} else {
				sb.WriteString(delimiter)
				sb.WriteString(v2)
				sb.WriteString(delimiter2)
			}
		} else {
			if where.Field == "*" {
				sb.WriteByte('*')
			} else {
				sb.WriteString(delimiter)
				sb.WriteString(where.Field)
				sb.WriteString(delimiter2)
			}
		}
		sb.WriteByte(' ')
		sb.WriteString(where.Operation)
		sb.WriteByte(' ')
		sb.WriteString(where.Qmark)
	}

	if sql.WhereRaws != "" {
		sb.WriteString(" AND ")
		sb.WriteString(sql.WhereRaws)
	}

	return sb.String()
}

func (sql *SQLComponent) prepareUpdate(delimiter, delimiter2 string) {
	var sb strings.Builder
	sb.Grow(512)
	sb.WriteString("UPDATE ")
	sb.WriteString(delimiter)
	sb.WriteString(sql.TableName)
	sb.WriteString(delimiter2)
	sb.WriteString(" SET ")

	args  := make([]interface{}, 0, len(sql.Values) + len(sql.UpdateRaws) * 4 + len(sql.Args))
	first := true

	for key, value := range sql.Values {
		if first {
			first = false
		} else {
			sb.WriteByte(',')
		}
		if key == "*" {
			sb.WriteByte('*')
		} else {
			sb.WriteString(delimiter)
			sb.WriteString(key)
			sb.WriteString(delimiter2)
		}
		sb.WriteString(" = ?")
		args = append(args, value)
	}

	for _, u := range sql.UpdateRaws {
		if first {
			first = false
		} else {
			sb.WriteByte(',')
		}
		sb.WriteString(u.Expression)
		args = append(args, u.Args...)
	}

	sql.Args = append(args, sql.Args...)

	sb.WriteString(sql.getWheres(delimiter, delimiter2))
	sql.Statement = sb.String()
}

func (sql *SQLComponent) prepareInsert(delimiter, delimiter2 string) {
	var sb strings.Builder
	sb.Grow(512)
	sb.WriteString("INSERT INTO ")
	sb.WriteString(delimiter)
	sb.WriteString(sql.TableName)
	sb.WriteString(delimiter2)
	sb.WriteString(" (")
	first := true

	for key, value := range sql.Values {
		if first {
			first = false
		} else {
			sb.WriteByte(',')
		}
		sb.WriteString(delimiter)
		sb.WriteString(key)
		sb.WriteString(delimiter2)
		sql.Args = append(sql.Args, value)
	}

	sb.WriteString(") VALUES (?")
	for i := len(sql.Values); i > 1; i-- {
		sb.WriteString(",?")
	}
	sb.WriteByte(')')

	sql.Statement = sb.String()
}
