package types

import "html/template"

type FilterOperator string

const (
	FilterOperatorLike           FilterOperator = "LIKE"
	FilterOperatorGreater        FilterOperator = ">"
	FilterOperatorGreaterOrEqual FilterOperator = ">="
	FilterOperatorEqual          FilterOperator = "="
	FilterOperatorNotEqual       FilterOperator = "!="
	FilterOperatorLess           FilterOperator = "<"
	FilterOperatorLessOrEqual    FilterOperator = "<="
	FilterOperatorFree           FilterOperator = "FREE"
)

func GetOperatorFromValue(value string) FilterOperator {
	switch value {
	case "LIKE", "like":
		return FilterOperatorLike
	case "GR", "gr":
		return FilterOperatorGreater
	case "GQ", "gq":
		return FilterOperatorGreaterOrEqual
	case "EQ", "eq":
		return FilterOperatorEqual
	case "NE", "ne":
		return FilterOperatorNotEqual
	case "LE", "le":
		return FilterOperatorLess
	case "LQ", "lq":
		return FilterOperatorLessOrEqual
	case "FREE", "free":
		return FilterOperatorFree
	default:
		return FilterOperatorEqual
	}
}

func (o FilterOperator) Value() string {
	switch o {
	case FilterOperatorLike:
		return "LIKE"
	case FilterOperatorGreater:
		return "GR"
	case FilterOperatorGreaterOrEqual:
		return "GQ"
	case FilterOperatorEqual:
		return "EQ"
	case FilterOperatorNotEqual:
		return "NE"
	case FilterOperatorLess:
		return "LE"
	case FilterOperatorLessOrEqual:
		return "LQ"
	case FilterOperatorFree:
		return "FREE"
	default:
		return "EQ"
	}
}

func (o FilterOperator) String() string {
	return string(o)
}

func (o FilterOperator) Label() template.HTML {
	if o == FilterOperatorLike {
		return ""
	}
	return template.HTML(o)
}

func (o FilterOperator) AddOrNot() bool {
	return string(o) != "" && o != FilterOperatorFree
}

func (o FilterOperator) Valid() bool {
	switch o {
	case FilterOperatorLike, FilterOperatorGreater, FilterOperatorGreaterOrEqual,
		FilterOperatorLess, FilterOperatorLessOrEqual, FilterOperatorFree:
		return true
	default:
		return false
	}
}
