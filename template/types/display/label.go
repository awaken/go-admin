package display

import (
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
)

type Label struct {
	types.BaseDisplayFnGenerator
}

func init() {
	types.RegisterDisplayFnGenerator("label", new(Label))
}

func (label *Label) Get(args ...interface{}) types.FieldFilterFn {
	params := args[0].([]types.FieldLabelParam)
	switch len(params) {
	case 0:
		return func(value types.FieldModel) interface{} {
			return template.Default().Label().SetContent(template.HTML(value.Value)).
				SetType("success").GetContent()
		}
	case 1:
		color := params[0].Color
		typ   := params[0].Type
		return func(value types.FieldModel) interface{} {
			return template.Default().Label().SetContent(template.HTML(value.Value)).
				SetColor(color).SetType(typ).GetContent()
		}
	}
	return func(value types.FieldModel) interface{} {
		return ""
	}
}
