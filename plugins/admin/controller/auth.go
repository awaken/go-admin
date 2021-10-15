package controller

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/captcha"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/response"
	"github.com/GoAdminGroup/go-admin/template"
)

// Auth check the input password and username for authentication.
func (h *Handler) Auth(ctx *context.Context) {
	var (
		user   models.UserModel
		ok     bool
		errMsg = "fail"
		s      = h.services.Get(auth.ServiceKey)
	)

	if capDriver, ok := h.captchaConfig["driver"]; ok {
		if capt, ok := captcha.Get(capDriver); ok {
			if !capt.Validate(ctx.FormValue("token")) {
				response.BadRequest(ctx, "wrong captcha")
				return
			}
		}
	}

	if s == nil {
		username := ctx.FormValue("username")
		password := ctx.FormValue("password")
		if username == "" || password == "" {
			response.BadRequest(ctx, "wrong username or password")
			return
		}
		user, ok = auth.Check(username, password, h.conn)
	} else {
		user, ok, errMsg = auth.GetService(s).P(ctx)
	}

	if !ok {
		response.BadRequest(ctx, errMsg)
		return
	}
	if user.IsDisabled() {
		response.BadRequest(ctx, "disabled account")
		return
	}

	err := auth.SetCookie(ctx, user, h.conn)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}

	if ref := ctx.Referer(); ref != "" {
		if u, err := url.Parse(ref); err == nil {
			if r := u.Query().Get("ref"); r != "" {
				rr, _ := url.QueryUnescape(r)
				response.OkWithData(ctx, map[string]interface{}{ "url": rr })
				return
			}
		}
	}

	response.OkWithData(ctx, map[string]interface{}{ "url": h.config.GetIndexURL() })
}

// Logout delete the cookie.
func (h *Handler) Logout(ctx *context.Context) {
	err := auth.DelCookie(ctx, db.GetConnection(h.services))
	if err != nil {
		logger.Error("user logout error:", err)
	}
	ctx.AddHeader("Location", h.config.Url(config.GetLoginUrl()))
	ctx.SetStatusCode(302)
}

// ShowLogin show the login page.
func (h *Handler) ShowLogin(ctx *context.Context) {
	ses, _ := auth.LoadSession(ctx, db.GetConnection(h.services))
	if ses != nil {
		ctx.AddHeader("Location", config.PrefixFixSlash())
		ctx.SetStatusCode(302)
		return
	}

	tmpl, name := template.GetComp("login").GetTemplate()
	var sb strings.Builder

	err := tmpl.ExecuteTemplate(&sb, name, struct {
		UrlPrefix string
		Title     string
		Logo      template.HTML
		CdnUrl    string
	}{
		UrlPrefix: h.config.AssertPrefix(),
		Title:     h.config.LoginTitle,
		Logo:      h.config.LoginLogo,
		CdnUrl:    h.config.AssetUrl,
	})

	if err == nil {
		ctx.HTML(http.StatusOK, sb.String())
	} else {
		logger.Error(err)
		ctx.HTML(http.StatusOK, "error: cannot parse login template")
	}
}
