package types

import (
	"fmt"
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/file"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	form2 "github.com/GoAdminGroup/go-admin/template/types/form"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type FieldOption struct {
	Text          string            `json:"text"`
	Value         string            `json:"value"`
	TextHTML      template.HTML     `json:"-"`
	Selected      bool              `json:"-"`
	SelectedLabel template.HTML     `json:"-"`
	Extra         map[string]string `json:"-"`
}

type FieldOptions []FieldOption

func (fo FieldOptions) Copy() FieldOptions {
	newOptions := make(FieldOptions, len(fo))
	copy(newOptions, fo)
	return newOptions
}

func (fo FieldOptions) SetSelected(val interface{}, labels []template.HTML) FieldOptions {
	label2, label1 := labels[1], labels[0]
	if valArr, ok := val.([]string); ok {
		for i, o := range fo {
			text := o.Text
			if text == "" { text = string(o.TextHTML) }
			o.Selected = utils.InArray(valArr, o.Value) || utils.InArray(valArr, text)
			if o.Selected {
				o.SelectedLabel = label1
			} else {
				o.SelectedLabel = label2
			}
			fo[i] = o
		}
	} else {
		for i, o := range fo {
			text := o.Text
			if text == "" { text = string(o.TextHTML) }
			o.Selected = o.Value == val || text == val
			if o.Selected {
				o.SelectedLabel = label1
			} else {
				o.SelectedLabel = label2
			}
			fo[i] = o
		}
	}
	return fo
}

func (fo FieldOptions) SetSelectedLabel(labels []template.HTML) FieldOptions {
	label2, label1 := labels[1], labels[0]
	for i, o := range fo {
		if o.Selected {
			fo[i].SelectedLabel = label1
		} else {
			fo[i].SelectedLabel = label2
		}
	}
	return fo
}

func (fo FieldOptions) Marshal() string {
	if len(fo) == 0 { return "" }
	eo, err := utils.JsonMarshal(fo)
	if err != nil { return "" }
	return string(eo)
}

type (
	OptionInitFn              func(val FieldModel) FieldOptions
	OptionArrInitFn           func(val FieldModel) []FieldOptions
	OptionTableQueryProcessFn func(sql *db.SQL) *db.SQL
	OptionProcessFn           func(options FieldOptions) FieldOptions

	OptionTable struct {
		Table          string
		TextField      string
		ValueField     string
		QueryProcessFn OptionTableQueryProcessFn
		ProcessFn      OptionProcessFn
	}
)

// FormField is the form field with different options.
type FormField struct {
	Field          string          `json:"field"`
	FieldClass     string          `json:"field_class"`
	TypeName       db.DatabaseType `json:"type_name"`
	Head           string          `json:"head"`
	Foot           template.HTML   `json:"foot"`
	FormType       form2.Type      `json:"form_type"`
	FatherFormType form2.Type      `json:"father_form_type"`
	FatherField    string          `json:"father_field"`

	RowWidth int
	RowFlag  uint8

	Default                template.HTML  `json:"default"`
	DefaultArr             interface{}    `json:"default_arr"`
	Value                  template.HTML  `json:"value"`
	Value2                 string         `json:"value_2"`
	ValueArr               []string       `json:"value_arr"`
	Value2Arr              []string       `json:"value_2_arr"`
	Options                FieldOptions   `json:"options"`
	OptionsArr             []FieldOptions `json:"options_arr"`
	DefaultOptionDelimiter string         `json:"default_option_delimiter"`
	Label                  template.HTML  `json:"label"`
	HideLabel              bool           `json:"hide_label"`

	Placeholder string `json:"placeholder"`

	CustomContent template.HTML `json:"custom_content"`
	CustomJs      template.JS   `json:"custom_js"`
	CustomCss     template.CSS  `json:"custom_css"`

	Editable         bool `json:"editable"`
	NotAllowEdit     bool `json:"not_allow_edit"`
	NotAllowAdd      bool `json:"not_allow_add"`
	DisplayButNotAdd bool `json:"display_but_not_add"`
	Must             bool `json:"must"`
	Hide             bool `json:"hide"`
	CreateHide       bool `json:"create_hide"`
	EditHide         bool `json:"edit_hide"`

	Width int `json:"width"`

	InputWidth int `json:"input_width"`
	HeadWidth  int `json:"head_width"`

	Joins Joins `json:"-"`

	Divider      bool   `json:"divider"`
	DividerTitle string `json:"divider_title"`

	HelpMsg template.HTML `json:"help_msg"`

	TableFields FormFields

	Style  template.HTMLAttr `json:"style"`
	NoIcon bool              `json:"no_icon"`

	OptionExt       template.JS     `json:"option_ext"`
	OptionExt2      template.JS     `json:"option_ext_2"`
	OptionInitFn    OptionInitFn    `json:"-"`
	OptionArrInitFn OptionArrInitFn `json:"-"`
	OptionTable     OptionTable     `json:"-"`

	FieldDisplay `json:"-"`
	PostFilterFn PostFieldFilterFn `json:"-"`
}

func (f *FormField) GetRawValue(columnMap map[string]struct{}, v interface{}) string {
	isJSON := len(columnMap) == 0
	if isJSON || utils.InMapT(columnMap, f.Field) {
		return db.GetValueFromDatabaseType(f.TypeName, v, isJSON).String()
	}
	return ""
}

func (f *FormField) UpdateValue(id, val string, res map[string]interface{}, sql *db.SQL) *FormField {
	return f.updateValue(id, val, res, PostTypeUpdate, sql)
}

func (f *FormField) UpdateDefaultValue(sql *db.SQL) *FormField {
	f.Value = f.Default
	return f.updateValue("", string(f.Value), make(map[string]interface{}), PostTypeCreate, sql)
}

func (f *FormField) setOptionsFromSQL(sql *db.SQL) {
	if sql != nil && f.OptionTable.Table != "" && len(f.Options) == 0 {
		sql.Table(f.OptionTable.Table).Select(f.OptionTable.ValueField, f.OptionTable.TextField)

		if f.OptionTable.QueryProcessFn != nil {
			f.OptionTable.QueryProcessFn(sql)
		}

		queryRes, err := sql.All()
		if err == nil {
			for _, item := range queryRes {
				f.Options = append(f.Options, FieldOption{
					Value: fmt.Sprintf("%v", item[f.OptionTable.ValueField]),
					Text:  fmt.Sprintf("%v", item[f.OptionTable.TextField]),
				})
			}
		}

		if f.OptionTable.ProcessFn != nil {
			f.Options = f.OptionTable.ProcessFn(f.Options)
		}
	}
}

func (f *FormField) isBelongToATable() bool {
	return f.FatherField != "" && f.FatherFormType.IsTable()
}

func (f *FormField) isNotBelongToATable() bool {
	return f.FatherField == "" && !f.FatherFormType.IsTable()
}

func (f *FormField) allowAdd() bool {
	return !f.NotAllowAdd
}

