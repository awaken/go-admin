package guard

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"io"
)

type ServerLoginParam struct {
	Account  string
	Password string
}

func (g *Guard) ServerLogin(ctx *context.Context) {
	var p ServerLoginParam

	body, err := io.ReadAll(ctx.Request.Body)

	if err != nil {
		logger.Error("get server login parameter error: ", err)
	}

	err = utils.JsonUnmarshal(body, &p)
	if err != nil {
		logger.Error("unmarshal server login parameter error: ", err)
	}

	ctx.SetUserValue(serverLoginParamKey, &p)
	ctx.Next()
}

func GetServerLoginParam(ctx *context.Context) *ServerLoginParam {
	return ctx.UserValue[serverLoginParamKey].(*ServerLoginParam)
}
