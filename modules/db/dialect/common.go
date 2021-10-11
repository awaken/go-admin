package dialect

import (
	"fmt"
	"github.com/GoAdminGroup/go-admin/modules/utils"
)

type commonDialect struct {
	delimiter  string
	delimiter2 string
}

func (c commonDialect) Insert(comp *SQLComponent) string {
	comp.prepareInsert(c.delimiter, c.delimiter2)
	return comp.Statement
}

func (c commonDialect) Delete(comp *SQLComponent) string {
	comp.Statement = utils.StrConcat("DELETE FROM ", c.WrapTableName(comp), comp.getWheres(c.delimiter, c.delimiter2))
	return comp.Statement
}

func (c commonDialect) Update(comp *SQLComponent) string {
	comp.prepareUpdate(c.delimiter, c.delimiter2)
	return comp.Statement
}

func (c commonDialect) Count(comp *SQLComponent) string {
	comp.prepareUpdate(c.delimiter, c.delimiter2)
	return comp.Statement
}

func (c commonDialect) Select(comp *SQLComponent) string {
	comp.Statement = utils.StrConcat("SELECT ", comp.getFields(c.delimiter, c.delimiter2), " FROM ", c.WrapTableName(comp), comp.getJoins(c.delimiter, c.delimiter2),
		comp.getWheres(c.delimiter, c.delimiter2), comp.getGroupBy(), comp.getOrderBy(), comp.getLimit(), comp.getOffset())
	return comp.Statement
}

func (c commonDialect) ShowColumns(table string) string {
	return fmt.Sprintf("SELECT * FROM information_schema.columns WHERE table_name='%s'", table)
}

func (c commonDialect) GetName() string {
	return "common"
}

func (c commonDialect) WrapTableName(comp *SQLComponent) string {
	return utils.StrConcat(c.delimiter, comp.TableName, c.delimiter2)
}

func (c commonDialect) ShowTables() string {
	return "SHOW TABLES"
}

func (c commonDialect) GetDelimiter() string {
	return c.delimiter
}

func (c commonDialect) GetDelimiter2() string {
	return c.delimiter2
}

func (c commonDialect) GetDelimiters() []string {
	return []string{c.delimiter, c.delimiter2}
}