func (f *FormField) updateValue(id, val string, row map[string]interface{}, typ PostType, sql *db.SQL) *FormField {
	m := FieldModel{
		ID:       id,
		Value:    val,
		Row:      row,
		PostType: typ,
	}

	if f.isBelongToATable() {
		if f.FormType.IsSelect() {
			selectedLabels := f.FormType.SelectedLabel()
			if len(f.OptionsArr) == 0 && f.OptionArrInitFn != nil {
				f.OptionsArr = f.OptionArrInitFn(m)
				for i, o := range f.OptionsArr {
					f.OptionsArr[i] = o.SetSelectedLabel(selectedLabels)
				}
			} else {
				f.setOptionsFromSQL(sql)
				if f.FormType.IsSingleSelect() {
					values := f.ToDisplayStringArray(m)
					f.OptionsArr = make([]FieldOptions, len(values))
					for i, value := range values {
						f.OptionsArr[i] = f.Options.Copy().SetSelected(value, selectedLabels)
					}
				} else {
					values := f.ToDisplayStringArrayArray(m)
					f.OptionsArr = make([]FieldOptions, len(values))
					for i, value := range values {
						f.OptionsArr[i] = f.Options.Copy().SetSelected(value, selectedLabels)
					}
				}
			}
		} else {
			f.ValueArr = f.ToDisplayStringArray(m)
		}
	} else {
		if f.FormType.IsSelect() {
			if len(f.Options) == 0 && f.OptionInitFn != nil {
				f.Options = f.OptionInitFn(m).SetSelectedLabel(f.FormType.SelectedLabel())
			} else {
				f.setOptionsFromSQL(sql)
				f.Options.SetSelected(f.ToDisplay(m), f.FormType.SelectedLabel())
			}
		} else if f.FormType.IsArray() {
			f.ValueArr = f.ToDisplayStringArray(m)
		} else {
			f.Value = f.ToDisplayHTML(m)
			if f.FormType.IsFile() {
				if f.Value != "" {
					f.Value2 = config.GetStore().URL(string(f.Value))
				}
			}
		}
	}

	return f
}

func (f *FormField) FillCustomContent() *FormField {
	// TODO: optimize
	if f.CustomContent != "" {
		f.CustomContent = template.HTML(f.fillCustom(string(f.CustomContent)))
	}
	if f.CustomJs != "" {
		f.CustomJs = template.JS(f.fillCustom(string(f.CustomJs)))
	}
	if f.CustomCss != "" {
		f.CustomCss = template.CSS(f.fillCustom(string(f.CustomCss)))
	}
	return f
}

func (f *FormField) fillCustom(src string) string {
	t := template.New("custom")
	t, err := t.Parse(src)
	if err != nil {
		logger.Error(err)
		return ""
	}
	var sb strings.Builder
	err = t.Execute(&sb, f)
	if err != nil {
		logger.Error(err)
		return ""
	}
	return sb.String()
}

// FormPanel
type FormPanel struct {
	FieldList         FormFields `json:"field_list"`
	curFieldListIndex int

	// Warn: may be deprecated in the future. `json:""
	TabGroups  TabGroups  `json:"tab_groups"`
	TabHeaders TabHeaders `json:"tab_headers"`

	Table       string `json:"table"`
	Title       string `json:"title"`
	Description string `json:"description"`

	Validator    FormPostFn       `json:"validator"`
	PostHook     FormPostFn       `json:"post_hook"`
	PreProcessFn FormPreProcessFn `json:"pre_process_fn"`

	Callbacks Callbacks `json:"callbacks"`

	primaryKey primaryKey

	UpdateFn FormPostFn `json:"update_fn"`
	InsertFn FormPostFn `json:"insert_fn"`

	IsHideContinueEditCheckBox bool `json:"is_hide_continue_edit_check_box"`
	IsHideContinueNewCheckBox  bool `json:"is_hide_continue_new_check_box"`
	IsHideResetButton          bool `json:"is_hide_reset_button"`
	IsHideBackButton           bool `json:"is_hide_back_button"`

	Layout form2.Layout `json:"layout"`

	HTMLContent template.HTML `json:"html_content"`

	Header template.HTML `json:"header"`

	InputWidth int `json:"input_width"`
	HeadWidth  int `json:"head_width"`

	FormNewTitle    template.HTML `json:"form_new_title"`
	FormNewBtnWord  template.HTML `json:"form_new_btn_word"`
	FormEditTitle   template.HTML `json:"form_edit_title"`
	FormEditBtnWord template.HTML `json:"form_edit_btn_word"`

	Ajax          bool        `json:"ajax"`
	AjaxSuccessJS template.JS `json:"ajax_success_js"`
	AjaxErrorJS   template.JS `json:"ajax_error_js"`

	Responder Responder `json:"responder"`

	Wrapper ContentWrapper `json:"wrapper"`

	HideSideBar bool `json:"hide_side_bar"`

	processChains DisplayProcessFnChains

	HeaderHtml template.HTML `json:"header_html"`
	FooterHtml template.HTML `json:"footer_html"`

	PageError     errors.PageError `json:"page_error"`
	PageErrorHTML template.HTML    `json:"page_error_html"`

	NoCompress bool `json:"no_compress"`
}

type Responder func(ctx *context.Context)

func NewFormPanel() *FormPanel {
	return &FormPanel{
		curFieldListIndex: -1,
		Callbacks:         make(Callbacks, 0),
		Layout:            form2.LayoutDefault,
		FormNewTitle:      "New",
		FormEditTitle:     "Edit",
		FormNewBtnWord:    language.GetFromHtml("Save"),
		FormEditBtnWord:   language.GetFromHtml("Save"),
	}
}

func (f *FormPanel) AddLimitFilter(limit int) *FormPanel {
	f.processChains = addLimit(limit, f.processChains)
	return f
}

func (f *FormPanel) AddTrimSpaceFilter() *FormPanel {
	f.processChains = addTrimSpace(f.processChains)
	return f
}

func (f *FormPanel) AddSubstrFilter(start int, end int) *FormPanel {
	f.processChains = addSubstr(start, end, f.processChains)
	return f
}

func (f *FormPanel) AddToTitleFilter() *FormPanel {
	f.processChains = addToTitle(f.processChains)
	return f
}

func (f *FormPanel) AddToUpperFilter() *FormPanel {
	f.processChains = addToUpper(f.processChains)
	return f
}

func (f *FormPanel) AddToLowerFilter() *FormPanel {
	f.processChains = addToLower(f.processChains)
	return f
}

func (f *FormPanel) AddXssFilter() *FormPanel {
	f.processChains = addXssFilter(f.processChains)
	return f
}

func (f *FormPanel) AddXssJsFilter() *FormPanel {
	f.processChains = addXssJsFilter(f.processChains)
	return f
}

func (f *FormPanel) SetPrimaryKey(name string, typ db.DatabaseType) *FormPanel {
	f.primaryKey = primaryKey{Name: name, Type: typ}
	return f
}

func (f *FormPanel) HideContinueEditCheckBox() *FormPanel {
	f.IsHideContinueEditCheckBox = true
	return f
}

func (f *FormPanel) HideContinueNewCheckBox() *FormPanel {
	f.IsHideContinueNewCheckBox = true
	return f
}

func (f *FormPanel) HideResetButton() *FormPanel {
	f.IsHideResetButton = true
	return f
}

func (f *FormPanel) HideBackButton() *FormPanel {
	f.IsHideBackButton = true
	return f
}

