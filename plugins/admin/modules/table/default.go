package table

import (
	"errors"
	"fmt"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
	errs "github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/paginator"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/template/types"
	"html/template"
	"io"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
)

// DefaultTable is an implementation of table.Table
type DefaultTable struct {
	*BaseTable
	connectionDriver     string
	connectionDriverMode string
	connection           string
	sourceURL            string
	getDataFun           GetDataFun
	dbObj                db.Connection
}

type GetDataFun func(params parameter.Parameters) ([]map[string]interface{}, int)

func NewDefaultTable(cfgs ...Config) Table {
	var cfg Config

	if len(cfgs) > 0 && cfgs[0].PrimaryKey.Name != "" {
		cfg = cfgs[0]
	} else {
		cfg = DefaultConfig()
	}

	return &DefaultTable{
		BaseTable: &BaseTable{
			Info:           types.NewInfoPanel(cfg.PrimaryKey.Name),
			Form:           types.NewFormPanel(),
			NewForm:        types.NewFormPanel(),
			Detail:         types.NewInfoPanel(cfg.PrimaryKey.Name),
			CanAdd:         cfg.CanAdd,
			Editable:       cfg.Editable,
			Deletable:      cfg.Deletable,
			Exportable:     cfg.Exportable,
			PrimaryKey:     cfg.PrimaryKey,
			OnlyNewForm:    cfg.OnlyNewForm,
			OnlyUpdateForm: cfg.OnlyUpdateForm,
			OnlyDetail:     cfg.OnlyDetail,
			OnlyInfo:       cfg.OnlyInfo,
		},
		connectionDriver:     cfg.Driver,
		connectionDriverMode: cfg.DriverMode,
		connection:           cfg.Connection,
		sourceURL:            cfg.SourceURL,
		getDataFun:           cfg.GetDataFun,
	}
}

// Copy copy a new table.Table from origin DefaultTable
func (tb *DefaultTable) Copy() Table {
	return &DefaultTable{
		BaseTable: &BaseTable{
			Form: types.NewFormPanel().SetTable(tb.Form.Table).
				SetDescription(tb.Form.Description).
				SetTitle(tb.Form.Title),
			NewForm: types.NewFormPanel().SetTable(tb.Form.Table).
				SetDescription(tb.Form.Description).
				SetTitle(tb.Form.Title),
			Info: types.NewInfoPanel(tb.PrimaryKey.Name).SetTable(tb.Info.Table).
				SetDescription(tb.Info.Description).
				SetTitle(tb.Info.Title).
				SetGetDataFn(tb.Info.GetDataFn),
			Detail: types.NewInfoPanel(tb.PrimaryKey.Name).SetTable(tb.Detail.Table).
				SetDescription(tb.Detail.Description).
				SetTitle(tb.Detail.Title).
				SetGetDataFn(tb.Detail.GetDataFn),
			CanAdd:     tb.CanAdd,
			Editable:   tb.Editable,
			Deletable:  tb.Deletable,
			Exportable: tb.Exportable,
			PrimaryKey: tb.PrimaryKey,
		},
		connectionDriver:     tb.connectionDriver,
		connectionDriverMode: tb.connectionDriverMode,
		connection:           tb.connection,
		sourceURL:            tb.sourceURL,
		getDataFun:           tb.getDataFun,
	}
}

// GetData query the data set.
func (tb *DefaultTable) GetData(params parameter.Parameters) (PanelInfo, error) {
	var (
		data      []map[string]interface{}
		size      int
		benchmark = utils.StartBenchmark()
	)

	if tb.Info.UpdateParametersFns != nil {
		for _, fn := range tb.Info.UpdateParametersFns {
			fn(&params)
		}
	}

	if tb.Info.QueryFilterFn != nil {
		var ids []string
		var stopQuery bool

		if tb.getDataFun == nil && tb.Info.GetDataFn == nil {
			ids, stopQuery = tb.Info.QueryFilterFn(params, tb.db())
		} else {
			ids, stopQuery = tb.Info.QueryFilterFn(params, nil)
		}

		if stopQuery {
			return tb.GetDataWithIds(params.WithPKs(ids...))
		}
	}

	if tb.getDataFun != nil {
		data, size = tb.getDataFun(params)
	} else if tb.sourceURL != "" {
		data, size = tb.getDataFromURL(params)
	} else if tb.Info.GetDataFn != nil {
		data, size = tb.Info.GetDataFn(params)
	} else if params.IsAll() {
		return tb.getAllDataFromDatabase(params)
	} else {
		return tb.getDataFromDatabase(params)
	}

	infoList := make(types.InfoList, len(data))
	for i, d := range data {
		infoList[i] = tb.getTempModelData(d, params, nil)
	}

	thead, _, _, _, _, filterForm := tb.getTheadAndFilterForm(params, nil)

	extraInfo := ""
	if !tb.Info.IsHideQueryInfo {
		extraInfo = elapsedQueryTime(benchmark)
	}

	return PanelInfo{
		Thead:    thead,
		InfoList: infoList,
		Paginator: paginator.Get(paginator.Config{
			Size:         size,
			Param:        params,
			PageSizeList: tb.Info.GetPageSizeList(),
		}).SetExtraInfo(template.HTML(extraInfo)),
		Title:          tb.Info.Title,
		FilterFormData: filterForm,
		Description:    tb.Info.Description,
	}, nil
}

type GetDataFromURLRes struct {
	Data []map[string]interface{}
	Size int
}

