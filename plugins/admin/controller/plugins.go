package controller

import (
	"net/http"
	"time"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/remote_server"
	"github.com/GoAdminGroup/go-admin/plugins"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/guard"
	"github.com/gin-gonic/gin"
)

type PluginBoxParam struct {
	Info     plugins.Info
	Name     string
	IndexURL string
}

func (h *Handler) ServerLogin(ctx *context.Context) {
	param := guard.GetServerLoginParam(ctx)
	res := remote_server.Login(param.Account, param.Password)
	if res.Code == 0 && res.Data.Token != "" {
		ctx.SetCookie(&http.Cookie{
			Name:     remote_server.TokenKey,
			Value:    res.Data.Token,
			Expires:  time.Now().Add(time.Second * time.Duration(res.Data.Expire/1000)),
			HttpOnly: true,
			Path:     "/",
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code": res.Code,
		"data": res.Data,
		"msg":  res.Msg,
	})
}

func plugWord(word string) string {
	return language.GetWithScope(word, "plugin")
}
