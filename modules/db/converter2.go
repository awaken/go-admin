package db

import (
	"regexp"
)

var _rexIsoDate = regexp.MustCompile("^([0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9])T([0-9][0-9]:[0-9][0-9]:[0-9][0-9])(?:\\.[0-9]*)?(?:Z|[-+][0-9][0-9]:?[0-9][0-9])$")

func fixIsoDateStr(s string) string {
	m := _rexIsoDate.FindStringSubmatch(s)
	if len(m) >= 3 {
		return m[1] + " " + m[2]
	}
	return s
}
