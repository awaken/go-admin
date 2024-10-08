package controller

import (
	"fmt"
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/file"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	form2 "github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/guard"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/response"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
	"net/http"
)

// ShowForm show form page.
func (h *Handler) ShowForm(ctx *context.Context) {
	param := guard.GetShowFormParam(ctx)
	h.showForm(ctx, "", param.Prefix, param.Param, false)
}

func (h *Handler) showForm(ctx *context.Context, alert template.HTML, prefix string, param parameter.Parameters, isEdit bool, animation ...bool) {
	panel := h.table(prefix, ctx)
	f     := panel.GetForm()

	if f.HasError() {
		if f.PageErrorHTML != "" {
			h.HTML(ctx, auth.Auth(ctx),
				types.Panel{ Content: f.PageErrorHTML}, template.ExecuteOptions{ Animation: param.Animation })
			return
		}
		h.HTML(ctx, auth.Auth(ctx),
			template.WarningPanel(f.PageError.Error(),
			template.GetPageTypeFromPageError(f.PageError)), template.ExecuteOptions{ Animation: param.Animation })
		return
	}

	var (
		user       = auth.Auth(ctx)
		paramStr   = param.GetRouteParamStr()
		footerKind = "edit"
		isAnim     = alert == "" || len(animation) > 0 && animation[0]
		newUrl     string
	)

	if panel.GetCanAdd() {
		newUrl = h.routePathWithPrefix("show_new", prefix) + paramStr
	}
	if newUrl == "" || !user.CheckPermissionByUrlMethod(newUrl, h.route("show_new").Method(), nil) {
		footerKind = "edit_only"
	}

	formInfo, err := panel.GetDataWithId(param)

	if err != nil {
		logger.Error("receive data error: ", err)
		h.HTML(ctx, user, template.
			WarningPanelWithDescAndTitle(err.Error(), f.Description, f.Title),
			template.ExecuteOptions{ Animation: isAnim })
		if isEdit {
			ctx.AddHeader(constant.PjaxUrlHeader, h.routePathWithPrefix("show_edit", prefix) +
				param.DeletePK().GetRouteParamStr())
		}
		return
	}

	showEditUrl := h.routePathWithPrefix("show_edit", prefix) + param.DeletePK().GetRouteParamStr()
	infoUrl     := h.routePathWithPrefix("info", prefix) + param.DeleteField(constant.EditPKKey).GetRouteParamStr()
	editUrl     := h.routePathWithPrefix("edit", prefix)
	referer     := ctx.Referer()

	if referer != "" && !utils.IsInfoUrl(referer) && !utils.IsEditUrl(referer, ctx.Query(constant.PrefixKey)) {
		infoUrl = referer
	}

	isNotIframe := ctx.Query(constant.IframeKey) != "true"

	hiddenFields := map[string]string{
		form2.TokenKey   : h.authSrv().AddToken(),
		form2.PreviousKey: infoUrl,
	}

	if ctx.Query(constant.IframeKey) != "" {
		hiddenFields[constant.IframeKey] = ctx.Query(constant.IframeKey)
	}

	if ctx.Query(constant.IframeIDKey) != "" {
		hiddenFields[constant.IframeIDKey] = ctx.Query(constant.IframeIDKey)
	}

	content := formContent(aForm().
		SetContent(formInfo.FieldList).
		SetFieldsHTML(f.HTMLContent).
		SetTabContents(formInfo.GroupFieldList).
		SetTabHeaders(formInfo.GroupFieldHeaders).
		SetPrefix(h.config.PrefixFixSlash()).
		SetInputWidth(f.InputWidth).
		SetHeadWidth(f.HeadWidth).
		SetPrimaryKey(panel.GetPrimaryKey().Name).
		SetUrl(editUrl).
		SetTitle(f.FormEditTitle).
		SetAjax(f.AjaxSuccessJS, f.AjaxErrorJS).
		SetLayout(f.Layout).
		SetHiddenFields(hiddenFields).
		SetOperationFooter(formFooter(footerKind,
			f.IsHideContinueEditCheckBox,
			f.IsHideContinueNewCheckBox,
			f.IsHideResetButton, f.FormEditBtnWord)).
		SetHeader(f.HeaderHtml).
		SetFooter(f.FooterHtml), len(formInfo.GroupFieldHeaders) > 0, !isNotIframe, f.IsHideBackButton, f.Header)

	if f.Wrapper != nil {
		content = f.Wrapper(content)
	}

	title := ""
	if isNotIframe { title = formInfo.Title }

	h.HTML(ctx, user, types.Panel{
		Content    : alert + content,
		Description: template.HTML(formInfo.Description),
		Title      : template.HTML(title),
		MiniSidebar: f.HideSideBar,
	}, template.ExecuteOptions{ Animation : isAnim, NoCompress: f.NoCompress })

	if isEdit {
		ctx.AddHeader(constant.PjaxUrlHeader, showEditUrl)
	}
}

