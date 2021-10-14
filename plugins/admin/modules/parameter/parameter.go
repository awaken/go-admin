package parameter

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/modules/utils"
)

type Parameters struct {
	Page          string
	PageInt       int
	PageSize      string
	PageSizeInt   int
	SortField     string
	Columns       []string
	SortType      string
	Animation     bool
	URLPath       string
	Fields        map[string][]string
	OrConditions  map[string]string
	cacheFixedStr url.Values
}

const (
	Page     = "__page"
	PageSize = "__pageSize"
	Sort     = "__sort"
	SortType = "__sort_type"
	Columns  = "__columns"
	Prefix   = "__prefix"
	Pjax     = "_pjax"

	sortTypeDesc = "desc"
	sortTypeAsc  = "asc"

	IsAll      = "__is_all"
	PrimaryKey = "__pk"

	True  = "true"
	False = "false"

	FilterRangeParamStartSuffix = "_start__goadmin"
	FilterRangeParamEndSuffix   = "_end__goadmin"
	FilterParamJoinInfix        = "_goadmin_join_"
	FilterParamOperatorSuffix   = "__goadmin_operator__"
	FilterParamCountInfix       = "__goadmin_index__"

	Separator = "__goadmin_separator__"
)

var operators = map[string]string{
	"like": "like",
	"gr":   ">",
	"gq":   ">=",
	"eq":   "=",
	"ne":   "!=",
	"le":   "<",
	"lq":   "<=",
	"free": "free",
}

var globKeyMap = map[string]struct{}{
	Page: {}, PageSize: {}, Sort: {}, Columns: {}, Prefix: {}, Pjax: {}, form.NoAnimationKey: {},
}

func BaseParam() Parameters {
	return Parameters{ Page: "1", PageSize: "10", PageInt: 1, PageSizeInt: 10, Fields: make(map[string][]string) }
}

func GetParam(u *url.URL, defaultPageSize int, p ...string) Parameters {
	values := u.Query()

	primaryKey := "id"
	defaultSortType := "desc"

	if len(p) > 0 {
		primaryKey = p[0]
		defaultSortType = p[1]
	}

	page := getDefault(values, Page, "1")
	pageSize := getDefault(values, PageSize, strconv.Itoa(defaultPageSize))
	sortField := getDefault(values, Sort, primaryKey)
	sortType := getDefault(values, SortType, defaultSortType)
	columns := getDefault(values, Columns, "")

	animation := true
	if values.Get(form.NoAnimationKey) == "true" {
		animation = false
	}

	fields := make(map[string][]string, 8)

	for key, value := range values {
		if len(value) > 0 && value[0] != "" {
			if _, ok := globKeyMap[key]; !ok {
				if key == SortType {
					if value[0] != sortTypeDesc && value[0] != sortTypeAsc {
						fields[key] = []string{ sortTypeDesc }
					}
				} else {
					if strings.Contains(key, FilterParamOperatorSuffix) &&
						values.Get(strings.ReplaceAll(key, FilterParamOperatorSuffix, "")) == "" {
						continue
					}
					fields[strings.ReplaceAll(key, "[]", "")] = value
				}
			}
		}
	}

	var columnsArr []string
	if columns != "" {
		columns, _ = url.QueryUnescape(columns)
		columnsArr = strings.Split(columns, ",")
	}

	pageInt    , _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	return Parameters{
		Page:         page,
		PageSize:     pageSize,
		PageSizeInt:  pageSizeInt,
		PageInt:      pageInt,
		URLPath:      u.Path,
		SortField:    sortField,
		SortType:     sortType,
		Fields:       fields,
		OrConditions: map[string]string{},
		Animation:    animation,
		Columns:      columnsArr,
	}
}

func GetParamFromURL(urlStr string, defaultPageSize int, defaultSortType, primaryKey string) Parameters {
	u, err := url.Parse(urlStr)
	if err != nil { return BaseParam() }
	return GetParam(u, defaultPageSize, primaryKey, defaultSortType)
}

func (param Parameters) WithPKs(id ...string) Parameters {
	param.Fields[PrimaryKey] = []string{ strings.Join(id, ",") }
	return param
}

