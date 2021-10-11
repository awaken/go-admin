package utils

import (
	"archive/zip"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/NebulousLabs/fastrand"
	"html/template"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	textTmpl "text/template"
	"time"
)

var uuidAlphabet = "1234567890abcdefghijvklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func Uuid(length int64) string {
	uuid := Random([]byte(uuidAlphabet))
	for i := int64(0); i < length; i++ {
		uuid[i] = uuid[fastrand.Intn(59)]
	}
	return string(uuid)
}

func Random(buf []byte) []byte {
	for i := len(buf) - 1; i > 0; i-- {
		num := fastrand.Intn(i + 1)
		buf[i], buf[num] = buf[num], buf[i]
	}
	return buf
}

func CompressedContent(h *template.HTML) {
	st := strings.Split(string(*h), "\n")
	ss := make([]string, 0, len(st))
	for _, s := range st {
		s = strings.TrimSpace(s)
		if s != "" {
			ss = append(ss, s)
		}
	}
	*h = template.HTML(strings.Join(ss, "\n"))
}

func ReplaceNth(s, old, new string, n int) string {
	i := 0
	for m := 1; m <= n; m++ {
		x := strings.Index(s[i:], old)
		if x < 0 {
			break
		}
		i += x
		if m == n {
			return s[:i] + new + s[i+len(old):]
		}
		i += len(old)
	}
	return s
}