func (f *FormPanel) AddFieldTr(ctx *context.Context, head, field string, filedType db.DatabaseType, formType form2.Type) *FormPanel {
	return f.AddFieldWithTranslation(ctx, head, field, filedType, formType)
}

func (f *FormPanel) AddFieldWithTranslation(ctx *context.Context, head, field string, filedType db.DatabaseType,
	formType form2.Type) *FormPanel {
	return f.AddField(language.GetWithLang(head, ctx.Lang()), field, filedType, formType)
}

func (f *FormPanel) AddField(head, field string, filedType db.DatabaseType, formType form2.Type) *FormPanel {
	f.FieldList = append(f.FieldList, FormField{
		Head:        head,
		Field:       field,
		FieldClass:  field,
		TypeName:    filedType,
		Editable:    true,
		Hide:        false,
		TableFields: make(FormFields, 0),
		Placeholder: language.Get("input") + " " + head,
		FormType:    formType,
		FieldDisplay: FieldDisplay{
			Display:              NoopFieldDisplay,
			DisplayProcessChains: chooseDisplayProcessChains(f.processChains),
		},
	})
	f.curFieldListIndex++

	// Set default options of different form type
	op1, op2, js := formType.GetDefaultOptions(field)
	f.FieldOptionExt(op1)
	f.FieldOptionExt2(op2)
	f.FieldOptionExtJS(js)

	// Set default Display Filter Function of different form type
	setDefaultDisplayFnOfFormType(f, formType)

	if formType.IsEditor() {
		f.NoCompress = true
	}

	return f
}

type AddFormFieldFn func(panel *FormPanel)

func (f *FormPanel) AddTable(head, field string, addFields AddFormFieldFn) *FormPanel {
	index := f.curFieldListIndex
	addFields(f)

	for i := index + 1; i <= f.curFieldListIndex; i++ {
		f.FieldList[i].FatherFormType = form2.Table
		f.FieldList[i].FatherField = field
	}

	fields := make(FormFields, f.curFieldListIndex - index)
	copy(fields, f.FieldList[index+1:f.curFieldListIndex+1])

	f.FieldList = append(f.FieldList, FormField{
		Head:        head,
		Field:       field,
		FieldClass:  field,
		TypeName:    db.Varchar,
		Editable:    true,
		Hide:        false,
		TableFields: fields,
		FormType:    form2.Table,
		FieldDisplay: FieldDisplay{
			Display:              NoopFieldDisplay,
			DisplayProcessChains: chooseDisplayProcessChains(f.processChains),
		},
	})

	f.curFieldListIndex++
	return f
}

func (f *FormPanel) AddRow(addFields AddFormFieldFn) *FormPanel {
	index := f.curFieldListIndex + 1
	addFields(f)
	if f.curFieldListIndex != index {
		for i := index; i <= f.curFieldListIndex; i++ {
			switch i {
			case index              : f.FieldList[i].RowFlag = 1
			case f.curFieldListIndex: f.FieldList[i].RowFlag = 2
			default                 : f.FieldList[i].RowFlag = 3
			}
		}
	}
	return f
}

// Field attribute setting functions
// ====================================================

func (f *FormPanel) FieldDisplay(filter FieldFilterFn) *FormPanel {
	f.FieldList[f.curFieldListIndex].Display = filter
	return f
}

func (f *FormPanel) SetTable(table string) *FormPanel {
	f.Table = table
	return f
}

func (f *FormPanel) FieldMust() *FormPanel {
	f.FieldList[f.curFieldListIndex].Must = true
	return f
}

func (f *FormPanel) FieldHide() *FormPanel {
	f.FieldList[f.curFieldListIndex].Hide = true
	return f
}

func (f *FormPanel) FieldPlaceholder(placeholder string) *FormPanel {
	f.FieldList[f.curFieldListIndex].Placeholder = placeholder
	return f
}

func (f *FormPanel) FieldWidth(width int) *FormPanel {
	f.FieldList[f.curFieldListIndex].Width = width
	return f
}

func (f *FormPanel) FieldInputWidth(width int) *FormPanel {
	f.FieldList[f.curFieldListIndex].InputWidth = width
	return f
}

func (f *FormPanel) FieldHeadWidth(width int) *FormPanel {
	f.FieldList[f.curFieldListIndex].HeadWidth = width
	return f
}

func (f *FormPanel) FieldRowWidth(width int) *FormPanel {
	f.FieldList[f.curFieldListIndex].RowWidth = width
	return f
}

func (f *FormPanel) FieldHideLabel() *FormPanel {
	f.FieldList[f.curFieldListIndex].HideLabel = true
	return f
}

func (f *FormPanel) FieldFoot(foot template.HTML) *FormPanel {
	f.FieldList[f.curFieldListIndex].Foot = foot
	return f
}

func (f *FormPanel) FieldDivider(title ...string) *FormPanel {
	f.FieldList[f.curFieldListIndex].Divider = true
	if len(title) > 0 {
		f.FieldList[f.curFieldListIndex].DividerTitle = title[0]
	}
	return f
}

func (f *FormPanel) FieldHelpMsg(s template.HTML) *FormPanel {
	f.FieldList[f.curFieldListIndex].HelpMsg = s
	return f
}

func (f *FormPanel) FieldOptionInitFn(fn OptionInitFn) *FormPanel {
	f.FieldList[f.curFieldListIndex].OptionInitFn = fn
	return f
}

func (f *FormPanel) FieldOptionExt(m map[string]interface{}) *FormPanel {
	if m == nil { return f }

	if f.FieldList[f.curFieldListIndex].FormType.IsCode() {
		f.FieldList[f.curFieldListIndex].OptionExt = template.JS(fmt.Sprintf(`
	theme = "%s";
	font_size = %s;
	language = "%s";
	options = %s;
`, m["theme"], m["font_size"], m["language"], m["options"]))
		return f
	}

	m = f.FieldList[f.curFieldListIndex].FormType.FixOptions(m)
	s, _ := utils.JsonMarshal(m)

	if f.FieldList[f.curFieldListIndex].OptionExt != "" {
		ss := string(f.FieldList[f.curFieldListIndex].OptionExt)
		ss = strings.Replace(ss, "}", "", strings.Count(ss, "}"))
		ss = strings.TrimRight(ss, " ")
		ss += ","
		f.FieldList[f.curFieldListIndex].OptionExt = template.JS(ss + strings.Replace(string(s), "{", "", 1))
	} else {
		f.FieldList[f.curFieldListIndex].OptionExt = template.JS(s)
	}

	return f
}

func (f *FormPanel) FieldOptionExt2(m map[string]interface{}) *FormPanel {
	if m == nil { return f }

	m = f.FieldList[f.curFieldListIndex].FormType.FixOptions(m)
	s, _ := utils.JsonMarshal(m)

	if f.FieldList[f.curFieldListIndex].OptionExt2 != "" {
		ss := string(f.FieldList[f.curFieldListIndex].OptionExt2)
		ss = strings.Replace(ss, "}", "", strings.Count(ss, "}"))
		ss = strings.TrimRight(ss, " ")
		ss += ","
		f.FieldList[f.curFieldListIndex].OptionExt2 = template.JS(ss + strings.Replace(string(s), "{", "", 1))
	} else {
		f.FieldList[f.curFieldListIndex].OptionExt2 = template.JS(s)
	}

	return f
}

