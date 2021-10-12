package display

import (
	"github.com/GoAdminGroup/go-admin/template/types"
	"html/template"
)

type Loading struct {
	types.BaseDisplayFnGenerator
}

func init() {
	types.RegisterDisplayFnGenerator("loading", new(Loading))
}

func (l *Loading) Get(args ...interface{}) types.FieldFilterFn {
	params := args[0].([]string)

	return func(value types.FieldModel) interface{} {
		for _, p := range params {
			if value.Value == p {
				return template.HTML(`<i class="fa fa-refresh fa-spin text-primary"></i>`)
			}
		}
		return value.Value
	}
}
