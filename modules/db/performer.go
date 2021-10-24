// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"database/sql"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"strings"
)

// CommonQuery is a common method of query.
func CommonQuery(db *sql.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rs, err := db.Query(query, args...)
	if err != nil {
		logger.Errorf("error on sql query: %s\nwith args: %s", query, utils.JSON(args))
		panic(err)
	}
	defer rs.Close()

	col, err := rs.Columns()
	if err != nil { return nil, err }

	typeVals, err := rs.ColumnTypes()
	if err != nil { return nil, err }

	typeNames := make([]string, len(typeVals))
	for i, tv := range typeVals {
		typeNames[i] = strings.ToUpper(utils.RexCommonQuery.ReplaceAllString(tv.DatabaseTypeName(), ""))
	}

	// TODO: regular expressions for sqlite, use the dialect module
	// tell the driver to reduce the performance loss

	nCol   := len(col)
	res    := make([]map[string]interface{}, 0, 32)
	colVar := make([]interface{}, nCol)

	for rs.Next() {
		for i, typeName := range typeNames {
			colVar[i] = GetColVarType(typeName)
			//SetColVarType(colVar, i, typeName)
		}
		if err := rs.Scan(colVar...); err != nil {
			return nil, err
		}
		row := make(map[string]interface{}, nCol)
		for i, c := range col {
			row[c] = GetResultValue(colVar[i], typeNames[i])
			//SetResultValue(row, c, colVar[i], typeNames[i])
		}
		res = append(res, row)
	}

	if err := rs.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// CommonExec is a common method of exec.
func CommonExec(db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	rs, err := db.Exec(query, args...)
	if err != nil { return nil, err }
	return rs, nil
}

// CommonQueryWithTx is a common method of query.
func CommonQueryWithTx(tx *sql.Tx, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rs, err := tx.Query(query, args...)
	if err != nil { panic(err) }
	defer rs.Close()

	col, err := rs.Columns()
	if err != nil { return nil, err }

	typeVal, err := rs.ColumnTypes()
	if err != nil { return nil, err }

	// TODO: regular expressions for sqlite, use the dialect module
	// tell the drive to reduce the performance loss
	nCol := len(col)
	res  := make([]map[string]interface{}, 0, 32)
	r    := utils.RexCommonQuery

	for rs.Next() {
		colVar := make([]interface{}, len(col))
		for i, tv := range typeVal {
			typeName := strings.ToUpper(r.ReplaceAllString(tv.DatabaseTypeName(), ""))
			colVar[i] = GetColVarType(typeName)
			//SetColVarType(colVar, i, typeName)
		}
		if err := rs.Scan(colVar...); err != nil {
			return nil, err
		}
		row := make(map[string]interface{}, nCol)
		for i, c := range col {
			typeName := strings.ToUpper(r.ReplaceAllString(typeVal[i].DatabaseTypeName(), ""))
			row[c]    = GetResultValue(colVar[i], typeName)
			//SetResultValue(row, c, colVar[i], typeName)
		}
		res = append(res, row)
	}

	if err := rs.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

// CommonExecWithTx is a common method of exec.
func CommonExecWithTx(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	rs, err := tx.Exec(query, args...)
	if err != nil { return nil, err }
	return rs, nil
}

// CommonBeginTxWithLevel starts a transaction with given transaction isolation level and db connection.
func CommonBeginTxWithLevel(db *sql.DB, level sql.IsolationLevel) *sql.Tx {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: level})
	if err != nil { panic(err) }
	return tx
}
