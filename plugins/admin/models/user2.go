package models

import (
	"net/url"
	"strconv"
	"strings"
)

const (
	StrTrue  = "y"
	StrFalse = "n"

	UserDisabledValue = StrTrue
	UserEnabledValue  = StrFalse

	varUserId = "${uid}"
)

func normMatchPath(matchPath string) string {
	return strings.ReplaceAll(matchPath, "/*", "/.*")
}

func normStrBool(s, def string) string {
	if len(s) > 0 {
		switch s[0] {
		case 'y', 'Y', 't', 'T', '1': return StrTrue
		case 'n', 'N', 'f', 'F', '0': return StrFalse
		}
	}
	return def
}

func normUserDisabled(s string) string {
	return normStrBool(s, UserEnabledValue)
}

func normUserRoot(s string) string {
	return normStrBool(s, StrFalse)
}

func (t UserModel) patchPathParams(path string) string {
	if strings.Contains(path, varUserId) {
		return strings.ReplaceAll(path, varUserId, strconv.Itoa(int(t.Id)))
	}
	return path
}

func (t UserModel) IsDisabled() bool {
	return t.Disabled == UserDisabledValue
}

func (t UserModel) IsRootAdmin() bool {
	return t.Root == StrTrue
}

func (t UserModel) isMySettingRequest(method, path string, params url.Values) bool {
	return UserIdToEdit(method, path, params) == t.Id
}

func UserIdToEdit(method, path string, params url.Values) int64 {
	const editPath = "/edit/"
	if strings.EqualFold(method, "POST") {
		if i := strings.Index(path, editPath); i >= 0 {
			if i += len(editPath); i < len(path) {
				switch path[i:] {
				case "manager", "normal_manager":
					if p := params.Get("id"); p != "" {
						if id, err := strconv.Atoi(p); err == nil {
							return int64(id)
						}
					}
				}
			}
		}
	}
	return 0
}
