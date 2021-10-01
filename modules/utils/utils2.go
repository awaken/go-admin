package utils

import (
	"regexp"
	"strings"
)

var (
	rexCache Cache

	rexCompareVersion, rexInfoUrl, rexFormUrl, rexTypeName, rexTypeName2, rexMaskPassword, rexIsoDate *regexp.Regexp

	RexCommonQuery, RexSqlSelect, RexMenuActiveClass *regexp.Regexp

	logoutUrl string
)

func IsLogoutUrl(s string) bool {
	return s == logoutUrl
}

func IsInfoUrl(s string) bool {
	sub := rexInfoUrl.FindStringSubmatch(s)
	return len(sub) > 2 && !strings.Contains(sub[2], "/")
}

func IsNewUrl(s string, p string) bool {
	return strings.Contains(s, "info/" + p + "/new")
}

func IsEditUrl(s string, p string) bool {
	return strings.Contains(s, "info/" + p + "/edit")
}

func IsFormUrl(s string) bool {
	return rexFormUrl.MatchString(s)
}

func GetTypeName(typeName string) string {
	typeName = rexTypeName.ReplaceAllString(typeName, "")
	return strings.TrimSpace(strings.Title(strings.ToLower(rexTypeName2.ReplaceAllString(typeName, ""))))
}

func MaskContentToLog(input string) string {
	return rexMaskPassword.ReplaceAllString(input, `$1:["****"]`)
}

func StrIsoDateToDateTime(s string) string {
	if m := rexIsoDate.FindStringSubmatch(s); len(m) >= 3 {
		return m[1] + " " + m[2]
	}
	return s
}

func InitUtils(cacheSize int, urler func(string) string) {
	rexCache = MustNewCache(cacheSize)

	RexCommonQuery     = regexp.MustCompile(`\\((.*)\\)`)
	RexSqlSelect       = regexp.MustCompile(`(.*?)\((.*?)\)`)
	RexMenuActiveClass = regexp.MustCompile(`\?(.*)`)

	rexCompareVersion  = regexp.MustCompile(`-(.*)`)
	rexInfoUrl         = regexp.MustCompile(`(.*?)info/(.*?)$`)
	rexFormUrl         = regexp.MustCompile(`info/(.*)/(new|edit)`)
	rexTypeName        = regexp.MustCompile(`\(.*?\)`)
	rexTypeName2       = regexp.MustCompile(`unsigned(.*)`)
	rexMaskPassword    = regexp.MustCompile(`("password[^"]*")\s*:\s*\[\s*".*?"\s*]`)
	rexIsoDate         = regexp.MustCompile("^([0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9])T([0-9][0-9]:[0-9][0-9]:[0-9][0-9])(?:\\.[0-9]*)?(?:Z|[-+][0-9][0-9]:?[0-9][0-9])$")

	logoutUrl = urler("/logout")
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