func (f *FormPanel) FieldOptionExtJS(js template.JS) *FormPanel {
	if js != "" {
		f.FieldList[f.curFieldListIndex].OptionExt = js
	}
	return f
}

func (f *FormPanel) FieldOptionExtJS2(js template.JS) *FormPanel {
	f.FieldList[f.curFieldListIndex].OptionExt2 = js
	return f
}

func (f *FormPanel) FieldEnableFileUpload(data ...interface{}) *FormPanel {
	var url string
	if len(data) > 0 {
		url = data[0].(string)
	} else {
		url = f.OperationURL("/file/upload")
	}

	field := f.FieldList[f.curFieldListIndex].Field

	f.FieldList[f.curFieldListIndex].OptionExt = template.JS(fmt.Sprintf(`
	%seditor.customConfig.uploadImgServer = '%s';
	%seditor.customConfig.uploadImgMaxSize = 3 * 1024 * 1024;
	%seditor.customConfig.uploadImgMaxLength = 5;
	%seditor.customConfig.uploadFileName = 'file';
`, field, url, field, field, field))

	var fileUploadHandler context.Handler
	if len(data) > 1 {
		fileUploadHandler = data[1].(context.Handler)
	} else {
		fileUploadHandler = func(ctx *context.Context) {
			multiForm := ctx.Request.MultipartForm

			if len(multiForm.File) == 0 {
				ctx.JSON(http.StatusOK, map[string]interface{}{
					"errno": 400,
				})
				return
			}

			err := file.GetFileEngine(config.GetFileUploadEngine().Name).Upload(multiForm)
			if err != nil {
				ctx.JSON(http.StatusOK, map[string]interface{}{ "errno": 500 })
				return
			}

			fileValue := multiForm.Value["file"]
			store     := config.GetStore()
			imgPath   := make([]string, len(fileValue))
			for i, path := range fileValue {
				imgPath[i] = store.URL(path)
			}

			ctx.JSON(http.StatusOK, map[string]interface{}{ "errno": 0, "data": imgPath })
		}
	}

	f.Callbacks = f.Callbacks.AddCallback(context.Node{
		Path:     url,
		Method:   "post",
		Value:    map[string]interface{}{ constant.ContextNodeNeedAuth: 1 },
		Handlers: []context.Handler{ fileUploadHandler },
	})

	return f
}

func (f *FormPanel) FieldDefault(def string) *FormPanel {
	f.FieldList[f.curFieldListIndex].Default = template.HTML(def)
	return f
}

// FieldNotAllowEdit means when update record the field can not be edited but will still be displayed and submitted.
// Deprecated: Use FieldDisplayButCanNotEditWhenUpdate instead.
func (f *FormPanel) FieldNotAllowEdit() *FormPanel {
	f.FieldList[f.curFieldListIndex].Editable = false
	return f
}

// FieldDisplayButCanNotEditWhenUpdate means when update record the field can not be edited but will still be displayed and submitted.
func (f *FormPanel) FieldDisplayButCanNotEditWhenUpdate() *FormPanel {
	f.FieldList[f.curFieldListIndex].Editable = false
	return f
}

// FieldDisableWhenUpdate means when update record the field can not be edited, displayed and submitted.
func (f *FormPanel) FieldDisableWhenUpdate() *FormPanel {
	f.FieldList[f.curFieldListIndex].NotAllowEdit = true
	return f
}

// FieldNotAllowAdd means when create record the field can not be edited, displayed and submitted.
// Deprecated: Use FieldDisableWhenCreate instead.
func (f *FormPanel) FieldNotAllowAdd() *FormPanel {
	f.FieldList[f.curFieldListIndex].NotAllowAdd = true
	return f
}

// FieldDisableWhenCreate means when create record the field can not be edited, displayed and submitted.
func (f *FormPanel) FieldDisableWhenCreate() *FormPanel {
	f.FieldList[f.curFieldListIndex].NotAllowAdd = true
	return f
}

// FieldDisplayButCanNotEditWhenCreate means when create record the field can not be edited but will still be displayed and submitted.
func (f *FormPanel) FieldDisplayButCanNotEditWhenCreate() *FormPanel {
	f.FieldList[f.curFieldListIndex].DisplayButNotAdd = true
	return f
}

// FieldHideWhenCreate means when create record the field can not be edited and displayed, but will be submitted.
func (f *FormPanel) FieldHideWhenCreate() *FormPanel {
	f.FieldList[f.curFieldListIndex].CreateHide = true
	return f
}

// FieldHideWhenUpdate means when update record the field can not be edited and displayed, but will be submitted.
func (f *FormPanel) FieldHideWhenUpdate() *FormPanel {
	f.FieldList[f.curFieldListIndex].EditHide = true
	return f
}

func (f *FormPanel) FieldFormType(formType form2.Type) *FormPanel {
	f.FieldList[f.curFieldListIndex].FormType = formType
	return f
}

func (f *FormPanel) FieldValue(value string) *FormPanel {
	f.FieldList[f.curFieldListIndex].Value = template.HTML(value)
	return f
}

func (f *FormPanel) FieldOptionsFromTable(table, textFieldName, valueFieldName string, process ...OptionTableQueryProcessFn) *FormPanel {
	var fn OptionTableQueryProcessFn
	if len(process) > 0 {
		fn = process[0]
	}
	f.FieldList[f.curFieldListIndex].OptionTable = OptionTable{
		Table:          table,
		TextField:      textFieldName,
		ValueField:     valueFieldName,
		QueryProcessFn: fn,
	}
	return f
}

func (f *FormPanel) FieldOptionsTableProcessFn(fn OptionProcessFn) *FormPanel {
	f.FieldList[f.curFieldListIndex].OptionTable.ProcessFn = fn
	return f
}

func (f *FormPanel) FieldOptions(options FieldOptions) *FormPanel {
	f.FieldList[f.curFieldListIndex].Options = options
	return f
}

func (f *FormPanel) FieldDefaultOptionDelimiter(delimiter string) *FormPanel {
	f.FieldList[f.curFieldListIndex].DefaultOptionDelimiter = delimiter
	return f
}

func (f *FormPanel) FieldPostFilterFn(post PostFieldFilterFn) *FormPanel {
	f.FieldList[f.curFieldListIndex].PostFilterFn = post
	return f
}

func (f *FormPanel) FieldNow() *FormPanel {
	f.FieldList[f.curFieldListIndex].PostFilterFn = func(value PostFieldModel) interface{} {
		return utils.NowStr()
	}
	return f
}

func (f *FormPanel) FieldNowWhenUpdate() *FormPanel {
	f.FieldList[f.curFieldListIndex].PostFilterFn = func(value PostFieldModel) interface{} {
		if value.IsUpdate() { return utils.NowStr() }
		return value.Value.Value()
	}
	return f
}

func (f *FormPanel) FieldNowWhenInsert() *FormPanel {
	f.FieldList[f.curFieldListIndex].PostFilterFn = func(value PostFieldModel) interface{} {
		if value.IsCreate() { return utils.NowStr() }
		return value.Value.Value()
	}
	return f
}

