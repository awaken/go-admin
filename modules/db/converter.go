// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"github.com/GoAdminGroup/go-admin/modules/utils"
)

func GetColVarType(typeName string) interface{} {
	dt := DT(typeName)
	switch {
	case Contains(dt, BoolTypeList  ): return &sql.NullBool{}
	case Contains(dt, IntTypeList   ): return &sql.NullInt64{}
	case Contains(dt, FloatTypeList ): return &sql.NullFloat64{}
	case Contains(dt, UintTypeList  ): var s []uint8; return &s
	case Contains(dt, StringTypeList): return &sql.NullString{}
	default                          : var s interface{}; return &s
	}
}

// SetColVarType set the column type.
func SetColVarType(colVar []interface{}, i int, typeName string) {
	dt := DT(typeName)
	switch {
	case Contains(dt, BoolTypeList):
		var s sql.NullBool
		colVar[i] = &s
	case Contains(dt, IntTypeList):
		var s sql.NullInt64
		colVar[i] = &s
	case Contains(dt, FloatTypeList):
		var s sql.NullFloat64
		colVar[i] = &s
	case Contains(dt, UintTypeList):
		var s []uint8
		colVar[i] = &s
	case Contains(dt, StringTypeList):
		var s sql.NullString
		colVar[i] = &s
	default:
		var s interface{}
		colVar[i] = &s
	}
}

func GetResultValue(colVar interface{}, typeName string) interface{} {
	dt := DT(typeName)
	switch {
	case Contains(dt, BoolTypeList):
		temp := colVar.(*sql.NullBool)
		if temp.Valid { return temp.Bool }
	case Contains(dt, IntTypeList):
		temp := colVar.(*sql.NullInt64)
		if temp.Valid { return temp.Int64 }
	case Contains(dt, FloatTypeList):
		temp := colVar.(*sql.NullFloat64)
		if temp.Valid { return temp.Float64 }
	case Contains(dt, UintTypeList):
		return *(colVar.(*[]uint8))
	case Contains(dt, StringTypeList):
		temp := colVar.(*sql.NullString)
		if temp.Valid { return utils.StrIsoDateToDateTime(temp.String) }
	default:
		if colVar2, ok := colVar.(*interface{}); ok {
			switch v := (*colVar2).(type) {
			case int64  : return v
			case string : return v
			case float64: return v
			case []uint8: return v
			case bool   : return v
			}
		}
	}
	return nil
}

// SetResultValue set the result value.
func SetResultValue(result map[string]interface{}, index string, colVar interface{}, typeName string) {
	dt := DT(typeName)
	switch {
	case Contains(dt, BoolTypeList):
		temp := colVar.(*sql.NullBool)
		if temp.Valid {
			result[index] = temp.Bool
		} else {
			result[index] = nil
		}
	case Contains(dt, IntTypeList):
		temp := colVar.(*sql.NullInt64)
		if temp.Valid {
			result[index] = temp.Int64
		} else {
			result[index] = nil
		}
	case Contains(dt, FloatTypeList):
		temp := colVar.(*sql.NullFloat64)
		if temp.Valid {
			result[index] = temp.Float64
		} else {
			result[index] = nil
		}
	case Contains(dt, UintTypeList):
		result[index] = *(colVar.(*[]uint8))
	case Contains(dt, StringTypeList):
		temp := colVar.(*sql.NullString)
		if temp.Valid {
			result[index] = utils.StrIsoDateToDateTime(temp.String)
		} else {
			result[index] = nil
		}
	default:
		if colVar2, ok := colVar.(*interface{}); ok {
			switch colVar := (*colVar2).(type) {
			case int64  : result[index] = colVar
			case string : result[index] = colVar
			case float64: result[index] = colVar
			case []uint8: result[index] = colVar
			case bool   : result[index] = colVar
			default     : result[index] = nil
			}
		}
	}
}
