package types

import (
	"fmt"
	"html/template"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/template/types/form"
	"github.com/GoAdminGroup/go-admin/template/types/table"
)

// FieldModel is the single query result.
type FieldModel struct {
	// The primaryKey of the table.
	ID string

	// The value of the single query result.
	Value string

	// The current row data.
	Row map[string]interface{}

	// Post type
	PostType PostType
}

type PostType uint8

const (
	PostTypeCreate = iota
	PostTypeUpdate
)

func (m FieldModel) IsCreate() bool {
	return m.PostType == PostTypeCreate
}

func (m FieldModel) IsUpdate() bool {
	return m.PostType == PostTypeUpdate
}

// PostFieldModel contains ID and value of the single query result and the current row data.
type PostFieldModel struct {
	ID       string
	Value    FieldModelValue
	Row      map[string]string
	PostType PostType
}

func (m PostFieldModel) IsCreate() bool {
	return m.PostType == PostTypeCreate
}

func (m PostFieldModel) IsUpdate() bool {
	return m.PostType == PostTypeUpdate
}

type InfoItem struct {
	Content template.HTML `json:"content"`
	Value   string        `json:"value"`
}

type InfoList []map[string]InfoItem

func (il InfoList) GroupBy(groups TabGroups) []InfoList {
	res := make([]InfoList, len(groups))
	for i, value := range groups {
		newInfoList := make(InfoList, len(il))
		for j, info := range il {
			newRow := make(map[string]InfoItem)
			for k, m := range info {
				if modules.InArray(value, k) {
					newRow[k] = m
				}
			}
			newInfoList[j] = newRow
		}
		res[i] = newInfoList
	}
	return res
}

type Callbacks []context.Node

func (c Callbacks) AddCallback(node context.Node) Callbacks {
	if node.Path != "" && node.Method != "" && len(node.Handlers) > 0 {
		for _, item := range c {
			if strings.EqualFold(item.Path, node.Path) &&
				strings.EqualFold(item.Method, node.Method) {
				return c
			}
		}
		if p := strings.IndexByte(node.Path, '?'); p >= 0 {
			node.Path = node.Path[:p]
		}
		return append(c, node)
	}
	return c
}

type FieldModelValue []string

func (r FieldModelValue) Value() string {
	return r.First()
}

func (r FieldModelValue) First() string {
	if len(r) > 0 { return r[0] }
	return ""
}

// FieldDisplay is filter function of data.
type FieldFilterFn func(value FieldModel) interface{}

// PostFieldFilterFn is filter function of data.
type PostFieldFilterFn func(value PostFieldModel) interface{}

// Field is the table field.
type Field struct {
	Head     string
	Field    string
	TypeName db.DatabaseType

	Joins Joins

	Width      int
	Sortable   bool
	EditAble   bool
	Fixed      bool
	Filterable bool
	Hide       bool

	EditType    table.Type
	EditOptions FieldOptions

	FilterFormFields []FilterFormField

	IsEditParam   bool
	IsDeleteParam bool
	IsDetailParam bool

	FieldDisplay
}

type QueryFilterFn func(param parameter.Parameters, conn db.Connection) (ids []string, stopQuery bool)
type UpdateParametersFn func(param *parameter.Parameters)

type FilterFormField struct {
	Type        form.Type
	Options     FieldOptions
	OptionTable OptionTable
	Width       int
	HeadWidth   int
	InputWidth  int
	Style       template.HTMLAttr
	Operator    FilterOperator
	OptionExt   template.JS
	Head        string
	Placeholder string
	HelpMsg     template.HTML
	NoIcon      bool
	ProcessFn   func(string) string
}

func (f Field) GetFilterFormFields(params parameter.Parameters, headField string, sqls ...*db.SQL) []FormField {
	var value, value2, keySuffix string
	var sql *db.SQL
	if len(sqls) > 0 { sql = sqls[0] }

	filterForm := make([]FormField, 0, len(f.FilterFormFields))
	headFieldWithKeySuffix := headField

	for i, filter := range f.FilterFormFields {
		if i > 0 {
			keySuffix = parameter.FilterParamCountInfix + strconv.Itoa(i)
			headFieldWithKeySuffix += keySuffix
		}

		if filter.Type.IsRange() {
			value  = params.GetFilterFieldValueStart(headField)
			value2 = params.GetFilterFieldValueEnd(headField)
		} else if filter.Type.IsMultiSelect() {
			value  = params.GetFieldValuesStr(headField)
		} else {
			if filter.Operator == FilterOperatorFree {
				value2 = GetOperatorFromValue(params.GetFieldOperator(headField, keySuffix)).String()
			}
			value = params.GetFieldValue(headFieldWithKeySuffix)
		}

		var (
			optionExt1 = filter.OptionExt
			optionExt2 template.JS
		)

		if filter.OptionExt == "" {
			op1, op2, js := filter.Type.GetDefaultOptions(headFieldWithKeySuffix)
			if op1 != nil {
				s, _ := utils.JsonMarshal(op1)
				optionExt1 = template.JS(s)
			}
			if op2 != nil {
				s, _ := utils.JsonMarshal(op2)
				optionExt2 = template.JS(s)
			}
			if js != "" {
				optionExt1 = js
			}
		}

		field := &FormField{
			Field:       headFieldWithKeySuffix,
			FieldClass:  headFieldWithKeySuffix,
			Head:        filter.Head,
			TypeName:    f.TypeName,
			HelpMsg:     filter.HelpMsg,
			NoIcon:      filter.NoIcon,
			FormType:    filter.Type,
			Editable:    true,
			Width:       filter.Width,
			HeadWidth:   filter.HeadWidth,
			InputWidth:  filter.InputWidth,
			Style:       filter.Style,
			Placeholder: filter.Placeholder,
			Value:       template.HTML(value),
			Value2:      value2,
			Options:     filter.Options,
			OptionExt:   optionExt1,
			OptionExt2:  optionExt2,
			OptionTable: filter.OptionTable,
			Label:       filter.Operator.Label(),
		}

		field.setOptionsFromSQL(sql)

		if filter.Type.IsSingleSelect() {
			field.Options = field.Options.SetSelected(params.GetFieldValue(f.Field), filter.Type.SelectedLabel())
		} else if filter.Type.IsMultiSelect() {
			field.Options = field.Options.SetSelected(params.GetFieldValues(f.Field), filter.Type.SelectedLabel())
		}

		filterForm = append(filterForm, *field)

		if filter.Operator.AddOrNot() {
			ff := utils.StrConcat(headField, parameter.FilterParamOperatorSuffix, keySuffix)
			filterForm = append(filterForm, FormField{
				Field:      ff,
				FieldClass: ff,
				Head:       f.Head,
				TypeName:   f.TypeName,
				Value:      template.HTML(filter.Operator.Value()),
				FormType:   filter.Type,
				Hide:       true,
			})
		}
	}

	return filterForm
}

