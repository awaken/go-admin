package display

import (
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/template/types"
	"html/template"
)

type Dot struct {
	types.BaseDisplayFnGenerator
}

func init() {
	types.RegisterDisplayFnGenerator("dot", new(Dot))
}

func (d *Dot) Get(args ...interface{}) types.FieldFilterFn {
	icons := args[0].(map[string]types.FieldDotColor)
	defaultDot := types.FieldDotColor("")
	if len(args) > 1 {
		defaultDot = args[1].(types.FieldDotColor)
	}

	return func(value types.FieldModel) interface{} {
		if style, ok := icons[value.Value]; ok {
			return template.HTML(utils.StrConcat(`<span class="label-`, string(style),
				`" style="width: 8px;height: 8px;padding: 0;border-radius: 50%;display: inline-block;"></span>&nbsp;&nbsp;`, value.Value))
		}
		if defaultDot != "" {
			return template.HTML(utils.StrConcat(`<span class="label-`, string(defaultDot),
				`" style="width: 8px;height: 8px;padding: 0;border-radius: 50%;display: inline-block;"></span>&nbsp;&nbsp;`, value.Value))
		}
		return value.Value
	}
}
