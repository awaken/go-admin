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
	defaultDot := ""
	if len(args) > 0 {
		defaultDot = string(args[1].(types.FieldDotColor))
	}

	if defaultDot == "" {
		return func(value types.FieldModel) interface{} {
			if style, ok := icons[value.Value]; ok {
				return template.HTML(utils.StrConcat(`<span class="label-`, string(style),
					`" style="width: 8px;height: 8px;padding: 0;border-radius: 50%;display: inline-block;"></span>&nbsp;&nbsp;`, value.Value))
			}
			return value.Value
		}
	}

	defaultDot = utils.StrConcat(`<span class="label-`, defaultDot,
		`" style="width: 8px;height: 8px;padding: 0;border-radius: 50%;display: inline-block;"></span>&nbsp;&nbsp;`)

	return func(value types.FieldModel) interface{} {
		if style, ok := icons[value.Value]; ok {
			return template.HTML(utils.StrConcat(`<span class="label-`, string(style),
				`" style="width: 8px;height: 8px;padding: 0;border-radius: 50%;display: inline-block;"></span>&nbsp;&nbsp;`, value.Value))
		}
		return template.HTML(defaultDot + value.Value)
	}
}