func (param Parameters) PKs() []string {
	pk := param.GetFieldValue(PrimaryKey)
	if pk == "" { return nil }
	return strings.Split(pk, ",")
}

func (param Parameters) DeletePK() Parameters {
	delete(param.Fields, PrimaryKey)
	return param
}

func (param Parameters) PK() string {
	pks := param.PKs()
	if len(pks) > 0 { return pks[0] }
	return ""
}

func (param Parameters) IsAll() bool {
	return param.GetFieldValue(IsAll) == True
}

func (param Parameters) WithURLPath(path string) Parameters {
	param.URLPath = path
	return param
}

func (param Parameters) WithIsAll(isAll bool) Parameters {
	if isAll {
		param.Fields[IsAll] = []string{ True }
	} else {
		param.Fields[IsAll] = []string{ False }
	}
	return param
}

func (param Parameters) DeleteIsAll() Parameters {
	delete(param.Fields, IsAll)
	return param
}

func (param Parameters) GetFilterFieldValueStart(field string) string {
	return param.GetFieldValue(field + FilterRangeParamStartSuffix)
}

func (param Parameters) GetFilterFieldValueEnd(field string) string {
	return param.GetFieldValue(field + FilterRangeParamEndSuffix)
}

func (param Parameters) GetFieldValue(field string) string {
	value, _ := param.Fields[field]
	if len(value) > 0 {
		return value[0]
	}
	return ""
}

func (param Parameters) AddField(field, value string) Parameters {
	param.Fields[field] = []string{ value }
	return param
}

func (param Parameters) DeleteField(fields ...string) Parameters {
	for _, field := range fields {
		delete(param.Fields, field)
	}
	return param
}

func (param Parameters) DeleteEditPk() Parameters {
	delete(param.Fields, constant.EditPKKey)
	return param
}

func (param Parameters) DeleteDetailPk() Parameters {
	delete(param.Fields, constant.DetailPKKey)
	return param
}

func (param Parameters) GetFieldValues(field string) []string {
	return param.Fields[field]
}

func (param Parameters) GetFieldValuesStr(field string) string {
	return strings.Join(param.Fields[field], Separator)
}

func (param Parameters) GetFieldOperator(field, suffix string) string {
	op := param.GetFieldValue(field + FilterParamOperatorSuffix + suffix)
	if op == "" { return "eq" }
	return op
}

func (param Parameters) Join() string {
	p := param.GetFixedParamStr()
	p.Add(Page, param.Page)
	return p.Encode()
}

func (param Parameters) SetPage(page string) Parameters {
	param.Page = page
	param.PageInt, _ = strconv.Atoi(page)
	return param
}

func (param Parameters) SetPageSize(pageSize string) Parameters {
	param.PageSize = pageSize
	param.PageSizeInt, _ = strconv.Atoi(pageSize)
	return param
}

func (param Parameters) GetRouteParamStr() string {
	p := param.GetFixedParamStr()
	p.Add(Page, param.Page)
	return "?" + p.Encode()
}

func (param Parameters) URL(page string) string {
	return param.URLPath + param.SetPage(page).GetRouteParamStr()
}

func (param Parameters) URLNoAnimation(page string) string {
	return param.URLPath + param.SetPage(page).GetRouteParamStr() + "&" + form.NoAnimationKey + "=true"
}

func (param Parameters) GetRouteParamStrWithoutPageSize(page string) string {
	p := make(url.Values, 4 + len(param.Fields))
	p.Add(Sort, param.SortField)
	p.Add(Page, page)
	p.Add(SortType, param.SortType)
	if len(param.Columns) > 0 {
		p.Add(Columns, strings.Join(param.Columns, ","))
	}
	for key, value := range param.Fields {
		p[key] = value
	}
	return "?" + p.Encode()
}

func (param Parameters) GetFixedParamStrFromCache() url.Values {
	if param.cacheFixedStr != nil { return param.cacheFixedStr }
	p := param.GetFixedParamStr()
	param.cacheFixedStr = p
	return p
}

