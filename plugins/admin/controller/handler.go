package controller

import (
	"github.com/GoAdminGroup/go-admin/modules/utils"
	template2 "html/template"
	"runtime/debug"
	"strings"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/response"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
)

// GlobalDeferHandler is a global error handler of admin plugin.
func (h *Handler) GlobalDeferHandler(ctx *context.Context) {
	logger.Access(ctx)

	if !h.config.OperationLogOff {
		h.RecordOperationLog(ctx)
	}

	if r := recover(); r != nil {
		logger.Error(r)
		logger.Error(string(debug.Stack()))
		errMsg := utils.RecoveryToMsg(r)

		if ctx.WantJSON() {
			response.Error(ctx, errMsg)
			return
		}

		ctxPath := utils.UrlWithoutQuery(ctx.Path()[len(h.config.UrlPrefix):])
		for _, action := range [...]string{ "/edit", "/new" } {
			if strings.HasPrefix(ctxPath, action) || strings.HasSuffix(ctxPath, action) {
				h.setFormWithReturnErrMessage(ctx, errMsg, action[1:])
				return
			}
		}

		h.HTML(ctx, auth.Auth(ctx), template.WarningPanelWithDescAndTitle(errMsg, errors.Msg, errors.Msg))
	}
}

func (h *Handler) setFormWithReturnErrMessage(ctx *context.Context, errMsg, kind string) {
	var (
		formInfo table.FormInfo
		f        *types.FormPanel
		btnWord  template2.HTML
		prefix   = ctx.Query(constant.PrefixKey)
		panel    = h.table(prefix, ctx)
		info     = panel.GetInfo()
	)

	if kind == "edit" {
		id := ctx.Query("id")
		switch id {
		case "":
			if ctx.Request.MultipartForm == nil { break }
			id = ctx.Request.MultipartForm.Value[panel.GetPrimaryKey().Name][0]
			fallthrough
		default:
			f = panel.GetForm()
			formInfo, _ = panel.GetDataWithId(parameter.GetParam(ctx.Request.URL, info.DefaultPageSize, info.SortField, info.GetSort()).WithPKs(id))
			btnWord = f.FormEditBtnWord
		}
	}
	if f == nil {
		f = panel.GetActualNewForm()
		formInfo = panel.GetNewFormInfo()
		formInfo.Title = f.Title
		formInfo.Description = f.Description
		btnWord = f.FormNewBtnWord
	}

	queryParam := parameter.GetParam(ctx.Request.URL, info.DefaultPageSize, info.SortField, info.GetSort()).GetRouteParamStr()

	h.HTML(ctx, auth.Auth(ctx), types.Panel{
		Content: aAlert().Warning(errMsg) + formContent(aForm().
			SetContent(formInfo.FieldList).
			SetTabContents(formInfo.GroupFieldList).
			SetTabHeaders(formInfo.GroupFieldHeaders).
			SetTitle(template2.HTML(strings.Title(kind))).
			SetPrimaryKey(panel.GetPrimaryKey().Name).
			SetPrefix(h.config.PrefixFixSlash()).
			SetHiddenFields(map[string]string{
				form.TokenKey:    h.authSrv().AddToken(),
				form.PreviousKey: h.config.Url(utils.StrConcat("/info/", prefix, queryParam)),
			}).
			SetUrl(h.config.Url(utils.StrConcat("/", kind, "/", prefix))).
			SetOperationFooter(formFooter(kind, f.IsHideContinueEditCheckBox, f.IsHideContinueNewCheckBox,
				f.IsHideResetButton, btnWord)).
			SetHeader(f.HeaderHtml).
			SetFooter(f.FooterHtml), len(formInfo.GroupFieldHeaders) > 0,
			ctx.IsIframe(),
			f.IsHideBackButton, f.Header),
		Description: template2.HTML(formInfo.Description),
		Title:       template2.HTML(formInfo.Title),
	})

	ctx.AddHeader(constant.PjaxUrlHeader, h.config.Url(utils.StrConcat("/info/", prefix, "/", kind, queryParam)))
}
