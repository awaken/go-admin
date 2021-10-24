// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package dialect

import (
	"fmt"
	"github.com/GoAdminGroup/go-admin/modules/utils"
)

type postgresql struct {
	commonDialect
}

func (postgresql) GetName() string {
	return "postgresql"
}

func (postgresql) ShowTables() string {
	return "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema';"
}

func (postgresql) ShowColumns(table string) string {
	name1, name2 := utils.StrSplitByte2(table, '.')
	if name2 != "" {
		return fmt.Sprintf("SELECT * FROM information_schema.columns WHERE table_name = '%s' AND table_schema = '%s'", name2, name1)
	} else {
		return fmt.Sprintf("SELECT * FROM information_schema.columns WHERE table_name = '%s'", table)
	}
}
