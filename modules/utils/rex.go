package utils

import (
	"regexp"
	"strings"
)

var (
	rexCache Cache

	rexCompareVersion, rexInfoUrl, rexFormUrl,	rexTypeName, rexTypeName2, rexCleanPassword, rexIsoDate *regexp.Regexp

	RexCommonQuery, RegSqlSelect, RexMenuActiveClass, RexPathEdit, RexPathNew, RexPathLogout *regexp.Regexp
)

func IsInfoUrl(s string) bool {
	sub := rexInfoUrl.FindStringSubmatch(s)
	return len(sub) > 2 && !strings.Contains(sub[2], "/")
}

func IsNewUrl(s string, p string) bool {
	reg, _ := CachedRex("info/" + p + "/new")
	return reg.MatchString(s)
}

func IsEditUrl(s string, p string) bool {
	reg, _ := CachedRex("info/" + p + "/edit")
	return reg.MatchString(s)
}

func IsFormUrl(s string) bool {
	return rexFormUrl.MatchString(s)
}

func GetTypeName(typeName string) string {
	typeName = rexTypeName.ReplaceAllString(typeName, "")
	return strings.TrimSpace(strings.Title(strings.ToLower(rexTypeName2.ReplaceAllString(typeName, ""))))
}

func CleanContentToLog(input string) string {
	return rexCleanPassword.ReplaceAllString(input, `$1:["****"]`)
}

var _ = regexp.MustCompile("^([0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9])T([0-9][0-9]:[0-9][0-9]:[0-9][0-9])(?:\\.[0-9]*)?(?:Z|[-+][0-9][0-9]:?[0-9][0-9])$")

func StrIsoDateToDateTime(s string) string {
	if m := rexIsoDate.FindStringSubmatch(s); len(m) >= 3 {
		return m[1] + " " + m[2]
	}
	return s
}

func InitUiRex(size int) {
	rexCache = MustNewCache(size)

	RexCommonQuery     = regexp.MustCompile(`\\((.*)\\)`)
	RegSqlSelect       = regexp.MustCompile(`(.*?)\((.*?)\)`)
	RexMenuActiveClass = regexp.MustCompile(`\?(.*)`)
	RexPathEdit        = regexp.MustCompile(`/edit`)
	RexPathNew         = regexp.MustCompile(`/new`)
	RexPathLogout      = regexp.MustCompile(`/logout`)

	rexCompareVersion  = regexp.MustCompile(`-(.*)`)
	rexInfoUrl         = regexp.MustCompile(`(.*?)info/(.*?)$`)
	rexFormUrl         = regexp.MustCompile(`info/(.*)/(new|edit)`)
	rexTypeName        = regexp.MustCompile(`\(.*?\)`)
	rexTypeName2       = regexp.MustCompile(`unsigned(.*)`)
	rexCleanPassword   = regexp.MustCompile(`("password[^"]*")\s*:\s*\[\s*".*?"\s*]`)
	rexIsoDate         = regexp.MustCompile("^([0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9])T([0-9][0-9]:[0-9][0-9]:[0-9][0-9])(?:\\.[0-9]*)?(?:Z|[-+][0-9][0-9]:?[0-9][0-9])$")
}

func CachedRex(rexStr string) (*regexp.Regexp, error) {
	if v, ok := rexCache.Get(rexStr); ok {
		return v.(*regexp.Regexp), nil
	}
	rex, err := regexp.Compile(rexStr)
	if err != nil {
		return nil, err
	}
	rexCache.Add(rexStr, rex)
	return rex, nil
}