func InArray(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

func WrapURL(u string) string {
	if p := strings.IndexByte(u, '?'); p >= 0 && p < len(u) - 1 {
		v, err := url.ParseQuery(u[p + 1:])
		if err == nil {
			u  = url.QueryEscape(strings.ReplaceAll(u[:p], "/", "_"))
			w := strings.ReplaceAll(v.Encode(), "%7B%7B.Id%7D%7D", "{{.Id}}")
			var sb strings.Builder
			sb.Grow(len(u) + 1 + len(w))
			sb.WriteString(u)
			sb.WriteByte('?')
			sb.WriteString(w)
			return sb.String()
		}
	}
	return url.QueryEscape(strings.ReplaceAll(u, "/", "_"))
	/*uarr := strings.Split(u, "?")
	if len(uarr) < 2 {
		return url.QueryEscape(strings.ReplaceAll(u, "/", "_"))
	}
	v, err := url.ParseQuery(uarr[1])
	if err != nil {
		return url.QueryEscape(strings.ReplaceAll(u, "/", "_"))
	}
	return url.QueryEscape(strings.ReplaceAll(uarr[0], "/", "_")) + "?" +
		strings.ReplaceAll(v.Encode(), "%7B%7B.Id%7D%7D", "{{.Id}}")*/
}

func JSON(a interface{}) string {
	if a == nil {
		return ""
	}
	b, _ := json.Marshal(a)
	return string(b)
}

func ParseBool(s string) bool {
	b1, _ := strconv.ParseBool(s)
	return b1
}

func ParseFloat32(f string) float32 {
	s, _ := strconv.ParseFloat(f, 32)
	return float32(s)
}

func SetDefault(value, condition, def string) string {
	if value == condition {
		return def
	}
	return value
}

func AorB(condition bool, a, b string) string {
	if condition {
		return a
	}
	return b
}

func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func CopyMap(m map[string]string) map[string]string {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(m)
	if err != nil {
		panic(err)
	}
	var cm map[string]string
	err = dec.Decode(&cm)
	if err != nil {
		panic(err)
	}
	return cm
}

func ParseTime(stringTime string) time.Time {
	loc, _ := time.LoadLocation("UTC")
	theTime, _ := time.ParseInLocation("2006-01-02 15:04:05", stringTime, loc)
	return theTime
}

func ParseHTML(name, tmpl string, param interface{}) template.HTML {
	t := template.New(name)
	t, err := t.Parse(tmpl)
	if err != nil {
		fmt.Println("utils parseHTML error", err)
		return ""
	}
	var sb strings.Builder
	err = t.Execute(&sb, param)
	if err != nil {
		fmt.Println("utils parseHTML error", err)
		return ""
	}
	return template.HTML(sb.String())
}

func ParseText(name, tmpl string, param interface{}) string {
	t := textTmpl.New(name)
	t, err := t.Parse(tmpl)
	if err != nil {
		fmt.Println("utils parseHTML error", err)
		return ""
	}
	var sb strings.Builder
	err = t.Execute(&sb, param)
	if err != nil {
		fmt.Println("utils parseHTML error", err)
		return ""
	}
	return sb.String()
}

func CompareVersion(src, toCompare string) bool {
	if toCompare == "" {
		return false
	}

	src = rexCompareVersion.ReplaceAllString(src, "")
	toCompare = rexCompareVersion.ReplaceAllString(toCompare, "")

	src0, src1 := StrSplitByte2(src, 'v')
	srcArr := strings.Split(src1, ".")
	op := ">"
	src0 = strings.TrimSpace(src0)
	switch src0 {
	case ">=", "<=", "=", ">", "<": op = src0
	}
	/*srcs := strings.Split(src, "v")
	srcArr := strings.Split(srcs[1], ".")
	op := ">"
	srcs[0] = strings.TrimSpace(srcs[0])
	if InArray([]string{">=", "<=", "=", ">", "<"}, srcs[0]) {
		op = srcs[0]
	}*/

	toCompare = strings.ReplaceAll(toCompare, "v", "")

	if op == "=" {
		return src1 == toCompare
		//return srcs[1] == toCompare
	}

	//if srcs[1] == toCompare && (op == "<=" || op == ">=") {
	if src1 == toCompare && (op == "<=" || op == ">=") {
		return true
	}

	toCompareArr := strings.Split(strings.ReplaceAll(toCompare, "v", ""), ".")
	for i, s := range srcArr {
		v, err := strconv.Atoi(s)
		if err != nil {
			return false
		}
		vv, err := strconv.Atoi(toCompareArr[i])
		if err != nil {
			return false
		}
		switch op {
		case ">", ">=":
			if v < vv {
				return true
			} else if v > vv {
				return false
			} else {
				continue
			}
		case "<", "<=":
			if v > vv {
				return true
			} else if v < vv {
				return false
			} else {
				continue
			}
		}
	}

	return false
}

const (
	Byte  = 1
	KByte = Byte * 1024
	MByte = KByte * 1024
	GByte = MByte * 1024
	TByte = GByte * 1024
	PByte = TByte * 1024
	EByte = PByte * 1024
)

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanateBytes(s uint64, base float64, sizes []string) string {
	if s < 1024 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := float64(s) / math.Pow(base, math.Floor(e))
	f := "%.0f"
	if val < 10 {
		f = "%.1f"
	}

	return fmt.Sprintf(f + " %s", val, suffix)
}

var fileSizes = []string{ "B", "KB", "MB", "GB", "TB", "PB", "EB" }

// FileSize calculates the file size and generate user-friendly string.
func FileSize(s uint64) string {
	return humanateBytes(s, 1024, fileSizes)
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// TimeSincePro calculates the time interval and generate full user-friendly string.
func TimeSincePro(then time.Time, m map[string]string) string {
	now  := time.Now()
	diff := now.Unix() - then.Unix()

	if then.After(now) {
		return "future"
	}

	var timeStr, diffStr string
	for {
		if diff == 0 {
			break
		}

		diff, diffStr = computeTimeDiff(diff, m)
		timeStr += ", " + diffStr
	}
	return strings.TrimPrefix(timeStr, ", ")
}

// Seconds-based time units
const (
	Minute = 60
	Hour   = 60 * Minute
	Day    = 24 * Hour
	Week   = 7 * Day
	Month  = 30 * Day
	Year   = 12 * Month
)

func computeTimeDiff(diff int64, m map[string]string) (int64, string) {
	diffStr := ""
	switch {
	case diff <= 0:
		diff = 0
		diffStr = "now"
	case diff < 2:
		diff = 0
		diffStr = "1 " + m["second"]
	case diff < 1*Minute:
		diffStr = fmt.Sprintf("%d "+m["seconds"], diff)
		diff = 0

	case diff < 2*Minute:
		diff -= 1 * Minute
		diffStr = "1 " + m["minute"]
	case diff < 1*Hour:
		diffStr = fmt.Sprintf("%d "+m["minutes"], diff/Minute)
		diff -= diff / Minute * Minute

	case diff < 2*Hour:
		diff -= 1 * Hour
		diffStr = "1 " + m["hour"]
	case diff < 1*Day:
		diffStr = fmt.Sprintf("%d "+m["hours"], diff/Hour)
		diff -= diff / Hour * Hour

	case diff < 2*Day:
		diff -= 1 * Day
		diffStr = "1 " + m["day"]
	case diff < 1*Week:
		diffStr = fmt.Sprintf("%d "+m["days"], diff/Day)
		diff -= diff / Day * Day

	case diff < 2*Week:
		diff -= 1 * Week
		diffStr = "1 " + m["week"]
	case diff < 1*Month:
		diffStr = fmt.Sprintf("%d "+m["weeks"], diff/Week)
		diff -= diff / Week * Week

	case diff < 2*Month:
		diff -= 1 * Month
		diffStr = "1 " + m["month"]
	case diff < 1*Year:
		diffStr = fmt.Sprintf("%d "+m["months"], diff/Month)
		diff -= diff / Month * Month

	case diff < 2*Year:
		diff -= 1 * Year
		diffStr = "1 " + m["year"]
	default:
		diffStr = fmt.Sprintf("%d "+m["years"], diff/Year)
		diff = 0
	}
	return diff, diffStr
}

func DownloadTo(url, output string) (err error) {
	var req *http.Request
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	var res *http.Response
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()

	var file *os.File
	file, err = os.Create(output)
	if err != nil {
		return
	}

	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = e
		}
	}()

	_, err = io.Copy(file, res.Body)
	return
}

func UnzipDir(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	err = os.MkdirAll(dest, 0750)
	if err != nil {
		return err
	}

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(path, f.Mode())
			if err != nil {
				return err
			}
		} else {
			err = os.MkdirAll(filepath.Dir(path), f.Mode())
			if err != nil {
				return err
			}
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
