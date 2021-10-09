package utils

import (
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"regexp"
	"strings"
	"time"
)

var (
	rexCache Cache

	rexCompareVersion, rexInfoUrl, rexFormUrl, rexTypeName, rexTypeName2, rexMaskPassword, rexIsoDate *regexp.Regexp

	RexCommonQuery, RexSqlSelect, RexMenuActiveClass *regexp.Regexp

	PkReplacer, TableFormReplacer, JsonTmplReplacer, JumpTmplReplacer *strings.Replacer

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

func NowStr() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05")
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

	PkReplacer         = strings.NewReplacer(constant.EditPKKey, "id", constant.DetailPKKey, "id")
	TableFormReplacer  = strings.NewReplacer("table/", "", "form/", "")
	JsonTmplReplacer   = strings.NewReplacer(`"{%id}"`, "{{.Id}}", `"{%ids}"`, "{{.Ids}}", `"{{.Ids}}"`, "{{.Ids}}", `"{{.Id}}"`, "{{.Id}}")
	JumpTmplReplacer   = strings.NewReplacer("{%id}", "{{.Id}}", "{%ids}", "{{.Ids}}")

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

func StrConcat(args ...string) string {
	var sb strings.Builder
	switch len(args) {
	case 0: return ""
	case 1: return args[0]
	case 2: return args[0] + args[1]
	case 3:
		sb.Grow(len(args[0]) + len(args[1]) + len(args[2]))
		sb.WriteString(args[0])
		sb.WriteString(args[1])
		sb.WriteString(args[2])
		return sb.String()
	case 4:
		sb.Grow(len(args[0]) + len(args[1]) + len(args[2]) + len(args[3]))
		sb.WriteString(args[0])
		sb.WriteString(args[1])
		sb.WriteString(args[2])
		sb.WriteString(args[3])
	case 5:
		sb.Grow(len(args[0]) + len(args[1]) + len(args[2]) + len(args[3]) + len(args[4]))
		sb.WriteString(args[0])
		sb.WriteString(args[1])
		sb.WriteString(args[2])
		sb.WriteString(args[3])
		sb.WriteString(args[4])
	default:
		ss := args[5:]
		n  := len(args[0]) + len(args[1]) + len(args[2]) + len(args[3]) + len(args[4])
		for _, s := range ss { n += len(s) }
		sb.Grow(n)
		sb.WriteString(args[0])
		sb.WriteString(args[1])
		sb.WriteString(args[2])
		sb.WriteString(args[3])
		sb.WriteString(args[4])
		for _, s := range ss { sb.WriteString(s) }
	}
	return sb.String()
}
