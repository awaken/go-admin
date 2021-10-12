package display

import (
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"html/template"

	"github.com/GoAdminGroup/go-admin/template/types"
)

type Downloadable struct {
	types.BaseDisplayFnGenerator
}

func init() {
	types.RegisterDisplayFnGenerator("downloadable", new(Downloadable))
}

func (d *Downloadable) Get(args ...interface{}) types.FieldFilterFn {
	var prefix string
	params := args[0].([]string)
	if len(params) > 0 { prefix = params[0] }

	return func(value types.FieldModel) interface{} {
		return template.HTML(utils.StrConcat(`<a href="`, prefix, value.Value, `" download="`, value.Value,
			`" target="_blank" class="text-muted"><i class="fa fa-download"></i>`, value.Value, `</a>`))
	}
}