func (f Field) Exist() bool {
	return f.Field != ""
}

type FieldList []Field

type TableInfo struct {
	Table      string
	PrimaryKey string
	Delimiter  string
	Delimiter2 string
	Driver     string
}

func (f FieldList) GetTheadAndFilterForm(info TableInfo, params parameter.Parameters, columnMap map[string]struct{}, sqlFuncs ...func() *db.SQL) (Thead, string, string, string, map[string]struct{}, []FormField) {
	var (
		fields       strings.Builder
		joinFields   strings.Builder
		joins        strings.Builder
		joinTableMap map[string]struct{}
		sql          func() *db.SQL
	)
	if len(sqlFuncs) > 0 { sql = sqlFuncs[0] }
	if sql == nil { sql = db.NilSQL }

	thead      := make(Thead      , 0, len(f))
	filterForm := make([]FormField, 0, len(f))
	fields.Grow(256)

	for _, field := range f {
		if field.Field != info.PrimaryKey && utils.InMapT(columnMap, field.Field) && !field.Joins.Valid() {
			fields.WriteString(info.Delimiter)
			fields.WriteString(info.Table)
			fields.WriteString(info.Delimiter2)
			fields.WriteByte('.')
			fields.WriteString(info.Delimiter)
			fields.WriteString(field.Field)
			fields.WriteString(info.Delimiter2)
			fields.WriteByte(',')
		}

		headField := field.Field

		if field.Joins.Valid() {
			lastJoin := field.Joins.Last()
			headField = utils.StrConcat(lastJoin.GetTableName(), parameter.FilterParamJoinInfix, field.Field)
			var sb strings.Builder
			sb.Grow(64)
			sb.WriteString(lastJoin.GetTableName(info.Delimiter, info.Delimiter2))
			sb.WriteByte('.')
			sb.WriteString(info.Delimiter)
			sb.WriteString(field.Field)
			sb.WriteString(info.Delimiter2)
			joinFields.WriteString(db.GetAggregationExpression(info.Driver, sb.String(), headField, JoinFieldValueDelimiter))
			joinFields.WriteByte(',')
			for _, join := range field.Joins {
				joinTableName := join.GetTableName(info.Delimiter, info.Delimiter2)
				if _, ok := joinTableMap[joinTableName]; !ok {
					if joinTableMap == nil {
						joinTableMap = map[string]struct{}{ joinTableName: {} }
					} else {
						joinTableMap[joinTableName] = struct{}{}
					}
					if join.BaseTable == "" {
						join.BaseTable = info.Table
					}
					joins.WriteString(" LEFT JOIN ")
					joins.WriteString(modules.FilterField(join.Table, info.Delimiter, info.Delimiter2))
					if join.TableAlias != "" {
						joins.WriteByte(' ')
						joins.WriteString(join.TableAlias)
					}
					joins.WriteString(" ON ")
					joins.WriteString(joinTableName)
					joins.WriteByte('.')
					joins.WriteString(modules.FilterField(join.JoinField, info.Delimiter, info.Delimiter2))
					joins.WriteByte('=')
					joins.WriteString(modules.Delimiter(info.Delimiter, info.Delimiter2, join.BaseTable))
					joins.WriteByte('.')
					joins.WriteString(modules.FilterField(join.Field, info.Delimiter, info.Delimiter2))
				}
			}
		}

		if field.Filterable {
			filterForm = append(filterForm, field.GetFilterFormFields(params, headField, sql())...)
		}
		if field.Hide { continue }

		thead = append(thead, TheadItem{
			Head:       field.Head,
			Sortable:   field.Sortable,
			Field:      headField,
			Hide:       !modules.InArrayWithoutEmpty(params.Columns, headField),
			Editable:   field.EditAble,
			EditType:   field.EditType.String(),
			EditOption: field.EditOptions,
			Width:      strconv.Itoa(field.Width) + "px",
		})
	}

	return thead, fields.String(), joinFields.String(), joins.String(), joinTableMap, filterForm
}

func (f FieldList) GetThead(info TableInfo, params parameter.Parameters, columnMap map[string]struct{}) (Thead, string, string) {
	var (
		sbFields     strings.Builder
		sbJoins      strings.Builder
		joinTableMap map[string]struct{}
	)
	thead := make(Thead, 0, len(f))
	sbFields.Grow(256)

	for _, field := range f {
		if field.Field != info.PrimaryKey {
			if _, ok := columnMap[field.Field]; ok && !field.Joins.Valid() {
				sbFields.WriteString(info.Delimiter)
				sbFields.WriteString(info.Table)
				sbFields.WriteString(info.Delimiter2)
				sbFields.WriteByte('.')
				sbFields.WriteString(info.Delimiter)
				sbFields.WriteString(field.Field)
				sbFields.WriteString(info.Delimiter2)
				sbFields.WriteByte(',')
			}
		}

		headField := field.Field

		if field.Joins.Valid() {
			headField = field.Joins.Last().GetTableName(info.Delimiter, info.Delimiter2) + parameter.FilterParamJoinInfix + field.Field
			for _, join := range field.Joins {
				joinTableName := join.GetTableName(info.Delimiter, info.Delimiter2)
				if _, ok := joinTableMap[joinTableName]; !ok {
					if joinTableMap == nil {
						joinTableMap = map[string]struct{}{ joinTableName: {} }
					} else {
						joinTableMap[joinTableName] = struct{}{}
					}
					if join.BaseTable == "" {
						join.BaseTable = info.Table
					}
					if sbJoins.Len() == 0 {
						sbJoins.Grow(256)
					} else {
						sbJoins.WriteByte(' ')
					}
					sbJoins.WriteString("LEFT JOIN ")
					sbJoins.WriteString(info.Delimiter)
					sbJoins.WriteString(join.Table)
					sbJoins.WriteString(info.Delimiter2)
					if join.TableAlias != "" {
						sbJoins.WriteByte(' ')
						sbJoins.WriteString(join.TableAlias)
					}
					sbJoins.WriteString(" ON ")
					sbJoins.WriteString(joinTableName)
					sbJoins.WriteByte('.')
					sbJoins.WriteString(info.Delimiter)
					sbJoins.WriteString(join.JoinField)
					sbJoins.WriteString(info.Delimiter2)
					sbJoins.WriteByte('=')
					sbJoins.WriteString(info.Delimiter)
					sbJoins.WriteString(join.BaseTable)
					sbJoins.WriteString(info.Delimiter2)
					sbJoins.WriteByte('.')
					sbJoins.WriteString(info.Delimiter)
					sbJoins.WriteString(join.Field)
					sbJoins.WriteString(info.Delimiter2)
				}
			}
		}

		if field.Hide { continue }

		thead = append(thead, TheadItem{
			Head:       field.Head,
			Sortable:   field.Sortable,
			Field:      headField,
			Hide:       !modules.InArrayWithoutEmpty(params.Columns, headField),
			Editable:   field.EditAble,
			EditType:   field.EditType.String(),
			EditOption: field.EditOptions,
			Width:      strconv.Itoa(field.Width) + "px",
		})
	}

	return thead, sbFields.String(), sbJoins.String()
}