func (f *FormPanel) FieldLimit(limit int) *FormPanel {
	f.FieldList[f.curFieldListIndex].DisplayProcessChains = f.FieldList[f.curFieldListIndex].AddLimit(limit)
	return f
}

func (f *FormPanel) FieldTrimSpace() *FormPanel {
	f.FieldList[f.curFieldListIndex].DisplayProcessChains = f.FieldList[f.curFieldListIndex].AddTrimSpace()
	return f
}

func (f *FormPanel) FieldSubstr(start int, end int) *FormPanel {
	f.FieldList[f.curFieldListIndex].DisplayProcessChains = f.FieldList[f.curFieldListIndex].AddSubstr(start, end)
	return f
}

func (f *FormPanel) FieldToTitle() *FormPanel {
	f.FieldList[f.curFieldListIndex].DisplayProcessChains = f.FieldList[f.curFieldListIndex].AddToTitle()
	return f
}

func (f *FormPanel) FieldToUpper() *FormPanel {
	f.FieldList[f.curFieldListIndex].DisplayProcessChains = f.FieldList[f.curFieldListIndex].AddToUpper()
	return f
}

func (f *FormPanel) FieldToLower() *FormPanel {
	f.FieldList[f.curFieldListIndex].DisplayProcessChains = f.FieldList[f.curFieldListIndex].AddToLower()
	return f
}

func (f *FormPanel) FieldXssFilter() *FormPanel {
	f.FieldList[f.curFieldListIndex].DisplayProcessChains = f.FieldList[f.curFieldListIndex].DisplayProcessChains.Add(xssFilter)
	return f
}

func (f *FormPanel) FieldCustomContent(content template.HTML) *FormPanel {
	f.FieldList[f.curFieldListIndex].CustomContent = content
	return f
}

func (f *FormPanel) FieldCustomJs(js template.JS) *FormPanel {
	f.FieldList[f.curFieldListIndex].CustomJs = js
	return f
}

func (f *FormPanel) FieldCustomCss(css template.CSS) *FormPanel {
	f.FieldList[f.curFieldListIndex].CustomCss = css
	return f
}

func (f *FormPanel) FieldOnSearch(url string, handler Handler, delay ...int) *FormPanel {
	ext, callback := searchJS(f.FieldList[f.curFieldListIndex].OptionExt, f.OperationURL(url), handler, delay...)
	f.FieldList[f.curFieldListIndex].OptionExt = ext
	f.Callbacks = f.Callbacks.AddCallback(callback)
	return f
}

func (f *FormPanel) FieldOnChooseCustom(js template.HTML) *FormPanel {
	f.FooterHtml += chooseCustomJS(f.FieldList[f.curFieldListIndex].Field, js)
	return f
}

type LinkField struct {
	Field   string
	Value   template.HTML
	Hide    bool
	Disable bool
}

func (f *FormPanel) FieldOnChooseMap(m map[string]LinkField) *FormPanel {
	f.FooterHtml += chooseMapJS(f.FieldList[f.curFieldListIndex].Field, m)
	return f
}

func (f *FormPanel) FieldOnChoose(val, field string, value template.HTML) *FormPanel {
	f.FooterHtml += chooseJS(f.FieldList[f.curFieldListIndex].Field, field, val, value)
	return f
}

func (f *FormPanel) OperationURL(id string) string {
	return config.Url("/operation/" + utils.WrapURL(id))
}

func (f *FormPanel) FieldOnChooseAjax(field, url string, handler Handler, custom ...template.HTML) *FormPanel {
	js, callback := chooseAjax(f.FieldList[f.curFieldListIndex].Field, field, f.OperationURL(url), handler, custom...)
	f.FooterHtml += js
	f.Callbacks = f.Callbacks.AddCallback(callback)
	return f
}

func (f *FormPanel) FieldOnChooseHide(value string, field ...string) *FormPanel {
	f.FooterHtml += chooseHideJS(f.FieldList[f.curFieldListIndex].Field, []string{ value }, field...)
	return f
}

func (f *FormPanel) FieldOnChooseOptionsHide(values []string, field ...string) *FormPanel {
	f.FooterHtml += chooseHideJS(f.FieldList[f.curFieldListIndex].Field, values, field...)
	return f
}

func (f *FormPanel) FieldOnChooseShow(value string, field ...string) *FormPanel {
	f.FooterHtml += chooseShowJS(f.FieldList[f.curFieldListIndex].Field, []string{ value }, field...)
	return f
}

func (f *FormPanel) FieldOnChooseOptionsShow(values []string, field ...string) *FormPanel {
	f.FooterHtml += chooseShowJS(f.FieldList[f.curFieldListIndex].Field, values, field...)
	return f
}

func (f *FormPanel) FieldOnChooseDisable(value string, field ...string) *FormPanel {
	f.FooterHtml += chooseDisableJS(f.FieldList[f.curFieldListIndex].Field, []string{ value }, field...)
	return f
}

func (f *FormPanel) addFooterHTML(footer template.HTML) *FormPanel {
	f.FooterHtml += template.HTML(ParseTableDataTmpl(footer))
	return f
}

func (f *FormPanel) AddCSS(css template.CSS) *FormPanel {
	return f.addFooterHTML(template.HTML(utils.StrConcat("<style>", string(css), "</style>")))
}

func (f *FormPanel) AddJS(js template.JS) *FormPanel {
	return f.addFooterHTML(template.HTML(utils.StrConcat("<script>", string(js), "</script>")))
}

func searchJS(ext template.JS, url string, handler Handler, delay ...int) (template.JS, context.Node) {
	delayStr := "500"
	if len(delay) > 0 {
		delayStr = strconv.Itoa(delay[0])
	}

	if ext != "" {
		s := string(ext)
		s = strings.Replace(s, "{", "", 1)
		s = utils.ReplaceNth(s, "}", "", strings.Count(s, "}"))
		s = strings.TrimRight(s, " ")
		ext = template.JS(s + ",")
	}

	return template.JS(utils.StrConcat(`{
		`, string(ext), `
		ajax: {
			url: "`, url, `",
			dataType: 'json',
			data: function(params){ return { search: params.term, page: params.page || 1 }},
			delay: `, delayStr, `,
			processResults: function(data, params){ return data.data }
		}
	}`)), context.Node{
			Path:     url,
			Method:   "get",
			Handlers: context.Handlers{ handler.Wrap() },
			Value:    map[string]interface{}{ constant.ContextNodeNeedAuth: 1 },
		}
}

func chooseCustomJS(field string, js template.HTML) template.HTML {
	return utils.ParseHTML("choose_custom", tmpls["choose_custom"], struct {
		Field template.JS
		JS    template.JS
	}{
		Field: template.JS(field),
		JS:    template.JS(js),
	})
}

func chooseMapJS(field string, m map[string]LinkField) template.HTML {
	return utils.ParseHTML("choose_map", tmpls["choose_map"], struct {
		Field template.JS
		Data  map[string]LinkField
	}{
		Field: template.JS(field),
		Data:  m,
	})
}