func (param Parameters) GetLastPageRouteParamStr(cache ...bool) string {
	var p url.Values
	if len(cache) > 0 && cache[0] {
		p = param.GetFixedParamStrFromCache()
	} else {
		p = param.GetFixedParamStr()
	}
	p.Add(Page, strconv.Itoa(param.PageInt - 1))
	return "?" + p.Encode()
}

func (param Parameters) GetNextPageRouteParamStr(cache ...bool) string {
	var p url.Values
	if len(cache) > 0 && cache[0] {
		p = param.GetFixedParamStrFromCache()
	} else {
		p = param.GetFixedParamStr()
	}
	p.Add(Page, strconv.Itoa(param.PageInt + 1))
	return "?" + p.Encode()
}

func (param Parameters) GetFixedParamStr() url.Values {
	p := make(url.Values, 4 + len(param.Fields))
	p.Add(Sort, param.SortField)
	p.Add(PageSize, param.PageSize)
	p.Add(SortType, param.SortType)
	if len(param.Columns) > 0 {
		p.Add(Columns, strings.Join(param.Columns, ","))
	}
	for key, value := range param.Fields {
		p[key] = value
	}
	return p
}

func (param Parameters) GetFixedParamStrWithoutColumnsAndPage() string {
	p := make(url.Values, 4)
	p.Add(Sort, param.SortField)
	p.Add(PageSize, param.PageSize)
	if len(param.Columns) > 0 {
		p.Add(Columns, strings.Join(param.Columns, ","))
	}
	p.Add(SortType, param.SortType)
	return "?" + p.Encode()
}

func (param Parameters) GetFixedParamStrWithoutSort() string {
	p := make(url.Values, 3 + len(param.Columns))
	p.Add(PageSize, param.PageSize)
	for key, value := range param.Fields {
		p[key] = value
	}
	p.Add(form.NoAnimationKey, "true")
	if len(param.Columns) > 0 {
		p.Add(Columns, strings.Join(param.Columns, ","))
	}
	return "&" + p.Encode()
}

