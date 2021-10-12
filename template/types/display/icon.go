package display

import (
	"github.com/GoAdminGroup/go-admin/template/icon"
	"github.com/GoAdminGroup/go-admin/template/types"
)

type Icon struct {
	types.BaseDisplayFnGenerator
}

func init() {
	types.RegisterDisplayFnGenerator("icon", new(Icon))
}

func (i *Icon) Get(args ...interface{}) types.FieldFilterFn {
	icons := args[0].(map[string]string)
	defaultIcon := ""
	if len(args) > 1 {
		defaultIcon = args[1].(string)
	}

	return func(value types.FieldModel) interface{} {
		if iconClass, ok := icons[value.Value]; ok {
			return icon.Icon(iconClass)
		}
		if defaultIcon != "" {
			return icon.Icon(defaultIcon)
		}
		return value.Value
	}
}