func (tb *DefaultTable) getDataFromURL(params parameter.Parameters) ([]map[string]interface{}, int) {
	u := ""
	if strings.ContainsRune(tb.sourceURL, '?') {
		u = utils.StrConcat(tb.sourceURL, "&", params.Join())
	} else {
		u = utils.StrConcat(tb.sourceURL, "?", params.Join())
	}

	res, err := http.Get(utils.StrConcat(u, "&pk=", strings.Join(params.PKs(), ",")))
	if err != nil {
		return []map[string]interface{}{}, 0
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return []map[string]interface{}{}, 0
	}

	var data GetDataFromURLRes

	err = utils.JsonUnmarshal(body, &data)
	if err != nil {
		return []map[string]interface{}{}, 0
	}

	return data.Data, data.Size
}

// GetDataWithIds query the data set.
func (tb *DefaultTable) GetDataWithIds(params parameter.Parameters) (PanelInfo, error) {
	var (
		data      []map[string]interface{}
		size      int
		benchmark = utils.StartBenchmark()
	)

	if tb.getDataFun != nil {
		data, size = tb.getDataFun(params)
	} else if tb.sourceURL != "" {
		data, size = tb.getDataFromURL(params)
	} else if tb.Info.GetDataFn != nil {
		data, size = tb.Info.GetDataFn(params)
	} else {
		return tb.getDataFromDatabase(params)
	}

	infoList := make([]map[string]types.InfoItem, len(data))
	for i, d := range data {
		infoList[i] = tb.getTempModelData(d, params, nil)
	}

	thead, _, _, _, _, filterForm := tb.getTheadAndFilterForm(params, nil)

	return PanelInfo{
		Thead:    thead,
		InfoList: infoList,
		Paginator: paginator.Get(paginator.Config{
			Size:         size,
			Param:        params,
			PageSizeList: tb.Info.GetPageSizeList(),
		}).SetExtraInfo(template.HTML(elapsedQueryTime(benchmark))),
		Title:          tb.Info.Title,
		FilterFormData: filterForm,
		Description:    tb.Info.Description,
	}, nil
}

func (tb *DefaultTable) getTempModelData(res map[string]interface{}, params parameter.Parameters, columnMap map[string]struct{}) map[string]types.InfoItem {
	tempModelData := map[string]types.InfoItem{
		"__goadmin_edit_params"  : {},
		"__goadmin_delete_params": {},
		"__goadmin_detail_params": {},
	}
	var typeName db.DatabaseType
	headField    := ""
	editParams   := ""
	deleteParams := ""
	detailParams := ""
	noColumns    := len(columnMap) == 0

	primaryKeyValue := db.GetValueFromDatabaseType(tb.PrimaryKey.Type, res[tb.PrimaryKey.Name], noColumns)

	for _, field := range tb.Info.FieldList {
		if field.Hide { continue }

		validJoin := field.Joins.Valid()
		if validJoin {
			headField = utils.StrConcat(field.Joins.Last().GetTableName(), parameter.FilterParamJoinInfix, field.Field)
		} else {
			headField = field.Field
		}

		if !modules.InArrayWithoutEmpty(params.Columns, headField) {
			continue
		}

		if validJoin {
			typeName = db.Varchar
		} else {
			typeName = field.TypeName
		}

		combinedValue := db.GetValueFromDatabaseType(typeName, res[headField], noColumns).String()

		fieldModel := types.FieldModel{
			ID : primaryKeyValue.String(),
			Row: res,
		}
		if noColumns || validJoin || utils.InMapT(columnMap, headField) {
			fieldModel.Value = combinedValue
		}

		valueStr := ""
		switch t := field.ToDisplay(fieldModel).(type) {
		case string       : valueStr = t
		case template.HTML: valueStr = string(t)
		}

		tempModelData[headField] = types.InfoItem{
			Content: template.HTML(valueStr),
			Value  : combinedValue,
		}

		if field.IsEditParam {
			editParams = utils.StrConcat(editParams, "&__goadmin_edit_", field.Field, "=", valueStr)
			//editParams += "__goadmin_edit_" + field.Field + "=" + valueStr + "&"
		}
		if field.IsDeleteParam {
			deleteParams = utils.StrConcat(deleteParams, "&__goadmin_delete_", field.Field, "=", valueStr)
			//deleteParams += "__goadmin_delete_" + field.Field + "=" + valueStr + "&"
		}
		if field.IsDetailParam {
			detailParams = utils.StrConcat(detailParams, "&__goadmin_detail_", field.Field, "=", valueStr)
			//detailParams += "__goadmin_detail_" + field.Field + "=" + valueStr + "&"
		}
	}

	if editParams != "" {
		tempModelData["__goadmin_edit_params"] = types.InfoItem{ Content: template.HTML(editParams) }
		//tempModelData["__goadmin_edit_params"] = types.InfoItem{Content: template.HTML("&" + editParams[:len(editParams)-1])}
	}
	if deleteParams != "" {
		tempModelData["__goadmin_delete_params"] = types.InfoItem{ Content: template.HTML(deleteParams) }
		//tempModelData["__goadmin_delete_params"] = types.InfoItem{Content: template.HTML("&" + deleteParams[:len(deleteParams)-1])}
	}
	if detailParams != "" {
		tempModelData["__goadmin_detail_params"] = types.InfoItem{ Content: template.HTML(detailParams) }
		//tempModelData["__goadmin_detail_params"] = types.InfoItem{Content: template.HTML("&" + detailParams[:len(detailParams)-1])}
	}

	primaryKeyField := tb.Info.FieldList.GetFieldByFieldName(tb.PrimaryKey.Name)
	value := primaryKeyField.ToDisplay(types.FieldModel{
		ID:    primaryKeyValue.String(),
		Value: primaryKeyValue.String(),
		Row:   res,
	})

	var valueHtml template.HTML
	switch t := value.(type) {
	case string       : valueHtml = template.HTML(t)
	case template.HTML: valueHtml = t
	}

	tempModelData[tb.PrimaryKey.Name] = types.InfoItem{
		Content: valueHtml,
		Value  : primaryKeyValue.String(),
	}

	return tempModelData
}

