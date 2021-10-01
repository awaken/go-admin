package controller

import (
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestIsInfoUrl(t *testing.T) {
	u := "https://localhost:8098/admin/info/user?id=sdfs"
	assert.Equal(t, true, utils.IsInfoUrl(u))
}

func TestIsNewUrl(t *testing.T) {
	u := "https://localhost:8098/admin/info/user/new?id=sdfs"
	assert.Equal(t, true, utils.IsNewUrl(u, "user"))
}
