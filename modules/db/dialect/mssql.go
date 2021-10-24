// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package dialect

import (
	"fmt"
)

type mssql struct {
	commonDialect
}

func (mssql) GetName() string {
	return "mssql"
}

func (mssql) ShowColumns(table string) string {
	return fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '%s'", table)
}

func (mssql) ShowTables() string {
	return "SELECT * FROM information_schema.TABLES"
}