func (f FieldList) GetFieldFilterProcessValue(key, value, keyIndex string) string {
	field := f.GetFieldByFieldName(key)
	index := 0
	if keyIndex != "" {
		index, _ = strconv.Atoi(keyIndex)
	}
	if field.FilterFormFields != nil && index < len(field.FilterFormFields) {
		fn := field.FilterFormFields[index].ProcessFn
		if fn != nil { value = fn(value) }
	}
	return value
}

func (f FieldList) GetFieldJoinTable(key string) string {
	field := f.GetFieldByFieldName(key)
	if field.Exist() {
		return field.Joins.Last().Table
	}
	return ""
}

func (f FieldList) GetFieldByFieldName(name string) Field {
	for _, field := range f {
		switch name {
		case field.Field, JoinField(field.Joins.Last().GetTableName(), field.Field):
			return field
		}
	}
	return Field{}
}

// Join store join table info. For example:
//
// Join {
//     BaseTable:   "users",
//     Field:       "role_id",
//     Table:       "roles",
//     JoinField:   "id",
// }
//
// It will generate the join table sql like:
//
// ... left join roles on roles.id = users.role_id ...
//
type Join struct {
	Table      string
	TableAlias string
	Field      string
	JoinField  string
	BaseTable  string
}

type Joins []Join

func JoinField(table, field string) string {
	return utils.StrConcat(table, parameter.FilterParamJoinInfix, field)
}

func GetJoinField(field string) string {
	return strings.Split(field, parameter.FilterParamJoinInfix)[1]
}

func (j Joins) Valid() bool {
	for _, e := range j {
		if e.Valid() { return true }
	}
	return false
}

func (j Joins) Last() Join {
	if len(j) > 0 { return j[len(j) - 1] }
	return Join{}
}

func (j Join) Valid() bool {
	return j.Table != "" && j.Field != "" && j.JoinField != ""
}

func (j Join) GetTableName(delimiter ...string) string {
	if j.TableAlias != "" {
		return j.TableAlias
	}
	if len(delimiter) > 1 {
		_ = delimiter[1]
		return utils.StrConcat(delimiter[0], j.Table, delimiter[1])
	}
	return j.Table
}

var JoinFieldValueDelimiter = utils.Uuid(8)

type TabGroups [][]string

func (t TabGroups) Valid() bool {
	return len(t) > 0
}

func NewTabGroups(items ...string) TabGroups {
	if len(items) == 0 { return nil }
	return [][]string{ items }
}

func (t TabGroups) AddGroup(items ...string) TabGroups {
	return append(t, items)
}

type TabHeaders []string

func (t TabHeaders) Add(header string) TabHeaders {
	return append(t, header)
}

type GetDataFn func(param parameter.Parameters) ([]map[string]interface{}, int)

type DeleteFn func(ids []string) error
type DeleteFnWithRes func(ids []string, res error) error

type Sort uint8

const (
	SortDesc Sort = iota
	SortAsc
)

type primaryKey struct {
	Type db.DatabaseType
	Name string
}

type ExportProcessFn func(param parameter.Parameters) (PanelInfo, error)

// InfoPanel
type InfoPanel struct {
	FieldList         FieldList
	curFieldListIndex int

	Table       string
	Title       string
	Description string

	// Warn: may be deprecated future.
	TabGroups  TabGroups
	TabHeaders TabHeaders

	Sort      Sort
	SortField string

	PageSizeList    []int
	DefaultPageSize int

	ExportType      int
	ExportProcessFn ExportProcessFn

	primaryKey primaryKey

	IsHideNewButton    bool
	IsHideExportButton bool
	IsHideEditButton   bool
	IsHideDeleteButton bool
	IsHideDetailButton bool
	IsHideFilterButton bool
	IsHideRowSelector  bool
	IsHidePagination   bool
	IsHideFilterArea   bool
	IsHideQueryInfo    bool
	FilterFormLayout   form.Layout

	FilterFormHeadWidth  int
	FilterFormInputWidth int

	Wheres    Wheres
	WhereRaws WhereRaw

	Callbacks Callbacks

	Buttons Buttons

	TableLayout string

	DeleteHook  DeleteFn
	PreDeleteFn DeleteFn
	DeleteFn    DeleteFn

	DeleteHookWithRes DeleteFnWithRes

	GetDataFn GetDataFn

	processChains DisplayProcessFnChains

	ActionButtons    Buttons
	ActionButtonFold bool

	DisplayGeneratorRecords map[string]struct{}

	QueryFilterFn       QueryFilterFn
	UpdateParametersFns []UpdateParametersFn

	Wrapper ContentWrapper

	// column operation buttons
	Action     template.HTML
	HeaderHtml template.HTML
	FooterHtml template.HTML

	PageError     errors.PageError
	PageErrorHTML template.HTML

	NoCompress  bool
	HideSideBar bool

	AutoRefresh uint
}

type Where struct {
	Join     string
	Field    string
	Operator string
	Arg      interface{}
}

type Wheres []Where

func (whs Wheres) Statement(wheres, delimiter, delimiter2 string, whereArgs []interface{}, existKeys, columnMap map[string]struct{}) (string, []interface{}) {
	var pwheres strings.Builder
	hasWheres := wheres != ""
	//pwheres := ""
	last := len(whs) - 1

	for i, wh := range whs {
		whField, whTable := "", ""
		if p := strings.IndexByte(wh.Field, '.'); p >= 0 {
			whTable = wh.Field[:p]
			whField = wh.Field[p+1:]
		} else {
			whField = wh.Field
		}

		if _, ok := existKeys[whField]; ok {
			continue
		}

		// TODO: support like operation and join table
		if _, ok := columnMap[whField]; ok {
			if hasWheres {
				hasWheres = false
				pwheres.WriteString(wheres)
				pwheres.WriteString(" AND ")
			}
			if whTable != "" {
				pwheres.WriteString(delimiter)
				pwheres.WriteString(whTable)
				pwheres.WriteString(delimiter2)
				pwheres.WriteByte('.')
				//pwheres += whTable + "."
			}
			pwheres.WriteString(delimiter)
			pwheres.WriteString(whField)
			pwheres.WriteString(delimiter2)
			pwheres.WriteByte(' ')
			pwheres.WriteString(wh.Operator)
			pwheres.WriteString(" ?")
			if i != last {
				if joinMark := whs[i + 1].Join; joinMark != "" {
					pwheres.WriteByte(' ')
					pwheres.WriteString(joinMark)
				}
			}
			//pwheres += modules.FilterField(whField, delimiter, delimiter2) + " " + wh.Operator + " ? " + joinMark + " "
			whereArgs = append(whereArgs, wh.Arg)
		}
	}

	if hasWheres {
		return wheres, whereArgs
	}
	return pwheres.String(), whereArgs
	//if wheres != "" && pwheres != "" {
	//	wheres += " and "
	//}
	//return wheres + pwheres, whereArgs
}

