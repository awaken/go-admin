package types

import (
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/template/types/form"
	"strings"
)

var (
	HttpMethodFieldOptions = FieldOptions{
		{ Value: "GET"    , Text: "GET"     },
		{ Value: "PUT"    , Text: "PUT"     },
		{ Value: "POST"   , Text: "POST"    },
		{ Value: "DELETE" , Text: "DELETE"  },
		{ Value: "PATCH"  , Text: "PATCH"   },
		{ Value: "OPTIONS", Text: "OPTIONS" },
		{ Value: "HEAD"   , Text: "HEAD"    },
	}
)

func BoolFieldOptions() FieldOptions {
	return FieldOptions{
		{ Text: language.Get("Yes"), Value: models.StrTrue  },
		{ Text: language.Get("No" ), Value: models.StrFalse },
	}
}

func BoolFilterType() FilterType {
	return FilterType{
		FormType: form.SelectSingle,
		Options : BoolFieldOptions(),
	}
}

func BoolFieldDisplay(model FieldModel) interface{} {
	if model.Value == models.StrTrue { return language.Get("Yes") }
	return language.Get("No")
}

func BoolOptionalFieldDisplay(model FieldModel) interface{} {
	switch model.Value {
	case ""            : return ""
	case models.StrTrue: return language.Get("Yes")
	}
	return language.Get("No")
}

func NoopFieldDisplay(value FieldModel) interface{} {
	return value.Value
}

func EmptyFieldDisplay(_ FieldModel) interface{} {
	return ""
}

func CommaSplitFieldDisplay(model FieldModel) interface{} {
	return strings.Split(model.Value, ",")
}

func CommaSplitPostFilter(model PostFieldModel) interface{} {
	return strings.Join(model.Value, ",")
}

func TrimPostFilter(model PostFieldModel) interface{} {
	return strings.TrimSpace(model.Value.Value())
}