func (tb *DefaultTable) getAllDataFromDatabase(params parameter.Parameters) (PanelInfo, error) {
	conn   := tb.db()
	delim  := conn.GetDelimiter()
	delim2 := conn.GetDelimiter2()
	dl     := len(delim) + len(delim2)
	pkl    := dl + len(tb.PrimaryKey.Name)

	const q1 = "SELECT %s FROM %s %s %s %s ORDER BY "
	const q2 = "%s"
	const q3 = " %s"
	var queryStmt strings.Builder
	queryStmt.Grow(len(q1) + len(q2) * 2 + 1 + len(q3) + dl)
	queryStmt.WriteString(q1)
	queryStmt.WriteString(delim)
	queryStmt.WriteString(q2)
	queryStmt.WriteString(delim2)
	queryStmt.WriteByte('.')
	queryStmt.WriteString(delim)
	queryStmt.WriteString(q2)
	queryStmt.WriteString(delim2)
	queryStmt.WriteString(q3)

	columnMap, _ := tb.getColumnMap(tb.Info.Table)

	thead, fields, joins := tb.Info.FieldList.GetThead(types.TableInfo{
		Table:      tb.Info.Table,
		Delimiter:  delim,
		Delimiter2: delim2,
		Driver:     tb.connectionDriver,
		PrimaryKey: tb.PrimaryKey.Name,
	}, params, columnMap)

	{
		var sb strings.Builder
		sb.Grow(len(fields) + len(tb.Info.Table) + 1 + pkl)
		sb.WriteString(fields)
		sb.WriteString(tb.Info.Table)
		sb.WriteByte('.')
		sb.WriteString(delim)
		sb.WriteString(tb.PrimaryKey.Name)
		sb.WriteString(delim2)
		fields = sb.String()
	}

	var groupBy strings.Builder
	if joins != "" {
		const g = "GROUP BY "
		groupBy.Grow(len(g) + 1 + len(tb.Info.Table) + pkl)
		groupBy.WriteString(g)
		groupBy.WriteString(tb.Info.Table)
		groupBy.WriteByte('.')
		groupBy.WriteString(delim)
		groupBy.WriteString(tb.PrimaryKey.Name)
		groupBy.WriteString(delim2)
	}

	wheres, whereArgs, existKeys := params.Statement("", tb.Info.Table, delim, delim2,
		nil, columnMap, nil, tb.Info.FieldList.GetFieldFilterProcessValue)
	wheres, whereArgs = tb.Info.Wheres.Statement(wheres, delim, delim2, whereArgs, existKeys, columnMap)
	wheres, whereArgs = tb.Info.WhereRaws.Statement(wheres, whereArgs)

	if wheres != "" {
		wheres = "WHERE " + wheres
	}

	if _, ok := columnMap[params.SortField]; !ok {
		params.SortField = tb.PrimaryKey.Name
	}

	queryCmd := fmt.Sprintf(queryStmt.String(), fields, tb.Info.Table, joins,
		wheres, groupBy.String(), tb.Info.Table, params.SortField, params.SortType)
	logger.LogSQL(queryCmd, whereArgs)

	res, err := conn.QueryWithConnection(tb.connection, queryCmd, whereArgs...)

	if err != nil {
		return PanelInfo{}, err
	}

	infoList := make([]map[string]types.InfoItem, len(res))
	for i, e := range res {
		infoList[i] = tb.getTempModelData(e, params, columnMap)
	}

	return PanelInfo{
		InfoList:    infoList,
		Thead:       thead,
		Title:       tb.Info.Title,
		Description: tb.Info.Description,
	}, nil
}