type WhereRaw struct {
	Raw  string
	Args []interface{}
}

func (wh WhereRaw) nextBoolTokenAt() int {
	idx := 0
	n   := len(wh.Raw)
	for i, c := range wh.Raw {
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' { continue }
		if c == 'a' || c == 'A' {
			if n < i + 3 {				// look for "and"
				break
			} else if c2 := wh.Raw[i+1]; c2 == 'n' || c2 == 'N' {
				if c3 := wh.Raw[i+2]; c3 == 'd' || c3 == 'D' {
					idx = i + 3
				}
			}
			continue
		}
		if c == 'o' || c == 'O' { // look for "or"
			if n < i + 2 {
				break
			} else if c2 := wh.Raw[i+1]; c2 == 'r' || c2 == 'R' {
				idx = i + 2
			}
			continue
		}
		break
	}
	return idx
}

func (wh WhereRaw) Statement(wheres string, whereArgs []interface{}) (string, []interface{}) {
	if wh.Raw == "" {
		return wheres, whereArgs
	}
	if wheres == "" {
		wheres = wh.Raw[wh.nextBoolTokenAt():] + " "
	} else if wh.nextBoolTokenAt() != 0 {
		wheres = utils.StrConcat(wheres, wh.Raw, " ")
	} else {
		wheres = utils.StrConcat(wheres, " AND ", wh.Raw, " ")
	}
	return wheres, append(whereArgs, wh.Args...)
}

type Handler func(ctx *context.Context) (success bool, msg string, data interface{})

func (h Handler) Wrap() context.Handler {
	return func(ctx *context.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(r)
				logger.Error(string(debug.Stack()))

				ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
					"code": 500,
					"data": "",
					"msg" : "internal server error",
				})
			}
		}()

		code := 0
		s, m, d := h(ctx)
		if !s { code = 500 }

		ctx.JSON(http.StatusOK, map[string]interface{}{
			"code": code,
			"data": d,
			"msg" : m,
		})
	}
}

type ContentWrapper func(content template.HTML) template.HTML

type Action interface {
	Js() template.JS
	BtnAttribute() template.HTML
	BtnClass() template.HTML
	ExtContent() template.HTML
	FooterContent() template.HTML
	SetBtnId(btnId string)
	SetBtnData(data interface{})
	GetCallbacks() context.Node
}

type NilAction struct{}

func (def *NilAction) SetBtnId(btnId string)        {}
func (def *NilAction) SetBtnData(data interface{})  {}
func (def *NilAction) Js() template.JS              { return "" }
func (def *NilAction) BtnAttribute() template.HTML  { return "" }
func (def *NilAction) BtnClass() template.HTML      { return "" }
func (def *NilAction) ExtContent() template.HTML    { return "" }
func (def *NilAction) FooterContent() template.HTML { return "" }
func (def *NilAction) GetCallbacks() context.Node   { return context.Node{} }

type Actions []Action

type DefaultAction struct {
	Attr   template.HTML
	JS     template.JS
	Ext    template.HTML
	Footer template.HTML
}

func NewDefaultAction(attr, ext, footer template.HTML, js template.JS) *DefaultAction {
	return &DefaultAction{ Attr: attr, Ext: ext, Footer: footer, JS: js }
}

func (def *DefaultAction) SetBtnId(btnId string)        {}
func (def *DefaultAction) SetBtnData(data interface{})  {}
func (def *DefaultAction) Js() template.JS              { return def.JS }
func (def *DefaultAction) BtnAttribute() template.HTML  { return def.Attr }
func (def *DefaultAction) BtnClass() template.HTML      { return "" }
func (def *DefaultAction) ExtContent() template.HTML    { return def.Ext }
func (def *DefaultAction) FooterContent() template.HTML { return def.Footer }
func (def *DefaultAction) GetCallbacks() context.Node   { return context.Node{} }

var _ Action = (*DefaultAction)(nil)

var DefaultPageSizeList = []int{ 10, 20, 30, 50, 100 }

const DefaultPageSize = 10

func NewInfoPanel(pk string) *InfoPanel {
	return &InfoPanel{
		curFieldListIndex:       -1,
		PageSizeList:            DefaultPageSizeList,
		DefaultPageSize:         DefaultPageSize,
		//processChains:           make(DisplayProcessFnChains, 0),
		//Buttons:                 make(Buttons, 0),
		//Callbacks:               make(Callbacks, 0),
		DisplayGeneratorRecords: make(map[string]struct{}),
		//Wheres:                  make([]Where, 0),
		//WhereRaws:               WhereRaw{},
		SortField:               pk,
		TableLayout:             "auto",
		FilterFormInputWidth:    10,
		FilterFormHeadWidth:     2,
		AutoRefresh:             0,
	}
}

func (i *InfoPanel) Where(field string, operator string, arg interface{}) *InfoPanel {
	i.Wheres = append(i.Wheres, Where{ Field: field, Operator: operator, Arg: arg, Join: "and" })
	return i
}

func (i *InfoPanel) WhereOr(field string, operator string, arg interface{}) *InfoPanel {
	i.Wheres = append(i.Wheres, Where{ Field: field, Operator: operator, Arg: arg, Join: "or" })
	return i
}

func (i *InfoPanel) WhereRaw(raw string, arg ...interface{}) *InfoPanel {
	i.WhereRaws.Raw = raw
	i.WhereRaws.Args = arg
	return i
}

func (i *InfoPanel) AddSelectBox(placeholder string, options FieldOptions, action Action, width ...int) *InfoPanel {
	options = append(FieldOptions{{ Value: "", Text: language.Get("All") }}, options...)
	action.SetBtnData(options)
	i.addButton(GetDefaultSelection(placeholder, options, action, width...)).
		addFooterHTML(action.FooterContent()).
		addCallback(action.GetCallbacks())
	return i
}

func (i *InfoPanel) ExportValue() *InfoPanel {
	i.ExportType = 1
	return i
}

func (i *InfoPanel) IsExportValue() bool {
	return i.ExportType == 1
}

func (i *InfoPanel) AddButtonRaw(btn Button, action Action) *InfoPanel {
	i.Buttons = append(i.Buttons, btn)
	i.addFooterHTML(action.FooterContent()).addCallback(action.GetCallbacks())
	return i
}

