package utils

import (
	"regexp"
	"strings"
)

var (
	rexInfoUrl = regexp.MustCompile("(.*?)info/(.*?)$")
	rexFormUrl = regexp.MustCompile("info/(.*)/(new|edit)")

	rexTypeName  = regexp.MustCompile(`\(.*?\)`)
	rexTypeName2 = regexp.MustCompile(`unsigned(.*)`)

	rexCleanPassword = regexp.MustCompile(`("password[^"]*")\s*:\s*\[\s*".*?"\s*]`)
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