// TODO: refactor
func (tb *DefaultTable) getDataFromDatabase(params parameter.Parameters) (PanelInfo, error) {
	var (
		conn        = tb.db()
		delim       = conn.GetDelimiter()
		delim2      = conn.GetDelimiter2()
		placeholder = modules.Delimiter(delim, delim2, "%s")
		queryStmt   string
		countStmt   string
		ids         = params.PKs()
		table       = modules.Delimiter(delim, delim2, tb.Info.Table)
		pk          = utils.StrConcat(table, ".", modules.Delimiter(delim, delim2, tb.PrimaryKey.Name))
		isMssql     = conn.Name() == db.DriverMssql
	)

	benchmark := utils.StartBenchmark()

	if len(ids) > 0 {
		countExtra := ""
		if isMssql { countExtra = " AS [size]" }
		// %s means: fields, table, join table, pk values, group by, order by field,  order by type
		queryStmt = utils.StrConcat("SELECT %s FROM ", placeholder, " %s WHERE ", pk, " IN (%s) %s ORDER BY %s.", placeholder, " %s")
		// %s means: table, join table, pk values
		countStmt = utils.StrConcat("SELECT count(*)", countExtra, " FROM ", placeholder, " %s WHERE ", pk, " IN (%s)")
	} else {
		if isMssql {
			// %s means: order by field, order by type, fields, table, join table, wheres, group by
			queryStmt = utils.StrConcat("SELECT * FROM (SELECT ROW_NUMBER() OVER (ORDER BY %s.", placeholder, " %s) AS ROWNUMBER_, %s FROM ",
				placeholder, "%s %s %s ) as TMP_ WHERE TMP_.ROWNUMBER_ > ? AND TMP_.ROWNUMBER_ <= ?")
			// %s means: table, join table, wheres
			countStmt = utils.StrConcat("SELECT count(*) AS [size] FROM (SELECT count(*) AS [size] FROM ", placeholder, " %s %s %s) src")
		} else {
			// %s means: fields, table, join table, wheres, group by, order by field, order by type
			queryStmt = utils.StrConcat("SELECT %s FROM ", placeholder, "%s %s %s ORDER BY ", placeholder, ".", placeholder, " %s LIMIT ? OFFSET ?")
			// %s means: table, join table, wheres
			countStmt = utils.StrConcat("SELECT count(*) FROM (SELECT ", pk, " FROM ", placeholder, " %s %s %s) src")
		}
	}

	columnMap, _ := tb.getColumnMap(tb.Info.Table)
	thead, fields, joinFields, joins, joinTableMap, filterForm := tb.getTheadAndFilterForm(params, columnMap)

	fields      += pk
	allFields   := fields
	groupFields := fields

	if joinFields != "" {
		allFields = utils.StrConcat(allFields, ",", joinFields[:len(joinFields)-1])
		if isMssql {
			for _, field := range tb.Info.FieldList {
				if field.TypeName == db.Text || field.TypeName == db.Longtext {
					f := modules.Delimiter(delim, delim2, field.Field)
					headField  := utils.StrConcat(table, ".", f)
					casting    := utils.StrConcat("CAST(", headField, " AS NVARCHAR(MAX))")
					allFields   = strings.ReplaceAll(allFields, headField, utils.StrConcat(casting, " AS ", f))
					groupFields = strings.ReplaceAll(groupFields, headField, casting)
				}
			}
		}
	}

	if _, ok := columnMap[params.SortField]; !ok {
		params.SortField = tb.PrimaryKey.Name
	}

	var (
		wheres    = ""
		whereArgs []interface{}
		args      []interface{}
		existKeys map[string]struct{}
	)

	if len(ids) > 0 {
		var sb strings.Builder
		sb.Grow(64)
		for _, value := range ids {
			if value != "" {
				if sb.Len() == 0 {
					sb.WriteByte('?')
				} else {
					sb.WriteString(",?")
				}
				//wheres += "?,"
				args = append(args, value)
			}
		}
		wheres = sb.String()
		//wheres = wheres[:len(wheres)-1]
	} else {
		// parameter
		wheres, whereArgs, existKeys = params.Statement(wheres, tb.Info.Table, conn.GetDelimiter(), conn.GetDelimiter2(),
			whereArgs, columnMap, existKeys, tb.Info.FieldList.GetFieldFilterProcessValue)
		// pre query
		wheres, whereArgs = tb.Info.Wheres.Statement(wheres, conn.GetDelimiter(), conn.GetDelimiter2(), whereArgs, existKeys, columnMap)
		wheres, whereArgs = tb.Info.WhereRaws.Statement(wheres, whereArgs)

		if wheres != "" {
			wheres = "WHERE " + wheres
		}

		if isMssql {
			args = append(whereArgs, (params.PageInt - 1) * params.PageSizeInt, params.PageInt * params.PageSizeInt)
		} else {
			args = append(whereArgs, params.PageSizeInt, (params.PageInt - 1) * params.PageSizeInt)
		}
	}

	groupBy := ""
	if len(joinTableMap) > 0 {
		var sb strings.Builder
		sb.Grow(64)
		sb.WriteString("GROUP BY ")
		if isMssql {
			sb.WriteString(groupFields)
			//groupBy = " GROUP BY " + groupFields
		} else {
			sb.WriteString(pk)
			//groupBy = " GROUP BY " + pk
		}
		groupBy = sb.String()
	}

	queryCmd := ""
	if isMssql && len(ids) == 0 {
		queryCmd = fmt.Sprintf(queryStmt, tb.Info.Table, params.SortField, params.SortType,
			allFields, tb.Info.Table, joins, wheres, groupBy)
	} else {
		queryCmd = fmt.Sprintf(queryStmt, allFields, tb.Info.Table, joins, wheres, groupBy,
			tb.Info.Table, params.SortField, params.SortType)
	}

	logger.LogSQL(queryCmd, args)
	res, err := conn.QueryWithConnection(tb.connection, queryCmd, args...)

	if err != nil {
		return PanelInfo{}, err
	}

	infoList := make([]map[string]types.InfoItem, len(res))
	for i, e := range res {
		infoList[i] = tb.getTempModelData(e, params, columnMap)
	}

	// TODO: use the dialect
	var size int

	if len(ids) == 0 {
		countCmd := fmt.Sprintf(countStmt, tb.Info.Table, joins, wheres, groupBy)

		total, err := conn.QueryWithConnection(tb.connection, countCmd, whereArgs...)
		if err != nil { return PanelInfo{}, err }

		logger.LogSQL(countCmd, nil)

		switch tb.connectionDriver {
		case db.DriverPostgresql:
			if tb.connectionDriverMode == "h2" || config.GetDatabases().GetDefault().DriverMode == "h2" {
				size = int(total[0]["count(*)"].(int64))
			} else {
				size = int(total[0]["count"].(int64))
			}
		case db.DriverMssql:
			size = int(total[0]["size"].(int64))
		default:
			size = int(total[0]["count(*)"].(int64))
		}
		/*if tb.connectionDriver == db.DriverPostgresql {
			if tb.connectionDriverMode == "h2" {
				size = int(total[0]["count(*)"].(int64))
			} else if config.GetDatabases().GetDefault().DriverMode == "h2" {
				size = int(total[0]["count(*)"].(int64))
			} else {
				size = int(total[0]["count"].(int64))
			}
		} else if tb.connectionDriver == db.DriverMssql {
			size = int(total[0]["size"].(int64))
		} else {
			size = int(total[0]["count(*)"].(int64))
		}*/
	}

	return PanelInfo{
		Thead:          thead,
		InfoList:       infoList,
		Paginator:      tb.GetPaginator(size, params, template.HTML(elapsedQueryTime(benchmark))),
		Title:          tb.Info.Title,
		FilterFormData: filterForm,
		Description:    tb.Info.Description,
	}, nil
}

