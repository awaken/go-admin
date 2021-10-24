package display

import (
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
)

type Image struct {
	types.BaseDisplayFnGenerator
}

func init() {
	types.RegisterDisplayFnGenerator("image", new(Image))
}

func (image *Image) Get(args ...interface{}) types.FieldFilterFn {
	_ = args[2]
	width  := args[0].(string)
	height := args[1].(string)
	param  := args[2].([]string)
	prefix := param[0]

	return func(value types.FieldModel) interface{} {
		return template.Default().Image().SetWidth(width).SetHeight(height).
			SetSrc(template.HTML(prefix + value.Value)).GetContent()
	}
}
