package models

import (
	"strconv"
	"strings"
)

const varUserId = "${uid}"

func normMatchPath(matchPath string) string {
	return strings.ReplaceAll(matchPath, "/*", "/.*")
}

func (t UserModel) patchPathParams(path string) string {
	if strings.Contains(path, varUserId) {
		return strings.ReplaceAll(path, varUserId, strconv.Itoa(int(t.Id)))
	}
	return path
}

/*
func (t UserModel) isMyRequest(method, path string, params url.Values) bool {
	if strings.EqualFold(method, "POST") {
		if strings.HasSuffix(path, "/edit/normal_manager") {
			if p := params.Get("id"); p != "" {
				id, err := strconv.Atoi(p)
				return err == nil && id == int(t.Id)
			}
		}
	}
	return false
}
*/