func elapsedQueryTime(benchmark utils.Benchmark) string {
	elapsed := benchmark.ElapsedMillis()
	var sb strings.Builder
	sb.Grow(32)
	sb.WriteString("<b>")
	sb.WriteString(language.Get("query time"))
	sb.WriteString("</b>: ")
	_, _ = fmt.Fprintf(&sb, "%.3fms", elapsed)
	return sb.String()
}

func getDataRes(list []map[string]interface{}, _ int) map[string]interface{} {
	if len(list) > 0 { return list[0] }
	return nil
}

// GetDataWithId query the single row of data.
func (tb *DefaultTable) GetDataWithId(param parameter.Parameters) (FormInfo, error) {
	var (
		res       map[string]interface{}
		columnMap map[string]struct{}
		id        = param.PK()
	)

	if tb.getDataFun != nil {
		res = getDataRes(tb.getDataFun(param))
	} else if tb.sourceURL != "" {
		res = getDataRes(tb.getDataFromURL(param))
	} else if tb.Detail.GetDataFn != nil {
		res = getDataRes(tb.Detail.GetDataFn(param))
	} else if tb.Info.GetDataFn != nil {
		res = getDataRes(tb.Info.GetDataFn(param))
	} else {
		columnMap, _ = tb.getColumnMap(tb.Form.Table)
		var (
			queryStmt, fields, joins, joinFields, groupBy strings.Builder
			pk         string
			err        error
			joinTabMap map[string]struct{}
			args       = []interface{}{ id }
			conn       = tb.db()
			delim      = conn.GetDelimiter()
			delim2     = conn.GetDelimiter2()
			tableName  = modules.Delimiter(delim, delim2, tb.GetForm().Table)
		)
		{
			var sb strings.Builder
			sb.Grow(64)
			sb.WriteString(tableName)
			sb.WriteByte('.')
			sb.WriteString(delim)
			sb.WriteString(tb.PrimaryKey.Name)
			sb.WriteString(delim2)
			pk = sb.String()
			//pk = tableName + "." + modules.Delimiter(delim, delim2, tb.PrimaryKey.Name)
		}
		queryStmt.Grow(512)
		queryStmt.WriteString("SELECT %s FROM %s %s WHERE ")
		queryStmt.WriteString(pk)
		queryStmt.WriteString(" = ? %s")
		//queryStmt = "SELECT %s FROM %s %s WHERE " + pk + " = ? %s "

		fields.Grow(256)

		for _, formField := range tb.Form.FieldList {
			validJoin := formField.Joins.Valid()

			if formField.Field != pk && utils.InMapT(columnMap, formField.Field) && !validJoin {
				if fields.Len() > 0 { fields.WriteByte(',') }
				fields.WriteString(tableName)
				fields.WriteByte('.')
				fields.WriteString(delim)
				fields.WriteString(formField.Field)
				fields.WriteString(delim2)
				//fields += tableName + "." + modules.FilterField(formField.Field, delim, delim2) + ","
			}

			if validJoin {
				var sbField, sbHeadField strings.Builder
				sbField.Grow(128)
				sbField.WriteString(formField.Joins.Last().GetTableName(delim, delim2))
				sbField.WriteByte('.')
				sbField.WriteString(delim)
				sbField.WriteString(formField.Field)
				sbField.WriteString(delim2)
				sbHeadField.Grow(64)
				sbHeadField.WriteString(formField.Joins.Last().GetTableName())
				sbHeadField.WriteString(parameter.FilterParamJoinInfix)
				sbHeadField.WriteString(formField.Field)
				if joinFields.Cap() == 0 {
					joinFields.Grow(256)
					joins.Grow(256)
				}
				joinFields.WriteByte(',')
				joinFields.WriteString(db.GetAggregationExpression(conn.Name(), sbField.String(), sbHeadField.String(), types.JoinFieldValueDelimiter))
				//headField := formField.Joins.Last().GetTableName() + parameter.FilterParamJoinInfix + formField.Field
				//joinFields += db.GetAggregationExpression(conn.Name(), formField.Joins.Last().GetTableName(delim, delim2) + "." +
				//	modules.FilterField(formField.Field, delim, delim2), headField, types.JoinFieldValueDelimiter) + ","
				for _, join := range formField.Joins {
					joinTableName := join.GetTableName(delim, delim2)
					if _, ok := joinTabMap[joinTableName]; !ok {
						if joinTabMap == nil {
							joinTabMap = map[string]struct{}{ joinTableName: {} }
						} else {
							joinTabMap[joinTableName] = struct{}{}
						}
						if join.BaseTable == "" {
							join.BaseTable = tableName
						}
						joins.WriteString(" LEFT JOIN ")
						joins.WriteString(delim)
						joins.WriteString(join.Table)
						joins.WriteString(delim2)
						if join.TableAlias != "" {
							joins.WriteByte(' ')
							joins.WriteString(join.TableAlias)
						}
						joins.WriteString(" ON ")
						joins.WriteString(joinTableName)
						joins.WriteByte('.')
						joins.WriteString(delim)
						joins.WriteString(join.JoinField)
						joins.WriteString(delim2)
						joins.WriteByte('=')
						joins.WriteString(join.BaseTable)
						joins.WriteByte('.')
						joins.WriteString(delim)
						joins.WriteString(join.Field)
						joins.WriteString(delim2)
						//joins += " LEFT JOIN " + modules.FilterField(join.Table, delim, delim2) + " " + join.TableAlias + " ON " +
						//	joinTableName + "." + modules.FilterField(join.JoinField, delim, delim2) + " = " +
						//	join.BaseTable + "." + modules.FilterField(join.Field, delim, delim2)
					}
				}
			}
		}

		if fields.Len() > 0 { fields.WriteByte(',') }
		fields.WriteString(pk)
		//fields += pk

		isMssql := conn.Name() == db.DriverMssql
		useGroupFields := isMssql && len(joinTabMap) > 0
		var groupFields string
		if useGroupFields {
			groupFields = fields.String()
		}

		//if joinFields != "" {
		if joinFields.Len() > 0 {
			fields.WriteString(joinFields.String())
			//fields += joinFields.String()
			//fields += "," + joinFields[:len(joinFields)-1]
			if isMssql {
				strFields := fields.String()
				for _, formField := range tb.Form.FieldList {
					if formField.TypeName == db.Text || formField.TypeName == db.Longtext {
						f := modules.Delimiter(delim, delim2, formField.Field)
						headField := utils.StrConcat(tb.Info.Table, ".", f)
						strFields = strings.ReplaceAll(strFields, headField, utils.StrConcat("CAST(", headField, " AS NVARCHAR(MAX)) as ", f))
						if useGroupFields {
							groupFields = strings.ReplaceAll(groupFields, headField, utils.StrConcat("CAST(", headField, " AS NVARCHAR(MAX))"))
						}
					}
				}
				fields.Reset()
				fields.Grow(len(strFields))
				fields.WriteString(strFields)
			}
		}

		if len(joinTabMap) > 0 {
			groupBy.Grow(64)
			groupBy.WriteString(" GROUP BY ")
			if useGroupFields {
				groupBy.WriteString(groupFields)
			} else {
				groupBy.WriteString(pk)
			}
		}

		queryCmd := fmt.Sprintf(queryStmt.String(), fields.String(), tableName, joins.String(), groupBy.String())
		logger.LogSQL(queryCmd, args)
		result, err := conn.QueryWithConnection(tb.connection, queryCmd, args...)
		if err != nil {
			return FormInfo{ Title: tb.Form.Title, Description: tb.Form.Description }, err
		}

		if len(result) == 0 {
			return FormInfo{ Title: tb.Form.Title, Description: tb.Form.Description }, errors.New(errs.WrongID)
		}

		res = result[0]
	}

	var (
		groupFormList []types.FormFields
		groupHeaders  []string
	)

	if len(tb.Form.TabGroups) > 0 {
		groupFormList, groupHeaders = tb.Form.GroupFieldWithValue(tb.PrimaryKey.Name, id, columnMap, res, tb.sqlObjOrNil)
		return FormInfo{
			FieldList:         tb.Form.FieldList,
			GroupFieldList:    groupFormList,
			GroupFieldHeaders: groupHeaders,
			Title:             tb.Form.Title,
			Description:       tb.Form.Description,
		}, nil
	}

	return FormInfo{
		FieldList:         tb.Form.FieldsWithValue(tb.PrimaryKey.Name, id, columnMap, res, tb.sqlObjOrNil),
		GroupFieldList:    groupFormList,
		GroupFieldHeaders: groupHeaders,
		Title:             tb.Form.Title,
		Description:       tb.Form.Description,
	}, nil
}