func (i *InfoPanel) AddButton(title template.HTML, icon string, action Action, color ...template.HTML) *InfoPanel {
	i.addButton(GetDefaultButtonGroup(title, icon, action, color...)).
		addFooterHTML(action.FooterContent()).
		addCallback(action.GetCallbacks())
	return i
}

func (i *InfoPanel) AddActionIconButton(icon string, action Action, ids ...string) *InfoPanel {
	i.addActionButton(GetActionIconButton(icon, action, ids...)).
		addFooterHTML(action.FooterContent()).
		addCallback(action.GetCallbacks())
	return i
}

func (i *InfoPanel) AddActionButtonFront(title template.HTML, action Action, ids ...string) *InfoPanel {
	i.SetActionButtonFold()
	i.ActionButtons = append([]Button{GetActionButton(title, action, ids...)}, i.ActionButtons...)
	i.addFooterHTML(action.FooterContent()).
		addCallback(action.GetCallbacks())
	return i
}

func (i *InfoPanel) AddActionButton(title template.HTML, action Action, ids ...string) *InfoPanel {
	i.SetActionButtonFold()
	i.addActionButton(GetActionButton(title, action, ids...)).
		addFooterHTML(action.FooterContent()).
		addCallback(action.GetCallbacks())
	return i
}

func (i *InfoPanel) SetActionButtonFold() *InfoPanel {
	i.ActionButtonFold = true
	return i
}

func (i *InfoPanel) AddLimitFilter(limit int) *InfoPanel {
	i.processChains = addLimit(limit, i.processChains)
	return i
}

func (i *InfoPanel) AddTrimSpaceFilter() *InfoPanel {
	i.processChains = addTrimSpace(i.processChains)
	return i
}

func (i *InfoPanel) AddSubstrFilter(start int, end int) *InfoPanel {
	i.processChains = addSubstr(start, end, i.processChains)
	return i
}

func (i *InfoPanel) AddToTitleFilter() *InfoPanel {
	i.processChains = addToTitle(i.processChains)
	return i
}

func (i *InfoPanel) AddToUpperFilter() *InfoPanel {
	i.processChains = addToUpper(i.processChains)
	return i
}

func (i *InfoPanel) AddToLowerFilter() *InfoPanel {
	i.processChains = addToLower(i.processChains)
	return i
}

func (i *InfoPanel) AddXssFilter() *InfoPanel {
	i.processChains = addXssFilter(i.processChains)
	return i
}

func (i *InfoPanel) AddXssJsFilter() *InfoPanel {
	i.processChains = addXssJsFilter(i.processChains)
	return i
}

func (i *InfoPanel) SetExportProcessFn(fn ExportProcessFn) *InfoPanel {
	i.ExportProcessFn = fn
	return i
}

func (i *InfoPanel) SetDeleteHook(fn DeleteFn) *InfoPanel {
	i.DeleteHook = fn
	return i
}

func (i *InfoPanel) SetDeleteHookWithRes(fn DeleteFnWithRes) *InfoPanel {
	i.DeleteHookWithRes = fn
	return i
}

func (i *InfoPanel) SetQueryFilterFn(fn QueryFilterFn) *InfoPanel {
	i.QueryFilterFn = fn
	return i
}

func (i *InfoPanel) AddUpdateParametersFn(fn UpdateParametersFn) *InfoPanel {
	i.UpdateParametersFns = append(i.UpdateParametersFns, fn)
	return i
}

func (i *InfoPanel) SetWrapper(wrapper ContentWrapper) *InfoPanel {
	i.Wrapper = wrapper
	return i
}

func (i *InfoPanel) SetPreDeleteFn(fn DeleteFn) *InfoPanel {
	i.PreDeleteFn = fn
	return i
}

func (i *InfoPanel) SetDeleteFn(fn DeleteFn) *InfoPanel {
	i.DeleteFn = fn
	return i
}

func (i *InfoPanel) SetGetDataFn(fn GetDataFn) *InfoPanel {
	i.GetDataFn = fn
	return i
}

func (i *InfoPanel) SetPrimaryKey(name string, typ db.DatabaseType) *InfoPanel {
	i.primaryKey = primaryKey{Name: name, Type: typ}
	return i
}

func (i *InfoPanel) SetTableFixed() *InfoPanel {
	i.TableLayout = "fixed"
	return i
}

func (i *InfoPanel) AddColumn(head string, fun FieldFilterFn) *InfoPanel {
	i.FieldList = append(i.FieldList, Field{
		Head:     head,
		Field:    head,
		TypeName: db.Varchar,
		Sortable: false,
		EditAble: false,
		EditType: table.Text,
		FieldDisplay: FieldDisplay{
			Display:              fun,
			DisplayProcessChains: chooseDisplayProcessChains(i.processChains),
		},
	})
	i.curFieldListIndex++
	return i
}

func (i *InfoPanel) AddColumnButtons(head string, buttons ...Button) *InfoPanel {
	const bLen = 1024
	var content, js, footer strings.Builder
	content.Grow(bLen)
	js     .Grow(bLen)
	footer .Grow(bLen)
	footer.WriteString(string(i.FooterHtml))
	for _, btn := range buttons {
		btn.GetAction().SetBtnId("." + btn.ID())
		btnContent, btnJs := btn.Content()
		content.WriteString(string(btnContent))
		js     .WriteString(string(btnJs))
		footer .WriteString(ParseTableDataTmpl(btn.GetAction().FooterContent()))
		i.Callbacks = i.Callbacks.AddCallback(btn.GetAction().GetCallbacks())
	}
	footer.WriteString(`<script>`)
	footer.WriteString(ParseTableDataTmpl(js.String()))
	footer.WriteString(`</script>`)
	i.FooterHtml = template.HTML(footer.String())
	contentStr  := content.String()
	i.FieldList  = append(i.FieldList, Field{
		Head:     head,
		Field:    head,
		TypeName: db.Varchar,
		Sortable: false,
		EditAble: false,
		EditType: table.Text,
		FieldDisplay: FieldDisplay{
			Display: func(value FieldModel) interface{} {
				pk := db.GetValueFromDatabaseType(i.primaryKey.Type, value.Row[i.primaryKey.Name], i.isFromJSON())
				v  := make(map[string]InfoItem, len(value.Row))
				for key, item := range value.Row {
					itemValue := fmt.Sprintf("%v", item)
					v[key] = InfoItem{ Value: itemValue, Content: template.HTML(itemValue) }
				}
				return template.HTML(ParseTableDataTmplWithID(pk.HTML(), contentStr, v))
			},
			DisplayProcessChains: chooseDisplayProcessChains(i.processChains),
		},
	})
	i.curFieldListIndex++
	return i
}

func (i *InfoPanel) AddFieldTr(ctx *context.Context, head, field string, typeName db.DatabaseType) *InfoPanel {
	return i.AddFieldWithTranslation(ctx, head, field, typeName)
}

