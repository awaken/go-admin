package guard

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"html/template"
	"mime/multipart"
)

type ShowNewFormParam struct {
	Panel  table.Table
	Prefix string
	Param  parameter.Parameters
}

func (g *Guard) ShowNewForm(ctx *context.Context) {
	panel, prefix := g.table(ctx)

	if !panel.GetCanAdd() {
		alert(ctx, panel, errors.OperationNotAllow, g.conn, g.navBtns)
		ctx.Abort()
		return
	}

	if panel.GetOnlyInfo() {
		ctx.Redirect(config.Url("/info/" + prefix))
		ctx.Abort()
		return
	}

	if panel.GetOnlyDetail() {
		ctx.Redirect(config.Url("/info/" + prefix + "/detail"))
		ctx.Abort()
		return
	}

	if panel.GetOnlyUpdateForm() {
		ctx.Redirect(config.Url("/info/" + prefix + "/edit"))
		ctx.Abort()
		return
	}

	info := panel.GetInfo()

	ctx.SetUserValue(showNewFormParam, &ShowNewFormParam{
		Panel:  panel,
		Prefix: prefix,
		Param: parameter.GetParam(ctx.Request.URL, info.DefaultPageSize, info.SortField, info.GetSort()),
	})
	ctx.Next()
}

func GetShowNewFormParam(ctx *context.Context) *ShowNewFormParam {
	return ctx.UserValue[showNewFormParam].(*ShowNewFormParam)
}

type NewFormParam struct {
	Panel        table.Table
	Id           string
	Prefix       string
	Param        parameter.Parameters
	Path         string
	MultiForm    *multipart.Form
	PreviousPath string
	FromList     bool
	IsIframe     bool
	IframeID     string
	Alert        template.HTML
}

func (e NewFormParam) Value() form.Values {
	return e.MultiForm.Value
}

func (g *Guard) NewForm(ctx *context.Context) {
	var (
		previous      = ctx.FormValue(form.PreviousKey)
		panel, prefix = g.table(ctx)
		conn          = db.GetConnection(g.services)
		token         = ctx.FormValue(form.TokenKey)
	)

	if !auth.GetTokenService(g.services.Get(auth.TokenServiceKey)).CheckToken(token) {
		alert(ctx, panel, errors.CreateFailWrongToken, conn, g.navBtns)
		ctx.Abort()
		return
	}

	fromList  := utils.IsInfoUrl(previous)
	info := panel.GetInfo()
	param := parameter.GetParamFromURL(previous, info.DefaultPageSize, info.GetSort(), panel.GetPrimaryKey().Name)

	if fromList {
		previous = config.Url("/info/" + prefix + param.GetRouteParamStr())
	}

	multiForm := ctx.Request.MultipartForm
	values := multiForm.Value

	ctx.SetUserValue(newFormParamKey, &NewFormParam{
		Panel:        panel,
		Id:           "",
		Prefix:       prefix,
		Param:        param,
		IsIframe:     form.Values(values).Get(constant.IframeKey) == "true",
		IframeID:     form.Values(values).Get(constant.IframeIDKey),
		Path:         utils.UrlWithoutQuery(previous),
		MultiForm:    multiForm,
		PreviousPath: previous,
		FromList:     fromList,
	})
	ctx.Next()
}

func GetNewFormParam(ctx *context.Context) *NewFormParam {
	return ctx.UserValue[newFormParamKey].(*NewFormParam)
}