// UpdateData update data.
func (tb *DefaultTable) UpdateData(dataList form.Values) error {
	dataList.Add(form.PostTypeKey, "0")

	var (
		errMsg = ""
		err    error
	)

	if tb.Form.PostHook != nil {
		defer func() {
			dataList.Add(form.PostTypeKey, "0")
			dataList.Add(form.PostResultKey, errMsg)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error(r)
						logger.Error(string(debug.Stack()))
					}
				}()

				if err := tb.Form.PostHook(dataList); err != nil {
					logger.Error(err)
				}
			}()
		}()
	}

	if tb.Form.Validator != nil {
		if err := tb.Form.Validator(dataList); err != nil {
			errMsg = "post error: " + err.Error()
			return err
		}
	}

	if tb.Form.PreProcessFn != nil {
		dataList = tb.Form.PreProcessFn(dataList)
	}

	if tb.Form.UpdateFn != nil {
		dataList.Delete(form.PostTypeKey)
		err = tb.Form.UpdateFn(tb.PreProcessValue(dataList, types.PostTypeUpdate))
		if err != nil {
			errMsg = "post error: " + err.Error()
		}
		return err
	}

	if len(dataList) == 0 {
		return nil
	}

	_, err = tb.sql().Table(tb.Form.Table).
		Where(tb.PrimaryKey.Name, "=", dataList.Get(tb.PrimaryKey.Name)).
		Update(tb.getInjectValueFromFormValue(dataList, types.PostTypeUpdate))

	// NOTE: some errors should be ignored.
	if db.CheckError(err, db.UPDATE) {
		if err != nil {
			errMsg = "post error: " + err.Error()
		}
		return err
	}

	return nil
}