func (h *Handler) EditForm(ctx *context.Context) {
	param := guard.GetEditFormParam(ctx)
	pmf   := param.MultiForm.File

	if len(pmf) > 0 {
		err := file.GetFileEngine(h.config.FileUploadEngine.Name).Upload(param.MultiForm)
		if err != nil {
			logger.Error("get file engine error: ", err)
			if ctx.WantJSON() {
				response.Error(ctx, err.Error())
			} else {
				h.showForm(ctx, aAlert().Warning(err.Error()), param.Prefix, param.Param, true)
			}
			return
		}
	}

	pmv       := param.MultiForm.Value
	formPanel := param.Panel.GetForm()

	for _, ff := range formPanel.FieldList {
		if ff.FormType == form.File {
			if len(pmf[ff.Field]) == 0 {
				df := pmv[ff.Field + "__delete_flag"]
				if len(df) > 0 && df[0] != "1" {
					pmv[ff.Field] = []string{ "" }
				}
			}
			cf := pmv[ff.Field + "__change_flag"]
			if len(cf) > 0 && cf[0] != "1" {
				delete(pmv, ff.Field)
			}
		}
	}
	/*for i := 0; i < len(formPanel.FieldList); i++ {
		if formPanel.FieldList[i].FormType == form.File &&
			len(param.MultiForm.File[formPanel.FieldList[i].Field]) == 0 &&
			len(param.MultiForm.Value[formPanel.FieldList[i].Field+"__delete_flag"]) > 0 &&
			param.MultiForm.Value[formPanel.FieldList[i].Field+"__delete_flag"][0] != "1" {
			param.MultiForm.Value[formPanel.FieldList[i].Field] = []string{""}
		}
		if formPanel.FieldList[i].FormType == form.File &&
			len(param.MultiForm.Value[formPanel.FieldList[i].Field+"__change_flag"]) > 0 &&
			param.MultiForm.Value[formPanel.FieldList[i].Field+"__change_flag"][0] != "1" {
			delete(param.MultiForm.Value, formPanel.FieldList[i].Field)
		}
	}*/

	err := param.Panel.UpdateData(param.Value())
	if err != nil {
		logger.Error("update data error: ", err)
		if ctx.WantJSON() {
			response.Error(ctx, err.Error(), map[string]interface{}{
				"token": h.authSrv().AddToken(),
			})
		} else {
			h.showForm(ctx, aAlert().Warning(err.Error()), param.Prefix, param.Param, true)
		}
		return
	}

	if formPanel.Responder != nil {
		formPanel.Responder(ctx)
		return
	}

	if ctx.WantJSON() && !param.IsIframe {
		response.OkWithData(ctx, map[string]interface{}{
			"url"  : param.PreviousPath,
			"token": h.authSrv().AddToken(),
		})
		return
	}

	if !param.FromList {
		if utils.IsNewUrl(param.PreviousPath, param.Prefix) {
			h.showNewForm(ctx, param.Alert, param.Prefix, param.Param.DeleteEditPk().GetRouteParamStr(), true)
			return
		}
		if utils.IsEditUrl(param.PreviousPath, param.Prefix) {
			h.showForm(ctx, param.Alert, param.Prefix, param.Param, true, false)
			return
		}
		ctx.HTML(http.StatusOK, fmt.Sprintf(`<script>location.href="%s"</script>`, param.PreviousPath))
		ctx.AddHeader(constant.PjaxUrlHeader, param.PreviousPath)
		return
	}

	if param.IsIframe {
		ctx.HTML(http.StatusOK, fmt.Sprintf(`<script>
		swal('%s', '', 'success');
		setTimeout(function(){
			$("#%s", window.parent.document).hide();
			$('.modal-backdrop.fade.in', window.parent.document).hide();
		}, 1000)
</script>`, language.Get("success"), param.IframeID))
		return
	}

	buf := h.showTable(ctx, param.Prefix, param.Param.DeletePK().DeleteEditPk(), nil)

	ctx.HTML(http.StatusOK, buf.String())
	ctx.AddHeader(constant.PjaxUrlHeader, param.PreviousPath)
}