func (i *InfoPanel) AddFieldWithTranslation(ctx *context.Context, head, field string, typeName db.DatabaseType) *InfoPanel {
	return i.AddField(language.GetWithLang(head, ctx.Lang()), field, typeName)
}

func (i *InfoPanel) AddField(head, field string, typeName db.DatabaseType) *InfoPanel {
	i.FieldList = append(i.FieldList, Field{
		Head:     head,
		Field:    field,
		TypeName: typeName,
		EditType: table.Text,
		FieldDisplay: FieldDisplay{
			Display:              NoopFieldDisplay,
			DisplayProcessChains: chooseDisplayProcessChains(i.processChains),
		},
	})
	i.curFieldListIndex++
	return i
}

func (i *InfoPanel) AddFilter(head, field string, typeName db.DatabaseType, fn UpdateParametersFn, filterType ...FilterType) *InfoPanel {
	return i.AddField(head, field, typeName).FieldHide().FieldFilterable(filterType...).AddUpdateParametersFn(fn)
}

// Field attribute setting functions
// ====================================================

func (i *InfoPanel) FieldDisplay(filter FieldFilterFn) *InfoPanel {
	i.FieldList[i.curFieldListIndex].Display = filter
	return i
}

type FieldLabelParam struct {
	Color template.HTML
	Type  string
}

func (i *InfoPanel) FieldLabel(args ...FieldLabelParam) *InfoPanel {
	i.addDisplayChains(displayFnGens["label"].Get(args))
	return i
}

func (i *InfoPanel) FieldImage(width, height string, prefix ...string) *InfoPanel {
	i.addDisplayChains(displayFnGens["image"].Get(width, height, prefix))
	return i
}

func (i *InfoPanel) FieldBool(flags ...string) *InfoPanel {
	i.addDisplayChains(displayFnGens["bool"].Get(flags))
	return i
}

func (i *InfoPanel) FieldLink(src string, openInNewTab ...bool) *InfoPanel {
	i.addDisplayChains(displayFnGens["link"].Get(src, openInNewTab))
	return i
}

func (i *InfoPanel) FieldFileSize() *InfoPanel {
	i.addDisplayChains(displayFnGens["filesize"].Get())
	return i
}

func (i *InfoPanel) FieldDate(format string) *InfoPanel {
	i.addDisplayChains(displayFnGens["date"].Get())
	return i
}

func (i *InfoPanel) FieldIcon(icons map[string]string, defaultIcon string) *InfoPanel {
	i.addDisplayChains(displayFnGens["link"].Get(icons, defaultIcon))
	return i
}

type FieldDotColor string

const (
	FieldDotColorDanger  FieldDotColor = "danger"
	FieldDotColorInfo    FieldDotColor = "info"
	FieldDotColorPrimary FieldDotColor = "primary"
	FieldDotColorSuccess FieldDotColor = "success"
)

func (i *InfoPanel) FieldDot(icons map[string]FieldDotColor, defaultDot FieldDotColor) *InfoPanel {
	i.addDisplayChains(displayFnGens["dot"].Get(icons, defaultDot))
	return i
}

type FieldProgressBarData struct {
	Style string
	Size  string
	Max   int
}

func (i *InfoPanel) FieldProgressBar(data ...FieldProgressBarData) *InfoPanel {
	i.addDisplayChains(displayFnGens["progressbar"].Get(data))
	return i
}

func (i *InfoPanel) FieldLoading(data []string) *InfoPanel {
	i.addDisplayChains(displayFnGens["loading"].Get(data))
	return i
}

func (i *InfoPanel) FieldDownLoadable(prefix ...string) *InfoPanel {
	i.addDisplayChains(displayFnGens["downloadable"].Get(prefix))
	return i
}

func (i *InfoPanel) FieldCopyable(prefix ...string) *InfoPanel {
	i.addDisplayChains(displayFnGens["copyable"].Get(prefix))
	if _, ok := i.DisplayGeneratorRecords["copyable"]; !ok {
		i.addFooterHTML(template.HTML(utils.StrConcat(`<script>`, string(displayFnGens["copyable"].JS()), `</script>`)))
		i.DisplayGeneratorRecords["copyable"] = struct{}{}
	}
	return i
}

type FieldGetImgArrFn func(value string) []string

func (i *InfoPanel) FieldCarousel(fn FieldGetImgArrFn, size ...int) *InfoPanel {
	i.addDisplayChains(displayFnGens["carousel"].Get(fn, size))
	return i
}

func (i *InfoPanel) FieldQrcode() *InfoPanel {
	i.addDisplayChains(displayFnGens["qrcode"].Get())
	if _, ok := i.DisplayGeneratorRecords["qrcode"]; !ok {
		i.addFooterHTML(template.HTML(utils.StrConcat(`<script>`, string(displayFnGens["qrcode"].JS()), `</script>`)))
		i.DisplayGeneratorRecords["qrcode"] = struct{}{}
	}
	return i
}

func (i *InfoPanel) FieldWidth(width int) *InfoPanel {
	i.FieldList[i.curFieldListIndex].Width = width
	return i
}

func (i *InfoPanel) FieldSortable() *InfoPanel {
	i.FieldList[i.curFieldListIndex].Sortable = true
	return i
}

func (i *InfoPanel) FieldEditOptions(options FieldOptions, extra ...map[string]string) *InfoPanel {
	if i.FieldList[i.curFieldListIndex].EditType.IsSwitch() {
		if len(extra) == 0 {
			options[0].Extra = map[string]string{
				"size"    : "small",
				"onColor" : "primary",
				"offColor": "default",
			}
		} else {
			m := extra[0]
			if m["size"    ] == "" { m["size"    ] = "small"   }
			if m["onColor" ] == "" { m["onColor" ] = "primary" }
			if m["offColor"] == "" { m["offColor"] = "default" }
			options[0].Extra = m
		}
	}
	i.FieldList[i.curFieldListIndex].EditOptions = options
	return i
}

func (i *InfoPanel) FieldEditAble(editType ...table.Type) *InfoPanel {
	i.FieldList[i.curFieldListIndex].EditAble = true
	if len(editType) > 0 {
		i.FieldList[i.curFieldListIndex].EditType = editType[0]
	}
	return i
}

func (i *InfoPanel) FieldAsEditParam() *InfoPanel {
	i.FieldList[i.curFieldListIndex].IsEditParam = true
	return i
}

func (i *InfoPanel) FieldAsDeleteParam() *InfoPanel {
	i.FieldList[i.curFieldListIndex].IsDeleteParam = true
	return i
}

func (i *InfoPanel) FieldAsDetailParam() *InfoPanel {
	i.FieldList[i.curFieldListIndex].IsDetailParam = true
	return i
}

func (i *InfoPanel) FieldFixed() *InfoPanel {
	i.FieldList[i.curFieldListIndex].Fixed = true
	return i
}

