package modules

import (
	uuid "github.com/satori/go.uuid"
	"strconv"
	"strings"
)

func InArray(arr []string, str string) bool {
	for _, v := range arr {
		if v == str { return true }
	}
	return false
}

func Delimiter(del, del2, s string) string {
	var sb strings.Builder
	sb.Grow(len(del) + len(s) + len(del2))
	sb.WriteString(del)
	sb.WriteString(s)
	sb.WriteString(del2)
	return sb.String()
}

func FilterField(field, delimiter, delimiter2 string) string {
	var sb strings.Builder
	sb.Grow(len(delimiter) + len(field) + len(delimiter2))
	sb.WriteString(delimiter)
	sb.WriteString(field)
	sb.WriteString(delimiter2)
	return sb.String()
}

func InArrayWithoutEmpty(arr []string, str string) bool {
	if len(arr) == 0 { return true }
	for _, v := range arr {
		if v == str { return true }
	}
	return false
}

func RemoveBlankFromArray(s []string) []string {
	r := make([]string, 0, len(s))
	for _, str := range s {
		if str != "" { r = append(r, str) }
	}
	return r
}

func Uuid() string {
	u, _ := uuid.NewV4()
	return u.String()
}

func SetDefault(source, def string) string {
	if source == "" { return def }
	return source
}

func GetPage(page string) int {
	if page == "" { return 1 }
	pageInt, _ := strconv.Atoi(page)
	return pageInt
}

func AorEmpty(condition bool, a string) string {
	if condition { return a }
	return ""
}