// InsertData insert data.
func (tb *DefaultTable) InsertData(dataList form.Values) error {
	dataList.Add(form.PostTypeKey, "1")

	var (
		id     = int64(0)
		err    error
		errMsg = ""
		f      = tb.GetActualNewForm()
	)

	if f.PostHook != nil {
		defer func() {
			dataList.Add(form.PostTypeKey, "1")
			dataList.Add(tb.GetPrimaryKey().Name, strconv.Itoa(int(id)))
			dataList.Add(form.PostResultKey, errMsg)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error(r)
						logger.Error(string(debug.Stack()))
					}
				}()
				if err := f.PostHook(dataList); err != nil {
					logger.Error(err)
				}
			}()
		}()
	}

	if f.Validator != nil {
		if err = f.Validator(dataList); err != nil {
			errMsg = "post error: " + err.Error()
			return err
		}
	}

	if f.PreProcessFn != nil {
		dataList = f.PreProcessFn(dataList)
	}

	if f.InsertFn != nil {
		dataList.Delete(form.PostTypeKey)
		err = f.InsertFn(tb.PreProcessValue(dataList, types.PostTypeCreate))
		if err != nil {
			errMsg = "post error: " + err.Error()
		}
		return err
	}

	if len(dataList) == 0 {
		return nil
	}

	id, err = tb.sql().Table(f.Table).Insert(tb.getInjectValueFromFormValue(dataList, types.PostTypeCreate))

	// NOTE: some errors should be ignored.
	if db.CheckError(err, db.INSERT) {
		errMsg = "post error: " + err.Error()
		return err
	}

	return nil
}

func (tb *DefaultTable) getInjectValueFromFormValue(dataList form.Values, typ types.PostType) dialect.H {
	var (
		exceptString  map[string]struct{}
		value         = make(dialect.H)
		columnMap, auto = tb.getColumnMap(tb.Form.Table)
		fun           types.PostFieldFilterFn
	)

	// If a key is auto increment primary key, it cannot be inserted nor updated.
	if auto {
		exceptString = map[string]struct{}{
			tb.PrimaryKey.Name: {}, form.PreviousKey: {}, form.MethodKey: {}, form.TokenKey: {}, constant.IframeKey: {}, constant.IframeIDKey: {},
		}
	} else {
		exceptString = utils.DefaultExceptMap
	}

	if !dataList.IsSingleUpdatePost() {
		for _, field := range tb.Form.FieldList {
			if field.FormType.IsMultiSelect() {
				key := field.Field + "[]"
				if _, ok := dataList[key]; !ok {
					dataList[key] = []string{ "" }
				}
			}
		}
	}

	dataList.RemoveRemark()

	for k, v := range dataList {
		k = strings.ReplaceAll(k, "[]", "")
		if _, ok := exceptString[k]; !ok {
			if _, ok := columnMap[k]; ok {
				field := tb.Form.FieldList.FindByFieldName(k)
				delim := ","
				if field != nil {
					fun   = field.PostFilterFn
					delim = modules.SetDefault(field.DefaultOptionDelimiter, ",")
				}
				vv := modules.RemoveBlankFromArray(v)
				if fun != nil {
					value[k] = fun(types.PostFieldModel{
						ID:       dataList.Get(tb.PrimaryKey.Name),
						Value:    vv,
						Row:      dataList.ToMap(),
						PostType: typ,
					})
				} else {
					switch len(vv) {
					case 0 : value[k] = ""
					case 1 : value[k] = vv[0]
					default: value[k] = strings.Join(vv, delim)
					}
				}
			} else {
				field := tb.Form.FieldList.FindByFieldName(k)
				if field != nil && field.PostFilterFn != nil {
					field.PostFilterFn(types.PostFieldModel{
						ID:       dataList.Get(tb.PrimaryKey.Name),
						Value:    modules.RemoveBlankFromArray(v),
						Row:      dataList.ToMap(),
						PostType: typ,
					})
				}
			}
		}
	}
	return value
}

func (tb *DefaultTable) PreProcessValue(dataList form.Values, typ types.PostType) form.Values {
	dataList = dataList.RemoveRemark()
	var fun types.PostFieldFilterFn
	for k, v := range dataList {
		k = strings.ReplaceAll(k, "[]", "")
		if _, ok := utils.DefaultExceptMap[k]; !ok {
			field := tb.Form.FieldList.FindByFieldName(k)
			if field != nil {
				fun = field.PostFilterFn
			}
			if fun != nil {
				dataList.Add(k, fmt.Sprintf("%s", fun(types.PostFieldModel{
					ID:       dataList.Get(tb.PrimaryKey.Name),
					Value:    modules.RemoveBlankFromArray(v),
					Row:      dataList.ToMap(),
					PostType: typ,
				})))
			}
		}
	}
	return dataList
}

// DeleteData delete data.
func (tb *DefaultTable) DeleteData(id string) error {
	var err error
	ids := strings.Split(id, ",")

	if tb.Info.DeleteHook != nil {
		defer func() {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error(r)
						logger.Error(string(debug.Stack()))
					}
				}()
				if hookErr := tb.Info.DeleteHook(ids); hookErr != nil {
					logger.Error(hookErr)
				}
			}()
		}()
	}

	if tb.Info.DeleteHookWithRes != nil {
		defer func() {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error(r)
						logger.Error(string(debug.Stack()))
					}
				}()
				if hookErr := tb.Info.DeleteHookWithRes(ids, err); hookErr != nil {
					logger.Error(hookErr)
				}
			}()
		}()
	}

	if tb.Info.PreDeleteFn != nil {
		if err = tb.Info.PreDeleteFn(ids); err != nil {
			return err
		}
	}

	if tb.Info.DeleteFn != nil {
		err = tb.Info.DeleteFn(ids)
		return err
	}

	if len(ids) == 0 || tb.Info.Table == "" {
		err = errors.New("delete error: missing parameter")
		return err
	}

	err = tb.delete(tb.Info.Table, tb.PrimaryKey.Name, ids)
	return err
}

