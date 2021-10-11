package types

import (
	"fmt"
	"html"
	"html/template"
	"strings"

	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

type DisplayFnGenerator interface {
	Get(args ...interface{}) FieldFilterFn
	JS() template.HTML
	HTML() template.HTML
}

type BaseDisplayFnGenerator struct{}

func (base *BaseDisplayFnGenerator) JS  () template.HTML { return "" }
func (base *BaseDisplayFnGenerator) HTML() template.HTML { return "" }

var displayFnGens = make(map[string]DisplayFnGenerator)

func RegisterDisplayFnGenerator(key string, gen DisplayFnGenerator) {
	if _, ok := displayFnGens[key]; ok {
		panic("display function generator has been registered")
	}
	displayFnGens[key] = gen
}

type FieldDisplay struct {
	Display              FieldFilterFn
	DisplayProcessChains DisplayProcessFnChains
}

func (f FieldDisplay) ToDisplay(value FieldModel) interface{} {
	val := f.Display(value)
	if len(f.DisplayProcessChains) > 0 && f.IsNotSelectRes(val) {
		valStr := fmt.Sprintf("%v", val)
		for _, process := range f.DisplayProcessChains {
			valStr = fmt.Sprintf("%v", process(FieldModel{
				Row:   value.Row,
				Value: valStr,
				ID:    value.ID,
			}))
		}
		return valStr
	}
	return val
}

func (f FieldDisplay) IsNotSelectRes(v interface{}) bool {
	switch v.(type) {
	case template.HTML: return false
	case []string     : return false
	case [][]string   : return false
	default           : return true
	}
}

func (f FieldDisplay) ToDisplayHTML(value FieldModel) template.HTML {
	v := f.ToDisplay(value)
	switch t := v.(type) {
	case template.HTML:
		return t
	case string:
		return template.HTML(t)
	case []string:
		if len(t) > 0 { return template.HTML(t[0]) }
	case []template.HTML:
		if len(t) > 0 { return t[0] }
	case nil:
	default:
		return template.HTML(fmt.Sprintf("%v", v))
	}
	return ""
	/*if h, ok := v.(template.HTML); ok {
		return h
	} else if s, ok := v.(string); ok {
		return template.HTML(s)
	} else if arr, ok := v.([]string); ok && len(arr) > 0 {
		return template.HTML(arr[0])
	} else if arr, ok := v.([]template.HTML); ok && len(arr) > 0 {
		return arr[0]
	} else if v != nil {
		return template.HTML(fmt.Sprintf("%v", v))
	} else {
		return ""
	}*/
}

func (f FieldDisplay) ToDisplayString(value FieldModel) string {
	v := f.ToDisplay(value)
	switch t := v.(type) {
	case template.HTML:
		return string(t)
	case string:
		return t
	case []string:
		if len(t) > 0 { return t[0] }
	case []template.HTML:
		if len(t) > 0 { return string(t[0]) }
	case nil:
	default:
		return fmt.Sprintf("%v", v)
	}
	return ""
	/*if h, ok := v.(template.HTML); ok {
		return string(h)
	} else if s, ok := v.(string); ok {
		return s
	} else if arr, ok := v.([]string); ok && len(arr) > 0 {
		return arr[0]
	} else if arr, ok := v.([]template.HTML); ok && len(arr) > 0 {
		return string(arr[0])
	} else if v != nil {
		return fmt.Sprintf("%v", v)
	} else {
		return ""
	}*/
}

func (f FieldDisplay) ToDisplayStringArray(value FieldModel) []string {
	v := f.ToDisplay(value)
	switch t := v.(type) {
	case template.HTML:
		return []string{ string(t) }
	case string:
		return []string{ t }
	case []string:
		return t
	case []template.HTML:
		ss := make([]string, len(t))
		for i, s := range t {
			ss[i] = string(s)
		}
		return ss
	case nil:
	default:
		return []string{ fmt.Sprintf("%v", v) }
	}
	return nil
	/*if h, ok := v.(template.HTML); ok {
		return []string{string(h)}
	} else if s, ok := v.(string); ok {
		return []string{s}
	} else if arr, ok := v.([]string); ok && len(arr) > 0 {
		return arr
	} else if arr, ok := v.([]template.HTML); ok && len(arr) > 0 {
		ss := make([]string, len(arr))
		for k, a := range arr {
			ss[k] = string(a)
		}
		return ss
	} else if v != nil {
		return []string{fmt.Sprintf("%v", v)}
	} else {
		return []string{}
	}*/
}

func (f FieldDisplay) ToDisplayStringArrayArray(value FieldModel) [][]string {
	v := f.ToDisplay(value)
	switch t := v.(type) {
	case template.HTML:
		return [][]string{{ string(t) }}
	case string:
		return [][]string{{ t }}
	case []string:
		return [][]string{ t }
	case []template.HTML:
		ss := make([]string, len(t))
		for i, s := range t {
			ss[i] = string(s)
		}
		return [][]string{ ss }
	case nil:
	default:
		return [][]string{{ fmt.Sprintf("%v", v) }}
	}
	return nil
	/*if h, ok := v.(template.HTML); ok {
		return [][]string{{string(h)}}
	} else if s, ok := v.(string); ok {
		return [][]string{{s}}
	} else if arr, ok := v.([]string); ok && len(arr) > 0 {
		return [][]string{arr}
	} else if arr, ok := v.([][]string); ok && len(arr) > 0 {
		return arr
	} else if arr, ok := v.([]template.HTML); ok && len(arr) > 0 {
		ss := make([]string, len(arr))
		for k, a := range arr {
			ss[k] = string(a)
		}
		return [][]string{ss}
	} else if v != nil {
		return [][]string{{fmt.Sprintf("%v", v)}}
	} else {
		return [][]string{}
	}*/
}

func (f FieldDisplay) AddLimit(limit int) DisplayProcessFnChains {
	return f.DisplayProcessChains.Add(func(value FieldModel) interface{} {
		if limit > len(value.Value) {
			return value
		} else if limit < 0 {
			return ""
		}
		return value.Value[:limit]
	})
}

func (f FieldDisplay) AddTrimSpace() DisplayProcessFnChains {
	return f.DisplayProcessChains.Add(trimSpaceFilter)
}

func (f FieldDisplay) AddSubstr(start int, end int) DisplayProcessFnChains {
	return f.DisplayProcessChains.Add(func(value FieldModel) interface{} {
		if start > end || start > len(value.Value) || end < 0 {
			return ""
		}
		if start < 0 {
			start = 0
		}
		if end > len(value.Value) {
			end = len(value.Value)
		}
		return value.Value[start:end]
	})
}

func (f FieldDisplay) AddToTitle() DisplayProcessFnChains {
	return f.DisplayProcessChains.Add(toTitleFilter)
}

func (f FieldDisplay) AddToUpper() DisplayProcessFnChains {
	return f.DisplayProcessChains.Add(toUpperFilter)
}

func (f FieldDisplay) AddToLower() DisplayProcessFnChains {
	return f.DisplayProcessChains.Add(toLowerFilter)
}

type DisplayProcessFnChains []FieldFilterFn

func (d DisplayProcessFnChains) Valid() bool {
	return len(d) > 0
}

func (d DisplayProcessFnChains) Add(f FieldFilterFn) DisplayProcessFnChains {
	return append(d, f)
}

func (d DisplayProcessFnChains) Append(f DisplayProcessFnChains) DisplayProcessFnChains {
	return append(d, f...)
}

func (d DisplayProcessFnChains) Copy() DisplayProcessFnChains {
	if len(d) == 0 {
		return nil
	}
	newDisplayProcessFnChains := make(DisplayProcessFnChains, len(d))
	copy(newDisplayProcessFnChains, d)
	return newDisplayProcessFnChains
}

func chooseDisplayProcessChains(internal DisplayProcessFnChains) DisplayProcessFnChains {
	if len(internal) > 0 {
		return internal
	}
	return globalDisplayProcessChains.Copy()
}

var globalDisplayProcessChains DisplayProcessFnChains

func AddGlobalDisplayProcessFn(f FieldFilterFn) {
	globalDisplayProcessChains = globalDisplayProcessChains.Add(f)
}

func AddLimit(limit int) DisplayProcessFnChains {
	return addLimit(limit, globalDisplayProcessChains)
}

func AddTrimSpace() DisplayProcessFnChains {
	return addTrimSpace(globalDisplayProcessChains)
}

func AddSubstr(start int, end int) DisplayProcessFnChains {
	return addSubstr(start, end, globalDisplayProcessChains)
}

func AddToTitle() DisplayProcessFnChains {
	return addToTitle(globalDisplayProcessChains)
}

func AddToUpper() DisplayProcessFnChains {
	return addToUpper(globalDisplayProcessChains)
}

func AddToLower() DisplayProcessFnChains {
	return addToLower(globalDisplayProcessChains)
}

func AddXssFilter() DisplayProcessFnChains {
	return addXssFilter(globalDisplayProcessChains)
}

func AddXssJsFilter() DisplayProcessFnChains {
	return addXssJsFilter(globalDisplayProcessChains)
}

func addLimit(limit int, chains DisplayProcessFnChains) DisplayProcessFnChains {
	return chains.Add(func(value FieldModel) interface{} {
		if limit > len(value.Value) {
			return value
		} else if limit < 0 {
			return ""
		}
		return value.Value[:limit]
	})
}

func addTrimSpace(chains DisplayProcessFnChains) DisplayProcessFnChains {
	return chains.Add(trimSpaceFilter)
}

func trimSpaceFilter(value FieldModel) interface{} {
	return strings.TrimSpace(value.Value)
}

func addSubstr(start int, end int, chains DisplayProcessFnChains) DisplayProcessFnChains {
	return chains.Add(func(value FieldModel) interface{} {
		if start > end || start > len(value.Value) || end < 0 {
			return ""
		}
		if start < 0 {
			start = 0
		}
		if end > len(value.Value) {
			end = len(value.Value)
		}
		return value.Value[start:end]
	})
}

func addToTitle(chains DisplayProcessFnChains) DisplayProcessFnChains {
	return chains.Add(toTitleFilter)
}

func toTitleFilter(value FieldModel) interface{} {
	return strings.Title(value.Value)
}

func addToUpper(chains DisplayProcessFnChains) DisplayProcessFnChains {
	return chains.Add(toUpperFilter)
}

func toUpperFilter(value FieldModel) interface{} {
	return strings.ToUpper(value.Value)
}

func addToLower(chains DisplayProcessFnChains) DisplayProcessFnChains {
	return chains.Add(toLowerFilter)
}

func toLowerFilter(value FieldModel) interface{} {
	return strings.ToLower(value.Value)
}

func addXssFilter(chains DisplayProcessFnChains) DisplayProcessFnChains {
	return chains.Add(xssFilter)
}

func xssFilter(value FieldModel) interface{} {
	return html.EscapeString(value.Value)
}

func addXssJsFilter(chains DisplayProcessFnChains) DisplayProcessFnChains {
	return chains.Add(xssJsFilter)
}

func xssJsFilter(value FieldModel) interface{} {
	return utils.XssJsReplacer.Replace(value.Value)
}

func setDefaultDisplayFnOfFormType(f *FormPanel, typ form.Type) {
	if typ.IsMultiFile() {
		f.FieldList[f.curFieldListIndex].Display = multiFileFilter
	}
	if typ.IsSelect() {
		f.FieldList[f.curFieldListIndex].Display = splitFilter
	}
}

func multiFileFilter(value FieldModel) interface{} {
	if value.Value == "" {
		return ""
	}
	arr := strings.Split(value.Value, ",")
	store := config.GetStore()
	var sb strings.Builder
	sb.Grow(16 * len(arr))
	sb.WriteString("['")
	sb.WriteString(store.URL(arr[0]))
	if len(arr) > 1 {
		for _, item := range arr[1:] {
			sb.WriteString("','")
			sb.WriteString(store.URL(item))
		}
	}
	sb.WriteString("']")
	return sb.String()
	/*res := "["
	for i, item := range arr {
		if i == len(arr)-1 {
			res += "'" + store.URL(item) + "']"
		} else {
			res += "'" + store.URL(item) + "',"
		}
	}
	return res*/
}

func splitFilter(value FieldModel) interface{} {
	return strings.Split(value.Value, ",")
}