func chooseJS(field, chooseField, val string, value template.HTML) template.HTML {
	return utils.ParseHTML("choose", tmpls["choose"], struct {
		Field       template.JS
		ChooseField template.JS
		Val         template.JS
		Value       template.JS
	}{
		Field:       template.JS(field),
		ChooseField: template.JS(chooseField),
		Value:       decorateChooseValue([]string{string(value)}),
		Val:         decorateChooseValue([]string{string(val)}),
	})
}

func chooseAjax(field, chooseField, url string, handler Handler, js ...template.HTML) (template.HTML, context.Node) {
	actionJS  := template.JS("")
	passValue := template.JS("")

	switch len(js) {
	case 0:
	case 1:
		actionJS = template.JS(js[0])
	default:
		_ = js[1]
		actionJS  = template.JS(js[0])
		passValue = template.JS(js[1])
	}

	return utils.ParseHTML("choose_ajax", tmpls["choose_ajax"], struct {
			Field       template.JS
			ChooseField template.JS
			PassValue   template.JS
			ActionJS    template.JS
			Url         template.JS
		}{
			Url:         template.JS(url),
			Field:       template.JS(field),
			ChooseField: template.JS(chooseField),
			PassValue:   passValue,
			ActionJS:    actionJS,
		}), context.Node{
			Path:     url,
			Method:   "post",
			Handlers: context.Handlers{ handler.Wrap() },
			Value:    map[string]interface{}{ constant.ContextNodeNeedAuth: 1 },
		}
}

func chooseHideJS(field string, value []string, chooseFields ...string) template.HTML {
	if len(chooseFields) == 0 { return "" }

	return utils.ParseHTML("choose_hide", tmpls["choose_hide"], struct {
		Field        template.JS
		Value        template.JS
		ChooseFields []string
	}{
		Field:        template.JS(field),
		Value:        decorateChooseValue(value),
		ChooseFields: chooseFields,
	})
}

func chooseShowJS(field string, value []string, chooseFields ...string) template.HTML {
	if len(chooseFields) == 0 { return "" }

	return utils.ParseHTML("choose_show", tmpls["choose_show"], struct {
		Field        template.JS
		Value        template.JS
		ChooseFields []string
	}{
		Field:        template.JS(field),
		Value:        decorateChooseValue(value),
		ChooseFields: chooseFields,
	})
}

func chooseDisableJS(field string, value []string, chooseFields ...string) template.HTML {
	if len(chooseFields) == 0 { return "" }

	return utils.ParseHTML("choose_disable", tmpls["choose_disable"], struct {
		Field        template.JS
		Value        template.JS
		ChooseFields []string
	}{
		Field:        template.JS(field),
		Value:        decorateChooseValue(value),
		ChooseFields: chooseFields,
	})
}

func decorateChooseValue(val []string) template.JS {
	if len(val) == 0 { return "" }

	//res := make([]string, len(val))
	var sb strings.Builder
	sb.Grow(64)
	sb.WriteByte('[')

	for k, v := range val {
		if k > 0 { sb.WriteByte(',') }
		switch {
		case v == "":
			sb.WriteString(`""`)
		case v[0] != '"':
			sb.WriteByte('"')
			sb.WriteString(v)
			sb.WriteByte('"')
		default:
			sb.WriteString(v)
		}
		/*if v == "" {
			v = `""`
		}
		if v[0] != '"' {
			if strings.Contains(v, "$(this)") {
				res[k] = v
			}
			if v == "{{.vue}}" {
				res[k] = "$(this).v()"
			}
			if len(v) > 3 && v[:3] == "js:" {
				res[k] = v[3:]
			}
			res[k] = `"` + v + `"`
		} else {
			res[k] = v
		}*/
	}

	sb.WriteByte(']')
	return template.JS(sb.String())
	//return template.JS("[" + strings.Join(res, ",") + "]")
}

// FormPanel attribute setting functions
// ====================================================

func (f *FormPanel) SetTitle(title string) *FormPanel {
	f.Title = title
	return f
}

func (f *FormPanel) SetTabGroups(groups TabGroups) *FormPanel {
	f.TabGroups = groups
	return f
}

func (f *FormPanel) SetTabHeaders(headers ...string) *FormPanel {
	f.TabHeaders = headers
	return f
}

func (f *FormPanel) SetDescription(desc string) *FormPanel {
	f.Description = desc
	return f
}

func (f *FormPanel) SetHeaderHtml(header template.HTML) *FormPanel {
	f.HeaderHtml += header
	return f
}

func (f *FormPanel) SetFooterHtml(footer template.HTML) *FormPanel {
	f.FooterHtml += footer
	return f
}

func (f *FormPanel) HasError() bool {
	return f.PageError != nil
}

func (f *FormPanel) SetError(err errors.PageError, content ...template.HTML) *FormPanel {
	f.PageError = err
	if len(content) > 0 {
		f.PageErrorHTML = content[0]
	}
	return f
}

func (f *FormPanel) SetNoCompress() *FormPanel {
	f.NoCompress = true
	return f
}

func (f *FormPanel) Set404Error(content ...template.HTML) *FormPanel {
	f.SetError(errors.PageError404, content...)
	return f
}

func (f *FormPanel) Set403Error(content ...template.HTML) *FormPanel {
	f.SetError(errors.PageError403, content...)
	return f
}

func (f *FormPanel) Set400Error(content ...template.HTML) *FormPanel {
	f.SetError(errors.PageError401, content...)
	return f
}

func (f *FormPanel) Set500Error(content ...template.HTML) *FormPanel {
	f.SetError(errors.PageError500, content...)
	return f
}

func (f *FormPanel) SetLayout(layout form2.Layout) *FormPanel {
	f.Layout = layout
	return f
}

func (f *FormPanel) SetPostValidator(va FormPostFn) *FormPanel {
	f.Validator = va
	return f
}

func (f *FormPanel) SetPreProcessFn(fn FormPreProcessFn) *FormPanel {
	f.PreProcessFn = fn
	return f
}

func (f *FormPanel) SetHTMLContent(content template.HTML) *FormPanel {
	f.HTMLContent = content
	return f
}

func (f *FormPanel) SetHeader(content template.HTML) *FormPanel {
	f.Header = content
	return f
}

func (f *FormPanel) SetInputWidth(width int) *FormPanel {
	f.InputWidth = width
	return f
}

func (f *FormPanel) SetHeadWidth(width int) *FormPanel {
	f.HeadWidth = width
	return f
}

func (f *FormPanel) SetWrapper(wrapper ContentWrapper) *FormPanel {
	f.Wrapper = wrapper
	return f
}

func (f *FormPanel) SetHideSideBar() *FormPanel {
	f.HideSideBar = true
	return f
}

func (f *FormPanel) SetFormNewTitle(title template.HTML) *FormPanel {
	f.FormNewTitle = title
	return f
}

func (f *FormPanel) SetFormNewBtnWord(word template.HTML) *FormPanel {
	f.FormNewBtnWord = word
	return f
}

func (f *FormPanel) SetFormEditTitle(title template.HTML) *FormPanel {
	f.FormEditTitle = title
	return f
}

func (f *FormPanel) SetFormEditBtnWord(word template.HTML) *FormPanel {
	f.FormEditBtnWord = word
	return f
}