func (tb *DefaultTable) GetNewFormInfo() FormInfo {
	f := tb.GetActualNewForm()
	if len(f.TabGroups) == 0 {
		return FormInfo{ FieldList: f.FieldsWithDefaultValue(tb.sqlObjOrNil) }
	}
	newForm, headers := f.GroupField(tb.sqlObjOrNil)
	return FormInfo{GroupFieldList: newForm, GroupFieldHeaders: headers}
}

// ***************************************
// helper function for database operation
// ***************************************

func (tb *DefaultTable) delete(table, key string, values []string) error {
	var vals = make([]interface{}, len(values))
	for i, v := range values {
		vals[i] = v
	}
	return tb.sql().Table(table).WhereIn(key, vals).Delete()
}

func (tb *DefaultTable) getTheadAndFilterForm(params parameter.Parameters, columnMap map[string]struct{}) (types.Thead, string, string, string, map[string]struct{}, []types.FormField) {
	return tb.Info.FieldList.GetTheadAndFilterForm(types.TableInfo{
		Table:      tb.Info.Table,
		Delimiter:  tb.delimiter(),
		Delimiter2: tb.delimiter2(),
		Driver:     tb.connectionDriver,
		PrimaryKey: tb.PrimaryKey.Name,
	}, params, columnMap, tb.sqlObjOrNil)
}

// db is a helper function return raw db connection.
func (tb *DefaultTable) db() db.Connection {
	if tb.dbObj == nil {
		tb.dbObj = db.GetConnectionFromService(services.MustGet(tb.connectionDriver))
	}
	return tb.dbObj
}

func (tb *DefaultTable) delimiter() string {
	if tb.getDataFromDB() {
		return tb.db().GetDelimiter()
	}
	return ""
}

func (tb *DefaultTable) delimiter2() string {
	if tb.getDataFromDB() {
		return tb.db().GetDelimiter2()
	}
	return ""
}

func (tb *DefaultTable) getDataFromDB() bool {
	return tb.sourceURL == "" && tb.getDataFun == nil && tb.Info.GetDataFn == nil && tb.Detail.GetDataFn == nil
}

// sql is a helper function return db sql.
func (tb *DefaultTable) sql() *db.SQL {
	return db.WithDriverAndConnection(tb.connection, tb.db())
}

// sqlObjOrNil is a helper function return db sql obj or nil.
func (tb *DefaultTable) sqlObjOrNil() *db.SQL {
	if tb.connectionDriver != "" && tb.getDataFromDB() {
		return db.WithDriverAndConnection(tb.connection, tb.db())
	}
	return nil
}

type Columns []string

func (tb *DefaultTable) getColumns(table string) (Columns, bool) {
	columnsModel, _ := tb.sql().Table(table).ShowColumns()
	columns := make(Columns, len(columnsModel))

	switch tb.connectionDriver {
	case db.DriverPostgresql:
		auto := false
		for i, model := range columnsModel {
			col := model["column_name"].(string)
			columns[i] = col
			if !auto && col == tb.PrimaryKey.Name {
				v, _ := model["column_default"].(string)
				if strings.Contains(v, "nextval") {
					auto = true
				}
			}
		}
		return columns, auto
	case db.DriverMysql:
		auto := false
		for i, model := range columnsModel {
			col := model["Field"].(string)
			columns[i] = col
			if !auto && col == tb.PrimaryKey.Name {
				v, _ := model["Extra"].(string)
				if v == "auto_increment" {
					auto = true
				}
			}
		}
		return columns, auto
	case db.DriverSqlite:
		for i, model := range columnsModel {
			columns[i] = model["name"].(string)
		}
		num, _ := tb.sql().Table("sqlite_sequence").Where("name", "=", tb.GetForm().Table).Count()
		return columns, num > 0
	case db.DriverMssql:
		for i, model := range columnsModel {
			columns[i] = model["column_name"].(string)
		}
		return columns, true
	default:
		panic("wrong driver")
	}
}

func (tb *DefaultTable) getColumnMap(table string) (map[string]struct{}, bool) {
	columnsModel, _ := tb.sql().Table(table).ShowColumns()
	columns := make(map[string]struct{}, len(columnsModel))

	switch tb.connectionDriver {
	case db.DriverPostgresql:
		auto := false
		for _, model := range columnsModel {
			col := model["column_name"].(string)
			columns[col] = struct{}{}
			if !auto && col == tb.PrimaryKey.Name {
				v, _ := model["column_default"].(string)
				if strings.Contains(v, "nextval") {
					auto = true
				}
			}
		}
		return columns, auto
	case db.DriverMysql:
		auto := false
		for _, model := range columnsModel {
			col := model["Field"].(string)
			columns[col] = struct{}{}
			if !auto && col == tb.PrimaryKey.Name {
				v, _ := model["Extra"].(string)
				if v == "auto_increment" {
					auto = true
				}
			}
		}
		return columns, auto
	case db.DriverSqlite:
		for _, model := range columnsModel {
			columns[model["name"].(string)] = struct{}{}
		}
		num, _ := tb.sql().Table("sqlite_sequence").Where("name", "=", tb.GetForm().Table).Count()
		return columns, num > 0
	case db.DriverMssql:
		for _, model := range columnsModel {
			columns[model["column_name"].(string)] = struct{}{}
		}
		return columns, true
	default:
		panic("wrong driver")
	}
}
