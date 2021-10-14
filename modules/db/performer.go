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
	if rs != nil {
		defer rs.Close()
	}

	col, colErr := rs.Columns()
	if colErr != nil {
		return nil, colErr
	}

	typeVals, err := rs.ColumnTypes()
	if err != nil {
		return nil, err
	}
	typeNames := make([]string, len(typeVals))
	for i, tv := range typeVals {
		typeNames[i] = strings.ToUpper(utils.RexCommonQuery.ReplaceAllString(tv.DatabaseTypeName(), ""))
	}

	// TODO: regular expressions for sqlite, use the dialect module
	// tell the driver to reduce the performance loss

	res    := make([]map[string]interface{}, 0, 16)
	colVar := make([]interface{}, len(col))

	for rs.Next() {
		for i, typeName := range typeNames {
			SetColVarType(colVar, i, typeName)
		}
		row := make(map[string]interface{}, len(col))
		if scanErr := rs.Scan(colVar...); scanErr != nil {
			return nil, scanErr
		}
		for i, c := range col {
			SetResultValue(row, c, colVar[i], typeNames[i])
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
	if err != nil {
		return nil, err
	}
	return rs, nil
}

// CommonQueryWithTx is a common method of query.
func CommonQueryWithTx(tx *sql.Tx, query string, args ...interface{}) ([]map[string]interface{}, error) {

	rs, err := tx.Query(query, args...)

	if err != nil {
		panic(err)
	}

	if rs != nil {
		defer rs.Close()
	}

	col, colErr := rs.Columns()

	if colErr != nil {
		return nil, colErr
	}

	typeVal, err := rs.ColumnTypes()
	if err != nil {
		return nil, err
	}

	// TODO: regular expressions for sqlite, use the dialect module
	// tell the drive to reduce the performance loss
	results := make([]map[string]interface{}, 0)

	r := utils.RexCommonQuery
	for rs.Next() {
		var colVar = make([]interface{}, len(col))
		for i := 0; i < len(col); i++ {
			typeName := strings.ToUpper(r.ReplaceAllString(typeVal[i].DatabaseTypeName(), ""))
			SetColVarType(colVar, i, typeName)
		}
		result := make(map[string]interface{})
		if scanErr := rs.Scan(colVar...); scanErr != nil {
			return nil, scanErr
		}
		for j := 0; j < len(col); j++ {
			typeName := strings.ToUpper(r.ReplaceAllString(typeVal[j].DatabaseTypeName(), ""))
			SetResultValue(result, col[j], colVar[j], typeName)
		}
		results = append(results, result)
	}
	if err := rs.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// CommonExecWithTx is a common method of exec.
func CommonExecWithTx(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	rs, err := tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

// CommonBeginTxWithLevel starts a transaction with given transaction isolation level and db connection.
func CommonBeginTxWithLevel(db *sql.DB, level sql.IsolationLevel) *sql.Tx {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: level})
	if err != nil {
		panic(err)
	}
	return tx
}