func (f *FormPanel) SetResponder(responder Responder) *FormPanel {
	f.Responder = responder
	return f
}

type AjaxData struct {
	SuccessTitle   string
	SuccessText    string
	ErrorTitle     string
	ErrorText      string
	SuccessJumpURL string
	DisableJump    bool
	SuccessJS      string
	JumpInNewTab   string
}

func (f *FormPanel) EnableAjaxData(data AjaxData) *FormPanel {
	langGotIt := language.Get("got it")
	f.Ajax = true
	if f.AjaxSuccessJS == "" {
		var successMsg, errorMsg, jump, text, wrongText string
		if data.SuccessTitle != "" {
			successMsg = utils.StrConcat(`"`, data.SuccessTitle, `"`)
		} else {
			successMsg = "data.msg"
		}
		if data.ErrorTitle != "" {
			errorMsg = utils.StrConcat(`"`, data.ErrorTitle, `"`)
		} else {
			errorMsg = "data.msg"
		}
		if data.SuccessJumpURL != "" {
			jump = utils.StrConcat(`"`, data.SuccessJumpURL, `"`)
		} else {
			jump = "data.data.url"
		}
		if data.SuccessText != "" {
			text = utils.StrConcat(`text:"`, data.SuccessText, `",`)
		}
		if data.ErrorText != "" {
			wrongText = utils.StrConcat(`text:"`, data.ErrorText, `"`)
		} else {
			wrongText = "text:data.msg"
		}
		jumpURL := ""
		if !data.DisableJump {
			if data.JumpInNewTab != "" {
				jumpURL = utils.StrConcat(`listenerForAddNavTab(`, jump, `, "`, data.JumpInNewTab + `");`)
			}
			jumpURL = utils.StrConcat(jumpURL, `$.pjax({url: `, jump, `, container: '#pjax-container'});`)
		} else {
			jumpURL = `if(data.data && data.data.token !== "") { $("input[name='__go_admin_t_']").val(data.data.token) }`
		}
		f.AjaxSuccessJS = template.JS(utils.StrConcat(`if(typeof(data) === "string") { data = JSON.parse(data) }
	if(data.code === 200) {
		swal({
			type: "success",
			title: `, successMsg, `,
			`, text, `
			showCancelButton: false,
			confirmButtonColor: "#3c8dbc",
			confirmButtonText: '`, langGotIt, `',
        }, function() {
			$(".modal-backdrop.fade.in").remove();
			`, jumpURL, `
			`, data.SuccessJS, `
        });
	} else {
		if(data.data && data.data.token !== "") { $("input[name='__go_admin_t_']").val(data.data.token) }
		swal({
			type: "error",
			title: `, errorMsg, `,
			`, wrongText, `,
			showCancelButton: false,
			confirmButtonColor: "#3c8dbc",
			confirmButtonText: "`, langGotIt, `",
        })
	}
`))
	}
	if f.AjaxErrorJS == "" {
		var errorMsg, errorMsg2, wrongText string
		if data.ErrorTitle != "" {
			errorMsg = utils.StrConcat(`"`, data.ErrorTitle, `"`)
		} else {
			errorMsg = "data.responseJSON.msg"
		}
		if data.ErrorTitle != "" {
			errorMsg2 = data.ErrorTitle
		} else {
			errorMsg2 = language.Get("error")
		}
		if data.ErrorText != "" {
			wrongText = utils.StrConcat(`"`, data.ErrorText, `"`)
		} else {
			wrongText = "data.msg"
		}
		f.AjaxErrorJS = template.JS(utils.StrConcat(`
	if(data.responseText !== "") {
		if(data.responseJSON.data && data.responseJSON.data.token !== "") {
			$("input[name='__go_admin_t_']").val(data.responseJSON.data.token)
		}
		swal({
			type: "error",
			title: `, errorMsg, `,
			`, wrongText, `
			showCancelButton: false,
			confirmButtonColor: "#3c8dbc",
			confirmButtonText: '`, langGotIt, `',
        })
	} else {
		swal({
			type: "error",
			title: "`, errorMsg2, `",
			text: `, wrongText, `,
			showCancelButton: false,
			confirmButtonColor: "#3c8dbc",
			confirmButtonText: "`, langGotIt, `",
		})
	}
`))
	}
	return f
}

func (f *FormPanel) EnableAjax(msgs ...string) *FormPanel {
	var data AjaxData
	switch len(msgs) {
	case 0:
	case 1:
		if m := msgs[0]; m != "" { data.SuccessTitle = m }
	case 2:
		_ = msgs[1]
		if m := msgs[0]; m != "" { data.SuccessTitle = m }
		if m := msgs[1]; m != "" { data.ErrorTitle   = m }
	case 3:
		_ = msgs[2]
		if m := msgs[0]; m != "" { data.SuccessTitle   = m }
		if m := msgs[1]; m != "" { data.ErrorTitle     = m }
		if m := msgs[2]; m != "" { data.SuccessJumpURL = m }
	case 4:
		_ = msgs[3]
		if m := msgs[0]; m != "" { data.SuccessTitle   = m }
		if m := msgs[1]; m != "" { data.ErrorTitle     = m }
		if m := msgs[2]; m != "" { data.SuccessJumpURL = m }
		if m := msgs[3]; m != "" { data.SuccessText    = m }
	default:
		_ = msgs[4]
		if m := msgs[0]; m != "" { data.SuccessTitle   = m }
		if m := msgs[1]; m != "" { data.ErrorTitle     = m }
		if m := msgs[2]; m != "" { data.SuccessJumpURL = m }
		if m := msgs[3]; m != "" { data.SuccessText    = m }
		if m := msgs[4]; m != "" { data.ErrorText      = m }
	}
	return f.EnableAjaxData(data)
}

func (f *FormPanel) SetAjaxSuccessJS(js template.JS) *FormPanel {
	f.AjaxSuccessJS = js
	return f
}

func (f *FormPanel) SetAjaxErrorJS(js template.JS) *FormPanel {
	f.AjaxErrorJS = js
	return f
}

func (f *FormPanel) SetPostHook(fn FormPostFn) *FormPanel {
	f.PostHook = fn
	return f
}

func (f *FormPanel) SetUpdateFn(fn FormPostFn) *FormPanel {
	f.UpdateFn = fn
	return f
}

func (f *FormPanel) SetInsertFn(fn FormPostFn) *FormPanel {
	f.InsertFn = fn
	return f
}

