package utils

import (
	"fmt"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	jsoniter "github.com/json-iterator/go"
	"regexp"
	"strings"
	"time"
	"unsafe"
)

var (
	rexCache Cache

	rexCompareVersion, rexInfoUrl, rexFormUrl, rexTypeName, rexTypeName2, rexMaskPassword, rexIsoDate *regexp.Regexp

	RexCommonQuery, RexSqlSelect, RexMenuActiveClass *regexp.Regexp

	PkReplacer, TableFormReplacer, JsonTmplReplacer, JumpTmplReplacer, XssJsReplacer *strings.Replacer

	logoutUrl string

	DefaultExceptMap map[string]struct{}
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
		_ = m[2]
		return StrConcat(m[1], " ", m[2])
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
	XssJsReplacer      = strings.NewReplacer("<script>", "&lt;script&gt;", "</script>", "&lt;/script&gt;")

	DefaultExceptMap = map[string]struct{}{
		form.PreviousKey: {}, form.MethodKey: {}, form.TokenKey: {}, constant.IframeKey: {}, constant.IframeIDKey: {},
	}

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
	var buf []byte
	switch len(args) {
	case 0:
		return ""
	case 1:
		return args[0]
	case 2:
		_ = args[1]
		return args[0] + args[1]
	case 3:
		_ = args[2]
		buf = make([]byte, 0, len(args[0]) + len(args[1]) + len(args[2]))
		buf = append(append(append(buf, args[0]...), args[1]...), args[2]...)
	case 4:
		_ = args[3]
		buf = make([]byte, 0, len(args[0]) + len(args[1]) + len(args[2]) + len(args[3]))
		buf = append(append(append(append(buf, args[0]...), args[1]...), args[2]...), args[3]...)
	case 5:
		_ = args[4]
		buf = make([]byte, 0, len(args[0]) + len(args[1]) + len(args[2]) + len(args[3]) + len(args[4]))
		buf = append(append(append(append(append(buf, args[0]...), args[1]...), args[2]...), args[3]...), args[4]...)
	case 6:
		_ = args[5]
		buf = make([]byte, 0, len(args[0]) + len(args[1]) + len(args[2]) + len(args[3]) + len(args[4]) + len(args[5]))
		buf = append(append(append(append(append(append(buf, args[0]...), args[1]...), args[2]...), args[3]...), args[4]...), args[5]...)
	case 7:
		_ = args[6]
		buf = make([]byte, 0, len(args[0]) + len(args[1]) + len(args[2]) + len(args[3]) + len(args[4]) + len(args[5]) + len(args[6]))
		buf = append(append(append(append(append(append(append(buf, args[0]...), args[1]...), args[2]...), args[3]...), args[4]...), args[5]...), args[6]...)
	case 8:
		_ = args[7]
		buf = make([]byte, 0, len(args[0]) + len(args[1]) + len(args[2]) + len(args[3]) + len(args[4]) + len(args[5]) + len(args[6]) + len(args[7]))
		buf = append(append(append(append(append(append(append(append(buf, args[0]...), args[1]...), args[2]...), args[3]...), args[4]...), args[5]...), args[6]...), args[7]...)
	default:
		n := 0
		for _, s := range args { n += len(s) }
		buf = make([]byte, 0, n)
		for _, s := range args { buf = append(buf, s...) }
	}
	return *(*string)(unsafe.Pointer(&buf))
}

func UrlWithoutQuery(url string) string {
	if p := strings.IndexByte(url, '?'); p >= 0 {
		return url[:p]
	}
	return url
}

func StrSplit2(s, sep string) (string, string) {
	if p := strings.Index(s, sep); p >= 0 {
		return s[:p], s[p+1:]
	}
	return s, ""
}

func StrSplitByte2(s string, sep byte) (string, string) {
	if p := strings.IndexByte(s, sep); p >= 0 {
		return s[:p], s[p+1:]
	}
	return s, ""
}

func JsonMarshal(v interface{}) ([]byte, error) {
	return jsoniter.ConfigFastest.Marshal(v)
}

func JsonUnmarshal(data []byte, v interface{}) error {
	return jsoniter.ConfigFastest.Unmarshal(data, v)
}

func RecoveryToMsg(r interface{}) string {
	var msg string
	switch t := r.(type) {
	case string      : msg = t
	case error       : msg = t.Error()
	case fmt.Stringer: msg = t.String()
	}
	if msg == "" { return "system error" }
	return msg
}

func InMapT(m map[string]struct{}, key string) bool {
	_, ok := m[key]
	return ok
}
