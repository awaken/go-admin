package controller

import (
	"encoding/json"
	"github.com/GoAdminGroup/go-admin/modules/db"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/response"
)

func (h *Handler) Operation(ctx *context.Context) {
	id := ctx.Query("__goadmin_op_id")
	if !h.OperationHandler(config.Url("/operation/"+id), ctx) {
		errMsg := "not found"
		if ctx.IsDataRequest() {
			response.BadRequest(ctx, errMsg)
		} else {
			response.Alert(ctx, errMsg, errMsg, errMsg, h.conn, h.navButtons)
		}
		return
	}
}

// RecordOperationLog record all operation logs, store into database.
func (h *Handler) RecordOperationLog(ctx *context.Context) {
	RecordOperationLog(ctx, h.conn)
}

func RecordOperationLog(ctx *context.Context, conn db.Connection) {
	if user, ok := ctx.User().(models.UserModel); ok {
		var input []byte
		form := ctx.Request.MultipartForm
		if form != nil {
			input, _ = json.Marshal(form.Value)
		}
		_ = user
		_ = input
		//models.OperationLog().SetConn(conn).New(user.Id, ctx.Path(), ctx.Method(), ctx.LocalIP(), string(input))
	}
}