type FilterType struct {
	Options     FieldOptions
	Process     func(string) string
	OptionExt   map[string]interface{}
	FormType    form.Type
	HelpMsg     template.HTML
	Style       template.HTMLAttr
	Operator    FilterOperator
	Head        string
	Placeholder string
	Width       int
	HeadWidth   int
	InputWidth  int
	NoHead      bool
	NoIcon      bool
}

func (i *InfoPanel) FieldFilterable(filterType ...FilterType) *InfoPanel {
	curField := i.FieldList[i.curFieldListIndex]
	curField.Filterable = true
	langInput := language.Get("input") + " "

	if len(filterType) == 0 {
		curField.FilterFormFields = append(curField.FilterFormFields,
			FilterFormField{
				Type:        form.Text,
				Head:        curField.Head,
				Placeholder: langInput + curField.Head,
			})
	} else {
		for _, filter := range filterType {
			ff := FilterFormField{
				Width     : filter.Width,
				HeadWidth : filter.HeadWidth,
				InputWidth: filter.InputWidth,
				HelpMsg   : filter.HelpMsg,
				NoIcon    : filter.NoIcon,
				Style     : filter.Style,
				ProcessFn : filter.Process,
				Operator  : filter.Operator,
				Options   : filter.Options,
			}
			if filter.FormType == form.Default {
				ff.Type = form.Text
			} else {
				ff.Type = filter.FormType
			}
			if !filter.NoHead && filter.Head == "" {
				ff.Head = curField.Head
			} else {
				ff.Head = filter.Head
			}
			//ff.Head = modules.AorB(!filter.NoHead && filter.Head == "", i.FieldList[i.curFieldListIndex].Head, filter.Head)
			if filter.Placeholder == "" {
				ff.Placeholder = langInput + ff.Head
			} else {
				ff.Placeholder = filter.Placeholder
			}
			//ff.Placeholder = modules.AorB(filter.Placeholder == "", language.MustGet("input")+" "+ff.Head, filter.Placeholder)
			if len(filter.OptionExt) > 0 {
				s, _ := utils.JsonMarshal(filter.OptionExt)
				ff.OptionExt = template.JS(s)
			}
			curField.FilterFormFields = append(curField.FilterFormFields, ff)
		}
	}

	i.FieldList[i.curFieldListIndex] = curField
	return i
}

func (i *InfoPanel) FieldFilterOptions(options FieldOptions) *InfoPanel {
	i.FieldList[i.curFieldListIndex].FilterFormFields[0].Options = options
	i.FieldList[i.curFieldListIndex].FilterFormFields[0].OptionExt = `{"allowClear": "true"}`
	return i
}

func (i *InfoPanel) FieldFilterOptionsFromTable(table, textFieldName, valueFieldName string, process ...OptionTableQueryProcessFn) *InfoPanel {
	var fn OptionTableQueryProcessFn
	if len(process) > 0 { fn = process[0] }
	i.FieldList[i.curFieldListIndex].FilterFormFields[0].OptionTable = OptionTable{
		Table:          table,
		TextField:      textFieldName,
		ValueField:     valueFieldName,
		QueryProcessFn: fn,
	}
	return i
}

func (i *InfoPanel) FieldFilterProcess(process func(string) string) *InfoPanel {
	i.FieldList[i.curFieldListIndex].FilterFormFields[0].ProcessFn = process
	return i
}

func (i *InfoPanel) FieldFilterOptionExt(m map[string]interface{}) *InfoPanel {
	s, _ := utils.JsonMarshal(m)
	i.FieldList[i.curFieldListIndex].FilterFormFields[0].OptionExt = template.JS(s)
	return i
}

func (i *InfoPanel) FieldFilterOnSearch(url string, handler Handler, delay ...int) *InfoPanel {
	ext, callback := searchJS(i.FieldList[i.curFieldListIndex].FilterFormFields[0].OptionExt, url, handler, delay...)
	i.FieldList[i.curFieldListIndex].FilterFormFields[0].OptionExt = ext
	i.Callbacks = append(i.Callbacks, callback)
	return i
}

func (i *InfoPanel) FieldFilterOnChooseCustom(js template.HTML) *InfoPanel {
	i.FooterHtml += chooseCustomJS(i.FieldList[i.curFieldListIndex].Field, js)
	return i
}

func (i *InfoPanel) FieldFilterOnChooseMap(m map[string]LinkField) *InfoPanel {
	i.FooterHtml += chooseMapJS(i.FieldList[i.curFieldListIndex].Field, m)
	return i
}

func (i *InfoPanel) FieldFilterOnChoose(val, field string, value template.HTML) *InfoPanel {
	i.FooterHtml += chooseJS(i.FieldList[i.curFieldListIndex].Field, field, val, value)
	return i
}

func (i *InfoPanel) OperationURL(id string) string {
	return config.Url("/operation/" + utils.WrapURL(id))
}

func (i *InfoPanel) FieldFilterOnChooseAjax(field, url string, handler Handler) *InfoPanel {
	js, callback := chooseAjax(i.FieldList[i.curFieldListIndex].Field, field, i.OperationURL(url), handler)
	i.FooterHtml += js
	i.Callbacks = append(i.Callbacks, callback)
	return i
}

func (i *InfoPanel) FieldFilterOnChooseHide(value string, field ...string) *InfoPanel {
	i.FooterHtml += chooseHideJS(i.FieldList[i.curFieldListIndex].Field, []string{value}, field...)
	return i
}

func (i *InfoPanel) FieldFilterOnChooseShow(value string, field ...string) *InfoPanel {
	i.FooterHtml += chooseShowJS(i.FieldList[i.curFieldListIndex].Field, []string{value}, field...)
	return i
}

func (i *InfoPanel) FieldFilterOnChooseDisable(value string, field ...string) *InfoPanel {
	i.FooterHtml += chooseDisableJS(i.FieldList[i.curFieldListIndex].Field, []string{value}, field...)
	return i
}

func (i *InfoPanel) FieldHide() *InfoPanel {
	i.FieldList[i.curFieldListIndex].Hide = true
	return i
}

func (i *InfoPanel) FieldJoin(join Join) *InfoPanel {
	i.FieldList[i.curFieldListIndex].Joins = append(i.FieldList[i.curFieldListIndex].Joins, join)
	return i
}

func (i *InfoPanel) FieldLimit(limit int) *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].AddLimit(limit)
	return i
}

func (i *InfoPanel) FieldTrimSpace() *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].AddTrimSpace()
	return i
}

func (i *InfoPanel) FieldSubstr(start int, end int) *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].AddSubstr(start, end)
	return i
}

func (i *InfoPanel) FieldToTitle() *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].AddToTitle()
	return i
}

func (i *InfoPanel) FieldToUpper() *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].AddToUpper()
	return i
}

func (i *InfoPanel) FieldToLower() *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].AddToLower()
	return i
}