func (f *FormPanel) GroupFieldWithValue(pk, id string, columnMap map[string]struct{}, res map[string]interface{}, sql func() *db.SQL) ([]FormFields, []string) {
	groupFormList := make([]FormFields, len(f.TabGroups))
	groupHeaders  := make([]string    , len(f.TabGroups))
	existField    := make(map[string]struct{}, len(f.TabGroups))
	hasPK         := false

	for i, group := range f.TabGroups {
		list := make(FormFields, 0, len(group))
		for j, fieldName := range group {
			field := f.FieldList.FindByFieldName(fieldName)
			if field != nil && field.isNotBelongToATable() && !field.NotAllowEdit {
				if !field.Hide {
					field.Hide = field.EditHide
				}
				if field.FormType.IsTable() {
					for z, tf := range field.TableFields {
						rowValue := tf.GetRawValue(columnMap, res[tf.Field])
						if tf.Field == pk {
							hasPK = true
						}
						field.TableFields[z] = *(tf.UpdateValue(id, rowValue, res, sql()))
					}
					if _, ok := existField[field.Field]; ok {
						field.Field += "_ga_group_" + strconv.Itoa(j)
					}
					list = append(list, *field)
				} else {
					if field.Field == pk {
						hasPK = true
					}
					rowValue := field.GetRawValue(columnMap, res[field.Field])
					if _, ok := existField[field.Field]; ok {
						field.Field += "_ga_group_" + strconv.Itoa(j)
					}
					list = append(list, *(field.UpdateValue(id, rowValue, res, sql())))
				}
				existField[field.Field] = struct{}{}
			}
		}

		groupFormList[i] = list.FillCustomContent()
		groupHeaders [i] = f.TabHeaders[i]
	}

	if len(groupFormList) > 0 && !hasPK {
		groupFormList[len(groupFormList)-1] = groupFormList[len(groupFormList)-1].Add(&FormField{
			Head:       pk,
			FieldClass: pk,
			Field:      pk,
			Value:      template.HTML(id),
			Hide:       true,
		})
	}

	return groupFormList, groupHeaders
}

func (f *FormPanel) GroupField(sqls ...func() *db.SQL) ([]FormFields, []string) {
	var sql func() *db.SQL
	if len(sqls) > 0 { sql = sqls[0] }
	if sql == nil { sql = db.NilSQL }

	groupFormList := make([]FormFields, len(f.TabGroups))
	groupHeaders  := make([]string    , len(f.TabGroups))
	existField    := make(map[string]struct{}, len(f.TabGroups))

	for i, group := range f.TabGroups {
		list := make(FormFields, 0, len(group))
		for j, fieldName := range group {
			field := f.FieldList.FindByFieldName(fieldName)
			if field != nil && field.isNotBelongToATable() && field.allowAdd() {
				field.Editable = !field.DisplayButNotAdd
				if !field.Hide {
					field.Hide = field.CreateHide
				}
				if field.FormType.IsTable() {
					for z, tf := range field.TableFields {
						field.TableFields[z] = *(tf.UpdateDefaultValue(sql()))
					}
					if _, ok := existField[field.Field]; ok {
						field.Field += "_ga_group_" + strconv.Itoa(j)
					}
					list = append(list, *field)
				} else {
					if _, ok := existField[field.Field]; ok {
						field.Field += "_ga_group_" + strconv.Itoa(j)
					}
					list = append(list, *(field.UpdateDefaultValue(sql())))
				}
				existField[field.Field] = struct{}{}
			}
		}
		groupFormList[i] = list.FillCustomContent()
		groupHeaders [i] = f.TabHeaders[i]
	}

	return groupFormList, groupHeaders
}

func (f *FormPanel) FieldsWithValue(pk, id string, columnMap map[string]struct{}, res map[string]interface{}, sql func() *db.SQL) FormFields {
	if sql == nil { sql = db.NilSQL }
	list  := make(FormFields, 0, len(f.FieldList))
	hasPK := false
	for i, ff := range f.FieldList {
		if !ff.NotAllowEdit {
			if !ff.Hide {
				ff.Hide = ff.EditHide
			}
			rowValue := ff.GetRawValue(columnMap, res[ff.Field])
			if ff.FatherField != "" {
				f.FieldList.FindTableField(ff.Field, ff.FatherField).UpdateValue(id, rowValue, res, sql())
			} else if ff.FormType.IsTable() {
				list = append(list, ff)
			} else {
				list = append(list, *(ff.UpdateValue(id, rowValue, res, sql())))
			}
			if ff.Field == pk {
				hasPK = true
			}
			f.FieldList[i] = ff
		}
	}
	if !hasPK {
		list = list.Add(&FormField{
			Head:       pk,
			FieldClass: pk,
			Field:      pk,
			Value:      template.HTML(id),
			FormType:   form2.Default,
			Hide:       true,
		})
	}
	return list.FillCustomContent()
}

func (f *FormPanel) FieldsWithDefaultValue(sqls ...func() *db.SQL) FormFields {
	var sql func() *db.SQL
	if len(sqls) > 0 { sql = sqls[0] }
	if sql == nil { sql = db.NilSQL }

	list := make(FormFields, 0, len(f.FieldList))
	for i, ff := range f.FieldList {
		if ff.allowAdd() {
			ff.Editable = !ff.DisplayButNotAdd
			if !ff.Hide {
				ff.Hide = ff.CreateHide
			}
			if ff.FatherField != "" {
				f.FieldList.FindTableField(ff.Field, ff.FatherField).UpdateDefaultValue(sql())
			} else if ff.FormType.IsTable() {
				list = append(list, ff)
			} else {
				list = append(list, *(ff.UpdateDefaultValue(sql())))
			}
			f.FieldList[i] = ff
		}
	}
	return list.FillCustomContent().RemoveNotShow()
}

func (f *FormPanel) GetNewFormFields(sql ...func() *db.SQL) (FormFields, []FormFields, []string) {
	if len(f.TabGroups) > 0 {
		tabFields, tabHeaders := f.GroupField(sql...)
		return nil, tabFields, tabHeaders
	}
	return f.FieldsWithDefaultValue(sql...), make([]FormFields, 0), make([]string, 0)
}

type (
	FormPreProcessFn  func(values form.Values) form.Values
	FormPostFn        func(values form.Values) error
	FormFields        []FormField
	GroupFormFields   []FormFields
	GroupFieldHeaders []string
)

func (f FormFields) Copy() FormFields {
	formList := make(FormFields, len(f))
	for i, ff := range f {
		options := make(FieldOptions, len(ff.Options))
		copy(options, ff.Options)
		ff.Options = options
		formList[i] = ff
	}
	return formList
}

func (f FormFields) FindByFieldName(field string) *FormField {
	for i, ff := range f {
		if ff.Field == field {
			return &f[i]
		}
	}
	return nil
}

func (f FormFields) FindIndexByFieldName(field string) int {
	for i, ff := range f {
		if ff.Field == field {
			return i
		}
	}
	return -1
}

func (f FormFields) FindTableField(field, father string) *FormField {
	return f.FindByFieldName(father).TableFields.FindByFieldName(field)
}

func (f FormFields) FindTableChildren(father string) []*FormField {
	var list []*FormField
	for i, ff := range f {
		if ff.FatherField == father {
			list = append(list, &f[i])
		}
	}
	return list
}

func (f FormFields) FillCustomContent() FormFields {
	for i, ff := range f {
		if ff.FormType.IsCustom() {
			f[i] = *ff.FillCustomContent()
		}
	}
	return f
}

func (f FormFields) Add(field *FormField) FormFields {
	return append(f, *field)
}

func (f FormFields) RemoveNotShow() FormFields {
	ff := f
	for i := len(ff) - 1; i >= 0; {
		if ff[i].FatherFormType == form2.Table {
			ff = append(ff[:i], ff[i+1:]...)
		} else {
			i--
		}
	}
	return ff
}