func (param Parameters) Statement(wheres, table, delimiter, delimiter2 string, whereArgs []interface{}, columnMap, existKeys map[string]struct{}, filterProcess func(string, string, string) string) (string, []interface{}, map[string]struct{}) {
	var multiKey map[string]struct{}
	var sbWhr strings.Builder
	sbWhr.Grow(len(wheres) + 128)
	sbWhr.WriteString(wheres)

	for key, value := range param.Fields {
		keyIndexSuffix := ""
		if p := strings.Index(key, FilterParamCountInfix); p >= 0 {
			key = key[:p]
			keyIndexSuffix = key[p:]
			if multiKey == nil {
				multiKey = map[string]struct{}{ key: {} }
			} else {
				multiKey[key] = struct{}{}
			}
		} else if !utils.InMapT(multiKey, key) && utils.InMapT(existKeys, key) {
			continue
		}
		/*keyArr := strings.Split(key, FilterParamCountInfix)
		if len(keyArr) > 1 {
			key = keyArr[0]
			keyIndexSuffix = FilterParamCountInfix + keyArr[1]
			multiKey[key] = struct{}{}
		} else if _, ok := multiKey[key]; !ok && modules.InArray(existKeys, key) {
			continue
		}*/

		var op string
		if strings.Contains(key, FilterRangeParamEndSuffix) {
			key = strings.ReplaceAll(key, FilterRangeParamEndSuffix, "")
			op  = "<="
		} else if strings.Contains(key, FilterRangeParamStartSuffix) {
			key = strings.ReplaceAll(key, FilterRangeParamStartSuffix, "")
			op  = ">="
		} else if len(value) > 1 {
			op = "in"
		} else if !strings.Contains(key, FilterParamOperatorSuffix) {
			op = operators[param.GetFieldOperator(key, keyIndexSuffix)]
		} else {
			continue
		}

		if p := strings.Index(key, FilterParamJoinInfix); p >= 0 {
			if sbWhr.Len() > 0 { sbWhr.WriteString(" AND ") }
			sbWhr.WriteString(key[:p])
			sbWhr.WriteByte('.')
			sbWhr.WriteString(delimiter)
			sbWhr.WriteString(key[p + len(FilterParamJoinInfix):])
			sbWhr.WriteString(delimiter2)
			sbWhr.WriteByte(' ')
			sbWhr.WriteString(op)
			//keys := strings.Split(key, FilterParamJoinInfix)
			//wheres += keys[0] + "." + modules.FilterField(keys[1], delimiter, delimiter2) + " " + op
			if op == "in" {
				sbWhr.WriteString(" (?")
				for n := len(value); n > 1; n-- {
					sbWhr.WriteString(",?")
				}
				sbWhr.WriteByte(')')
				//qmark := ""
				//for range value { qmark += "?," }
				//wheres += " (" + qmark[:len(qmark)-1] + ") and "
			} else {
				sbWhr.WriteString(" ?")
				//wheres += " ? and "
			}
			val := filterProcess(key, value[0], keyIndexSuffix)
			if op == "like" && !strings.ContainsRune(val, '%') {
				whereArgs = append(whereArgs, utils.StrConcat("%", val, "%"))
			} else {
				for _, v := range value {
					whereArgs = append(whereArgs, filterProcess(key, v, keyIndexSuffix))
				}
			}
		} else if _, ok := columnMap[key]; ok {
			if sbWhr.Len() > 0 { sbWhr.WriteString(" AND ") }
			sbWhr.WriteString(table)
			sbWhr.WriteByte('.')
			sbWhr.WriteString(delimiter)
			sbWhr.WriteString(key)
			sbWhr.WriteString(delimiter2)
			sbWhr.WriteByte(' ')
			sbWhr.WriteString(op)
			//wheres += table + "." + modules.FilterField(key, delimiter, delimiter2) + " " + op
			if op == "in" {
				sbWhr.WriteString(" (?")
				for n := len(value); n > 1; n-- {
					sbWhr.WriteString(",?")
				}
				sbWhr.WriteByte(')')
				//qmark := ""
				//for range value { qmark += "?," }
				//wheres += " (" + qmark[:len(qmark)-1] + ") and "
			} else {
				sbWhr.WriteString(" ?")
				//wheres += " ? and "
			}
			if op == "like" && !strings.ContainsRune(value[0], '%') {
				whereArgs = append(whereArgs, utils.StrConcat("%", filterProcess(key, value[0], keyIndexSuffix), "%"))
			} else {
				for _, v := range value {
					whereArgs = append(whereArgs, filterProcess(key, v, keyIndexSuffix))
				}
			}
		} else {
			continue
		}

		if existKeys == nil {
			existKeys = map[string]struct{}{ key: {} }
		} else {
			existKeys[key] = struct{}{}
		}
	}

	//if len(wheres) > 3 {
	//	wheres = wheres[:len(wheres)-4]
	//}

	for key, value := range param.OrConditions {
		op := "="
		if strings.ContainsRune(value, '%') {
			op = "like"
		}
		if sbWhr.Len() > 0 {
			sbWhr.WriteString(" AND (")
		} else {
			sbWhr.WriteByte('(')
		}
		//if len(wheres) > 0 {
		//	wheres += " and "
		//}
		//wheres += "("
		for i, column := range strings.Split(key, ",") {
			if i > 0 { sbWhr.WriteString(" OR ") }
			if p := strings.Index(column, FilterParamJoinInfix); p >= 0 {
				sbWhr.WriteString(column[:p])
				sbWhr.WriteByte('.')
			}
			//keys := strings.Split(column, FilterParamJoinInfix)
			//if len(keys) > 1 {
			//	wheres += keys[0] + "."
			//}
			sbWhr.WriteString(delimiter)
			sbWhr.WriteString(column)
			sbWhr.WriteString(delimiter2)
			sbWhr.WriteByte(' ')
			sbWhr.WriteString(op)
			sbWhr.WriteString(" ?")
			//wheres += modules.FilterField(column, delimiter, delimiter2) + " " + op + " ? or "
			whereArgs = append(whereArgs, value)
		}
		sbWhr.WriteByte(')')
		//wheres = strings.TrimSuffix(wheres, "or ") + ")"
	}

	return sbWhr.String(), whereArgs, existKeys
	//return wheres, whereArgs, existKeys
}

func getDefault(values url.Values, key, def string) string {
	value := values.Get(key)
	if value == "" { return def }
	return value
}