func (i *InfoPanel) FieldXssFilter() *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].DisplayProcessChains.Add(xssFilter)
	return i
}

// InfoPanel attribute setting functions
// ====================================================

func (i *InfoPanel) SetTable(table string) *InfoPanel {
	i.Table = table
	return i
}

func (i *InfoPanel) SetPageSizeList(pageSizeList []int) *InfoPanel {
	i.PageSizeList = pageSizeList
	return i
}

func (i *InfoPanel) SetDefaultPageSize(defaultPageSize int) *InfoPanel {
	i.DefaultPageSize = defaultPageSize
	return i
}

func (i *InfoPanel) GetPageSizeList() []string {
	pageSizeList := make([]string, len(i.PageSizeList))
	for j, ps := range i.PageSizeList {
		pageSizeList[j] = strconv.Itoa(ps)
	}
	return pageSizeList
}

func (i *InfoPanel) GetSort() string {
	switch i.Sort {
	case SortAsc: return "asc"
	default     : return "desc"
	}
}

func (i *InfoPanel) SetTitle(title string) *InfoPanel {
	i.Title = title
	return i
}

func (i *InfoPanel) SetTabGroups(groups TabGroups) *InfoPanel {
	i.TabGroups = groups
	return i
}

func (i *InfoPanel) SetTabHeaders(headers ...string) *InfoPanel {
	i.TabHeaders = headers
	return i
}

func (i *InfoPanel) SetDescription(desc string) *InfoPanel {
	i.Description = desc
	return i
}

func (i *InfoPanel) SetFilterFormLayout(layout form.Layout) *InfoPanel {
	i.FilterFormLayout = layout
	return i
}

func (i *InfoPanel) SetFilterFormHeadWidth(w int) *InfoPanel {
	i.FilterFormHeadWidth = w
	return i
}

func (i *InfoPanel) SetFilterFormInputWidth(w int) *InfoPanel {
	i.FilterFormInputWidth = w
	return i
}

func (i *InfoPanel) SetSortField(field string) *InfoPanel {
	i.SortField = field
	return i
}

func (i *InfoPanel) SetSortAsc() *InfoPanel {
	i.Sort = SortAsc
	return i
}

func (i *InfoPanel) SetSortDesc() *InfoPanel {
	i.Sort = SortDesc
	return i
}

func (i *InfoPanel) SetAction(action template.HTML) *InfoPanel {
	i.Action = action
	return i
}

func (i *InfoPanel) SetHeaderHtml(header template.HTML) *InfoPanel {
	i.HeaderHtml += header
	return i
}

func (i *InfoPanel) SetFooterHtml(footer template.HTML) *InfoPanel {
	i.FooterHtml += footer
	return i
}

func (i *InfoPanel) HasError() bool {
	return i.PageError != nil
}

func (i *InfoPanel) SetError(err errors.PageError, content ...template.HTML) *InfoPanel {
	i.PageError = err
	if len(content) > 0 {
		i.PageErrorHTML = content[0]
	}
	return i
}

func (i *InfoPanel) SetNoCompress() *InfoPanel {
	i.NoCompress = true
	return i
}

func (i *InfoPanel) SetHideSideBar() *InfoPanel {
	i.HideSideBar = true
	return i
}

func (i *InfoPanel) SetAutoRefresh(interval uint) *InfoPanel {
	i.AutoRefresh = interval
	return i
}

func (i *InfoPanel) Set404Error(content ...template.HTML) *InfoPanel {
	i.SetError(errors.PageError404, content...)
	return i
}

func (i *InfoPanel) Set403Error(content ...template.HTML) *InfoPanel {
	i.SetError(errors.PageError403, content...)
	return i
}

func (i *InfoPanel) Set400Error(content ...template.HTML) *InfoPanel {
	i.SetError(errors.PageError401, content...)
	return i
}

func (i *InfoPanel) Set500Error(content ...template.HTML) *InfoPanel {
	i.SetError(errors.PageError500, content...)
	return i
}

func (i *InfoPanel) HideNewButton() *InfoPanel {
	i.IsHideNewButton = true
	return i
}

func (i *InfoPanel) HideExportButton() *InfoPanel {
	i.IsHideExportButton = true
	return i
}

func (i *InfoPanel) HideFilterButton() *InfoPanel {
	i.IsHideFilterButton = true
	return i
}

func (i *InfoPanel) HideRowSelector() *InfoPanel {
	i.IsHideRowSelector = true
	return i
}

func (i *InfoPanel) HidePagination() *InfoPanel {
	i.IsHidePagination = true
	return i
}

func (i *InfoPanel) HideFilterArea() *InfoPanel {
	i.IsHideFilterArea = true
	return i
}

func (i *InfoPanel) HideQueryInfo() *InfoPanel {
	i.IsHideQueryInfo = true
	return i
}

func (i *InfoPanel) HideEditButton() *InfoPanel {
	i.IsHideEditButton = true
	return i
}

func (i *InfoPanel) HideDeleteButton() *InfoPanel {
	i.IsHideDeleteButton = true
	return i
}

func (i *InfoPanel) HideDetailButton() *InfoPanel {
	i.IsHideDetailButton = true
	return i
}

func (i *InfoPanel) HideCheckBoxColumn() *InfoPanel {
	return i.HideColumn(1)
}

func (i *InfoPanel) HideColumn(n int) *InfoPanel {
	i.AddCSS(template.CSS(fmt.Sprintf(`
	.box-body table.table tbody tr td:nth-child(%v), .box-body table.table tbody tr th:nth-child(%v) { display: none }`, n, n)))
	return i
}

func (i *InfoPanel) addFooterHTML(footer template.HTML) *InfoPanel {
	i.FooterHtml += template.HTML(ParseTableDataTmpl(footer))
	return i
}

func (i *InfoPanel) AddCSS(css template.CSS) *InfoPanel {
	return i.addFooterHTML(template.HTML(utils.StrConcat("<style>", string(css), "</style>")))
}

func (i *InfoPanel) AddJS(js template.JS) *InfoPanel {
	return i.addFooterHTML(template.HTML(utils.StrConcat("<script>", string(js), "</script>")))
}

func (i *InfoPanel) addCallback(node context.Node) *InfoPanel {
	i.Callbacks = i.Callbacks.AddCallback(node)
	return i
}

func (i *InfoPanel) addButton(btn Button) *InfoPanel {
	i.Buttons = append(i.Buttons, btn)
	return i
}

func (i *InfoPanel) addActionButton(btn Button) *InfoPanel {
	i.ActionButtons = append(i.ActionButtons, btn)
	return i
}

func (i *InfoPanel) isFromJSON() bool {
	return i.GetDataFn != nil
}

func (i *InfoPanel) addDisplayChains(fn FieldFilterFn) *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].DisplayProcessChains.Add(fn)
	return i
}
