package display

import (
	"github.com/GoAdminGroup/go-admin/template/icon"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/html"
	"strings"
)

type Bool struct {
	types.BaseDisplayFnGenerator
}

func init() {
	types.RegisterDisplayFnGenerator("bool", new(Bool))
}

func (b *Bool) Get(args ...interface{}) types.FieldFilterFn {
	return func(value types.FieldModel) interface{} {
		pass := icon.IconWithStyle(icon.Check , html.Style{ "color": "green" })
		fail := icon.IconWithStyle(icon.Remove, html.Style{ "color": "red"   })
		params := args[0].([]string)
		switch len(params) {
		case 0:
			if value.Value == "0" || strings.ToLower(value.Value) == "false" { return fail }
			return pass
		case 1:
			if value.Value == params[0] { return pass }
			return fail
		}
		_ = params[1]
		switch value.Value {
		case params[0]: return pass
		case params[1]: return fail
		}
		return ""
	}
}
